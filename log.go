package main

import "fmt"

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type Logger struct {
	level LogLevel
}

func CreateLogger(level LogLevel) *Logger {
	return &Logger{level: level}
}

func (l *Logger) Debug(format string, values ...any) {
	if LogLevelDebug >= l.level {
		fmt.Printf(format, values...)
	}
}

func (l *Logger) Info(format string, values ...any) {
	if LogLevelInfo >= l.level {
		fmt.Printf(format, values...)
	}
}

func (l *Logger) Warn(format string, values ...any) {
	if LogLevelError >= l.level {
		fmt.Printf(format, values...)
	}
}

func (l *Logger) Error(format string, values ...any) {
	if LogLevelError >= l.level {
		fmt.Printf(format, values...)
	}
}
