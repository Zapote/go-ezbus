package logger

import (
	"context"
	"fmt"
	"log/slog"
)

var level LogLevel = InfoLevel

// LogLevel for logger
type LogLevel int

func (l LogLevel) String() string {
	return [...]string{"Debug", "Info", "Warn", "Error"}[l]
}

const (
	//DebugLevel level
	DebugLevel LogLevel = iota
	//InfoLevel level
	InfoLevel
	//WarnLevel level
	WarnLevel
	//ErrorLevel level
	ErrorLevel
)

func (l LogLevel) slogLevel() slog.Level {
	switch l {
	case DebugLevel:
		return slog.LevelDebug
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func write(l LogLevel, msg string) {
	if l < level {
		return
	}

	slog.Default().Log(context.Background(), l.slogLevel(), msg)
}

func writef(l LogLevel, format string, v ...interface{}) {
	if l < level {
		return
	}

	write(l, fmt.Sprintf(format, v...))
}

// Debug log
func Debug(msg string) {
	write(DebugLevel, msg)
}

// Debugf log with format
func Debugf(format string, v ...interface{}) {
	writef(DebugLevel, format, v...)
}

// Info log
func Info(msg string) {
	write(InfoLevel, msg)
}

// Infof log with format
func Infof(format string, v ...interface{}) {
	writef(InfoLevel, format, v...)
}

// Warn log
func Warn(msg string) {
	write(WarnLevel, msg)
}

// Warnf log with format
func Warnf(format string, v ...interface{}) {
	writef(WarnLevel, format, v...)
}

// Error log
func Error(msg string) {
	write(ErrorLevel, msg)
}

// Errorf log with format
func Errorf(format string, v ...interface{}) {
	writef(ErrorLevel, format, v...)
}

// SetLevel of logging: DebugLevel, InfoLevel, WarnLevel, ErrorLevel. Default InfoLevel
func SetLevel(l LogLevel) {
	level = l
}
