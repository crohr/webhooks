package webhooks

import (
	"io"
	"log"
)

// Level represents log levels which are sorted based on the underlying number associated to it.
type Level uint8

// Definitions of available levels: Debug < Info < Warning < Error.
const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

type Logger struct {
	lg    *log.Logger
	level Level
}

func NewLogger(w io.Writer, lv Level) *Logger {
	return &Logger{
		lg:    log.New(w, "", log.LstdFlags|log.Lshortfile),
		level: lv,
	}
}

func (l *Logger) SetLevel(lv Level) {
	l.level = lv
}

func (l *Logger) Fatalf(fmt string, msg ...interface{}) {
	l.lg.Fatalf("FATAL> "+fmt, msg...)
}

func (l *Logger) logf(lv Level, fmt string, msg ...interface{}) {
	if lv >= l.level {
		l.lg.Printf(fmt, msg...)
	}
}

func (l *Logger) Debugf(fmt string, msg interface{}) {
	l.logf(DEBUG, "DEBUG> "+fmt, msg)
}

func (l *Logger) Infof(fmt string, msg ...interface{}) {
	l.logf(INFO, "INFO> "+fmt, msg...)
}

func (l *Logger) Warningf(fmt string, msg interface{}) {
	l.logf(WARNING, "WARNING> "+fmt, msg)
}

func (l *Logger) Errorf(fmt string, msg ...interface{}) {
	l.logf(ERROR, "ERROR> "+fmt, msg...)

}

func (l *Logger) Fatal(msg interface{}) {
	l.lg.Fatal("FATAL> ", msg)
}

func (l *Logger) log(lv Level, msg ...interface{}) {
	if lv >= l.level {
		l.lg.Println(msg...)
	}
}

func (l *Logger) Debug(msg interface{}) {
	l.log(DEBUG, "DEBUG> ", msg)
}

func (l *Logger) Info(msg interface{}) {
	l.log(INFO, "INFO> ", msg)
}

func (l *Logger) Warning(msg interface{}) {
	l.log(WARNING, "WARNING> ", msg)
}

func (l *Logger) Error(msg interface{}) {
	l.log(ERROR, "ERROR> ", msg)
}
