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
	logContext(context.Background(), l, msg)
}

// logContext writes msg with attrs through slog with ctx, so that a handler
// that reads the context, such as one adding trace_id, can act on it.
func logContext(ctx context.Context, l LogLevel, msg string, attrs ...any) {
	if l < level {
		return
	}

	slog.Default().Log(ctx, l.slogLevel(), msg, attrs...)
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

// DebugContext logs msg with ctx and slog-style key-value attrs.
func DebugContext(ctx context.Context, msg string, attrs ...any) {
	logContext(ctx, DebugLevel, msg, attrs...)
}

// InfoContext logs msg with ctx and slog-style key-value attrs.
func InfoContext(ctx context.Context, msg string, attrs ...any) {
	logContext(ctx, InfoLevel, msg, attrs...)
}

// WarnContext logs msg with ctx and slog-style key-value attrs.
func WarnContext(ctx context.Context, msg string, attrs ...any) {
	logContext(ctx, WarnLevel, msg, attrs...)
}

// ErrorContext logs msg with ctx and slog-style key-value attrs.
func ErrorContext(ctx context.Context, msg string, attrs ...any) {
	logContext(ctx, ErrorLevel, msg, attrs...)
}

// SetLevel of logging: DebugLevel, InfoLevel, WarnLevel, ErrorLevel. Default InfoLevel
func SetLevel(l LogLevel) {
	level = l
}
