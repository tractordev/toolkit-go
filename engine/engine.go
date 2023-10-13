package engine

import (
	"context"
	stdlog "log"
	"os"
	"path/filepath"
	"reflect"

	_ "embed"

	"tractor.dev/toolkit/engine/cli"
	"tractor.dev/toolkit/engine/daemon"
	"tractor.dev/toolkit/engine/log"
)

var (
	Identifier string
)

// Initializer provides an initialization hook after assembly.
type Initializer interface {
	Initialize()
}

// PostInitializer provides a hook after initialization.
type PostInitializer interface {
	PostInitialize()
}

// Service is a long-running process managed by daemon.Framework.
type Service interface {
	Serve(ctx context.Context)
}

// Runner is a unit that can take over the program entrypoint.
type Runner interface {
	Run(ctx context.Context) error
}

type Terminator interface {
	Terminate()
}

type Depender interface {
	Assembly() []Unit
}

var (
	// Main is a global reference to the top-level main unit.
	Main Unit

	defaultAssembly *Assembly
)

func typeExists(units []Unit, unit Unit) bool {
	for _, u := range units {
		a := reflect.ValueOf(unitFrom(unit))
		b := reflect.ValueOf(unitFrom(u))
		if a.Type() == b.Type() {
			return true
		}
	}
	return false
}

func Dependencies(units ...Unit) []Unit {
	var deps []Unit
	for _, unit := range units {
		if d, ok := unit.(Depender); ok {
			for _, dep := range d.Assembly() {
				if !typeExists(deps, dep) && !typeExists(units, dep) {
					deps = append(deps, dep)
				}
			}
		}
	}
	return append(units, deps...)
}

func Assemble(units ...Unit) (asm *Assembly, err error) {
	if asm, err = New(Dependencies(units...)...); err != nil {
		return
	}

	// dependency injection
	if err = asm.SelfAssemble(); err != nil {
		return
	}

	// initialize units after DI, in reverse order (main last)
	for i := len(asm.Units()) - 1; i >= 0; i-- {
		u := asm.Units()[i]
		i, ok := u.(Initializer)
		if ok {
			i.Initialize()
		}
	}

	// postinitialize units, for units depending on main initialization
	for _, u := range asm.Units() {
		pi, ok := u.(PostInitializer)
		if ok {
			pi.PostInitialize()
		}
	}

	return
}

// Init only needs to be explicitly called if you
// need to run code before calling Run
func Init() {
	if Identifier == "" {
		path, err := os.Executable()
		if err != nil {
			panic(err)
		}
		Identifier = filepath.Base(path)
	}
	stdlog.SetOutput(log.DefaultLogger)
	stdlog.SetFlags(0)
}

// Run assembles units and starts the program.
func Run(units ...Unit) {
	Init()

	asm, err := New(units...)
	if err != nil {
		Fatal(err)
	}

	// add assembly
	if err := asm.Add(asm); err != nil {
		panic(err)
	}

	// add daemon framework
	d := &daemon.Framework{}
	if err := asm.Add(d); err != nil {
		panic(err)
	}

	// add cli framework
	c := &cli.Framework{}
	if err := asm.Add(c); err != nil {
		panic(err)
	}

	// add logger
	if err := asm.Add(log.DefaultLogger); err != nil {
		panic(err)
	}

	// re-assemble
	asm, err = Assemble(asm.Units()...)
	if err != nil {
		panic(err)
	}
	defaultAssembly = asm

	// make main unit global accessible
	Main = asm.Main()

	// find a runner; should always be the cli.Framework
	for i := len(asm.Units()) - 1; i >= 0; i-- {
		u := asm.Units()[i]
		r, ok := u.(Runner)
		if ok {
			if err := r.Run(context.Background()); err != nil {
				Fatal(err)
				os.Exit(1)
			}
			return
		}
	}

	panic("nothing to run")
}

func Terminate() {
	if defaultAssembly == nil {
		panic("no active default assembly")
	}
	for i := len(defaultAssembly.Units()) - 1; i >= 0; i-- {
		u := defaultAssembly.Units()[i]
		if t, ok := u.(Terminator); ok {
			t.Terminate()
			return
		}
	}
}

func Logger() *log.Logger {
	return log.DefaultLogger
}

func Info(v ...any) {
	Logger().Log(2, "INFO", v...)
}

func Debug(v ...any) {
	Logger().Log(2, "DEBUG", v...)
}

func Fatal(v ...any) {
	Logger().Log(2, "FATAL", v...)
}

func FatalErr(err error) {
	if err != nil {
		Logger().Log(2, "FATAL", err)
	}
}
