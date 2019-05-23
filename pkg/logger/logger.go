package logger

import (
	"log"
	"os"
)

// Logger ... Logger module.
type Logger struct {
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
}

// NewLogger ... Initialize new logger.
func NewLogger(f *os.File) *Logger {
	return &Logger{
		log.New(f, "[Info]    ", log.Ldate|log.Ltime),
		log.New(f, "[Warning] ", log.Ldate|log.Ltime),
		log.New(f, "[Error]   ", log.Ldate|log.Ltime),
	}
}

// Infof ... Just like Printf().
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Info.Printf(format, v...)
}

// Errorf ... Just like Printf().
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Error.Printf(format, v...)
}

// Warnf ... Just like Printf().
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Warn.Printf(format, v...)
}
