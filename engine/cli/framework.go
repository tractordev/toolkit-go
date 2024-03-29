package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

// Initializer is a hook to allow units to customize the root Command.
type Initializer interface {
	InitializeCLI(root *Command)
}

type Preprocessor interface {
	PreprocessCLI(args []string) []string
}

// Runner is a unit that takes over the program entrypoint.
type Runner interface {
	Run(ctx context.Context) error
}

// Framework manages a root command, allowing Initializers
// to modify it, which by default runs a DefaultRunner.
type Framework struct {
	DefaultRunner Runner
	Initializers  []Initializer
	Preprocessors []Preprocessor
	Root          *Command
}

// Initialize sets up a Root command that simply runs the
// DefaultRunner, and also runs any Initializers.
func (f *Framework) Initialize() {
	f.Root = &Command{}
	for _, i := range f.Initializers {
		i.InitializeCLI(f.Root)
	}
}

// Run executes the root command with os.Args and STDIO.
func (f *Framework) Run(ctx context.Context) error {
	args := os.Args[1:]
	for _, p := range f.Preprocessors {
		args = p.PreprocessCLI(args)
	}
	return Execute(ContextWithIO(ctx, os.Stdin, os.Stdout, os.Stderr), f.Root, args)
}

// Execute takes a root Command plus arguments, finds the Command to run,
// parses flags, checks for expected arguments, and runs the Command.
// It also adds a version flag if the root Command has Version set.
func Execute(ctx context.Context, root *Command, args []string) error {
	var (
		stdout io.Writer = os.Stdout
		stderr io.Writer = os.Stderr
		ioctx  *Context
	)
	if c, ok := ctx.(*Context); ok {
		stdout = c
		stderr = c
		ioctx = c
	} else {
		ioctx = ContextWithIO(ctx, os.Stdin, stdout, stderr)
	}

	var showVersion bool
	if root.Version != "" {
		root.Flags().BoolVar(&showVersion, "v", false, "show version")
	}

	cmd, n := root.Find(args)
	f := cmd.Flags()
	if f != nil {
		if err := f.Parse(args[n:]); err != nil {
			if err == flag.ErrHelp {
				return (&CommandHelp{cmd}).WriteHelp(stderr)
			}
			return err
		}
	}

	if showVersion {
		fmt.Fprintln(stdout, root.Version)
		return nil
	}

	if cmd.Args != nil {
		if err := cmd.Args(cmd, f.Args()); err != nil {
			return err
		}
	}

	if cmd.Run == nil {
		(&CommandHelp{cmd}).WriteHelp(stderr)
		return nil
	}

	cmd.Run(ioctx, f.Args())
	return nil
}

// Export wraps a function as a command.
func Export(fn interface{}, use string) *Command {
	rv := reflect.ValueOf(fn)
	t := rv.Type()
	if t.Kind() != reflect.Func {
		panic("can only export funcs")
	}
	return &Command{
		Usage: use,
		Args:  ExactArgs(t.NumIn()),
		Run: func(ctx *Context, args []string) {
			var in []reflect.Value
			for n := 0; n < t.NumIn(); n++ {
				switch t.In(n).Kind() {
				case reflect.String:
					in = append(in, reflect.ValueOf(args[n]))
				case reflect.Int:
					arg, err := strconv.Atoi(args[n])
					if err != nil {
						panic(err)
					}
					in = append(in, reflect.ValueOf(arg))
				default:
					panic("argument kind not supported: " + t.In(n).Kind().String())
				}
			}
			rv.Call(in)
		},
	}
}
