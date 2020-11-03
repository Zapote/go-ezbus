package logger

import (
	"log"
	"os"
)

var level LogLevel = InfoLevel
var (
	debugLogger *log.Logger = log.New(os.Stdout, "DEBUG:", log.Ldate|log.Ltime)
	infoLogger  *log.Logger = log.New(os.Stdout, "INFO:", log.Ldate|log.Ltime)
	warnLogger  *log.Logger = log.New(os.Stdout, "WARN:", log.Ldate|log.Ltime)
	errorLogger *log.Logger = log.New(os.Stderr, "ERR:", log.Ldate|log.Ltime)
)

//LogLevel for logger
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

func write(logger *log.Logger, l LogLevel, format string, v ...interface{}) {
	if l < level {
		return
	}
	logger.Printf(format, v...)
}

//Debug log
func Debug(msg string) {
	Debugf(msg)
}

//Debugf log with format
func Debugf(format string, v ...interface{}) {
	write(debugLogger, DebugLevel, format, v...)
}

//Info log
func Info(msg string) {
	Infof(msg)
}

//Infof log with format
func Infof(format string, v ...interface{}) {
	write(infoLogger, InfoLevel, format, v...)
}

//Warn log
func Warn(msg string) {
	Warnf(msg)
}

//Warnf log with format
func Warnf(format string, v ...interface{}) {
	write(warnLogger, WarnLevel, format, v...)
}

//Error log
func Error(msg string) {
	Errorf(msg)
}

//Errorf log with format
func Errorf(format string, v ...interface{}) {
	write(errorLogger, ErrorLevel, format, v...)
}

//SetLevel of logging: DebugLevel, InfoLevel, WarnLevel, ErrorLevel. Default InfoLevel
func SetLevel(l LogLevel) {
	level = l
}
