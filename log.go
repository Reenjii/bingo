package bingo

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

const (
	ERROR = 1 << iota
	WARN
	INFO
	TRACE
)

var Loggers struct {
	Error *log.Logger
	Warn  *log.Logger
	Info  *log.Logger
	Trace *log.Logger
}

func init() {
	// Loggers log to stdout by default
	setupLoggers(os.Stdout, log.Ldate|log.Lmicroseconds)
	setVerbosity(ERROR | WARN | INFO | TRACE)
}

// setupLoggers setup all loggers with given output destination and flags.
func setupLoggers(w io.Writer, flag int) {
	Loggers.Error = log.New(w, "ERROR ", flag)
	Loggers.Warn = log.New(w, "WARN  ", flag)
	Loggers.Info = log.New(w, "INFO  ", flag)
	Loggers.Trace = log.New(w, "TRACE ", flag)
}

// setVerbosity disables all loggers for which the verbosity does not match.
func setVerbosity(verbosity int) {
	if verbosity&ERROR == 0 {
		Loggers.Error.SetOutput(ioutil.Discard)
	}
	if verbosity&WARN == 0 {
		Loggers.Warn.SetOutput(ioutil.Discard)
	}
	if verbosity&INFO == 0 {
		Loggers.Info.SetOutput(ioutil.Discard)
	}
	if verbosity&TRACE == 0 {
		Loggers.Trace.SetOutput(ioutil.Discard)
	}
}

// setAllOutput sets the output destination for all loggers.
func setAllOutput(w io.Writer) {
	Loggers.Trace.SetOutput(w)
	Loggers.Info.SetOutput(w)
	Loggers.Warn.SetOutput(w)
	Loggers.Error.SetOutput(w)
}

// setAllFlags sets the output flags for all loggers.
func setAllFlags(flag int) {
	Loggers.Trace.SetFlags(flag)
	Loggers.Info.SetFlags(flag)
	Loggers.Warn.SetFlags(flag)
	Loggers.Error.SetFlags(flag)
}
