package logging

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/iostreams"
)

type Logger struct {
	ColorScheme *iostreams.ColorScheme
	IO          *iostreams.IOStreams
	quiet       bool
	verbose     bool
}

func NewLogger(io *iostreams.IOStreams, isQuiet, isVerbose bool) *Logger {
	return &Logger{
		ColorScheme: io.ColorScheme(),
		IO:          io,
		quiet:       isQuiet,
		verbose:     isVerbose,
	}
}

// NewDefaultLogger returns a Logger that with the default logging settings
func NewDefaultLogger(io *iostreams.IOStreams) *Logger {
	isQuiet := false
	isVerbose := false

	return NewLogger(io, isQuiet, isVerbose)
}

func NewTestLogger() *Logger {
	testIO, _, _, _ := iostreams.Test()
	return NewDefaultLogger(testIO)
}

// Printf writes the formatted arguments to the stderr writer.
func (l *Logger) Printf(f string, v ...interface{}) (int, error) {
	if l.quiet || !l.IO.IsStdoutTTY() {
		return 0, nil
	}
	return fmt.Fprintf(l.IO.ErrOut, f, v...)
}

// Println writes the arguments to the stderr writer with a newline at the end.
func (l *Logger) Println(v ...interface{}) (int, error) {
	if l.quiet || !l.IO.IsStdoutTTY() {
		return 0, nil
	}
	return fmt.Fprintln(l.IO.ErrOut, v...)
}

func (l *Logger) VerbosePrint(msg string) (int, error) {
	if !l.verbose || !l.IO.IsStdoutTTY() {
		return 0, nil
	}

	return fmt.Fprintln(l.IO.ErrOut, msg)
}

func (l *Logger) VerbosePrintf(f string, v ...interface{}) (int, error) {
	if !l.verbose || !l.IO.IsStdoutTTY() {
		return 0, nil
	}

	return fmt.Fprintf(l.IO.ErrOut, f, v...)
}
