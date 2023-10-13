// log is a minimal wrapper around standard logging that encourages three log levels:
// Debug, Info, and Fatal. Fatal and Info work like log.Fatal and log.Println. If an
// error is given as the last argument to Debug or Info it will upgrade to an Error
// level. Debug logs are only shown when SetDebug is true and the format changes to
// include more specificity in the level and source of the log.
package log

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func init() {
	DefaultLogger = &Logger{
		Logger: log.New(log.Writer(), "", log.LstdFlags),
	}
}

var DefaultLogger *Logger

type Logger struct {
	*log.Logger
	debug bool
	ident string
}

func (l *Logger) SetIdent(id string) {
	l.ident = id
}

func (l *Logger) SetDebug(enabled bool) {
	l.debug = enabled
	if l.debug {
		l.SetFlags(log.Lmicroseconds)
	}
}

func (l *Logger) Info(v ...any) {
	l.Log(2, "INFO", v...)
}

func (l *Logger) Debug(v ...any) {
	l.Log(2, "DEBUG", v...)
}

func (l *Logger) Fatal(v ...any) {
	l.Log(2, "FATAL", v...)
}

func (l *Logger) Log(depth int, level string, v ...any) {
	if !l.debug && level == "DEBUG" {
		return
	}
	if _, ok := v[len(v)-1].(error); ok && level != "FATAL" {
		level = "ERROR"
	}
	_, file, line, _ := runtime.Caller(depth)
	var source, ident string
	if l.ident != "" {
		ident = fmt.Sprintf("[%s]", l.ident)
	}
	if l.debug {
		source = fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(file)), filepath.Base(file), line)
		l.Printf("%s%s: %s: %s", level, ident, source, strings.TrimSpace(fmt.Sprintln(v...)))
	} else {
		source = filepath.Base(filepath.Dir(file))
		l.Printf("%s%s: %s", ident, source, strings.TrimSpace(fmt.Sprintln(v...)))
	}
	if level == "FATAL" {
		os.Exit(1)
	}
}

// Write makes Logger a writer to add as standard logger's output
func (l *Logger) Write(p []byte) (n int, err error) {
	scanner := bufio.NewScanner(bytes.NewBuffer(p))
	for scanner.Scan() {
		text := scanner.Text()
		l.Log(4, "INFO", text)
		n += len(text) + 1
	}
	return
}
