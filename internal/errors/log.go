package specerr

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging capabilities
type Logger struct {
	logger    *log.Logger
	level     LogLevel
	prefix    string
	timestamp bool
}

// DefaultLogger is the default logger used throughout the package
var DefaultLogger = NewLogger(os.Stderr, LevelInfo)

// NewLogger creates a new logger with the specified output and level
func NewLogger(output *os.File, level LogLevel) *Logger {
	return &Logger{
		logger:    log.New(output, "", 0),
		level:     level,
		timestamp: true,
	}
}

// WithPrefix sets a prefix for all log messages
func (l *Logger) WithPrefix(prefix string) *Logger {
	newLogger := *l
	newLogger.prefix = prefix
	return &newLogger
}

// WithTimestamp enables or disables timestamp in log messages
func (l *Logger) WithTimestamp(enabled bool) *Logger {
	newLogger := *l
	newLogger.timestamp = enabled
	return &newLogger
}

// log formats and outputs a log message
func (l *Logger) log(level LogLevel, format string, args ...any) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := ""
	if l.timestamp {
		timestamp = time.Now().Format("2006-01-02 15:04:05") + " "
	}

	prefix := l.prefix
	if prefix != "" {
		prefix = "[" + prefix + "] "
	}

	l.logger.Print(timestamp + prefix + "[" + level.String() + "] " + msg)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...any) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...any) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...any) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.log(LevelError, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...any) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// LogError logs an error with full context
func (l *Logger) LogError(err error, context map[string]any) {
	if err == nil {
		return
	}

	l.Error("Error: %v", err)

	// Log context if available
	if specErr, ok := err.(*SpecError); ok {
		l.Error("  Code: %s", specErr.Code)
		l.Error("  Category: %s", specErr.Category)
		if specErr.Underlying != nil {
			l.Error("  Underlying: %v", specErr.Underlying)
		}
	}

	// Log additional context
	if len(context) > 0 {
		l.Error("  Context:")
		for k, v := range context {
			l.Error("    %s: %v", k, v)
		}
	}
}

// LogRetry logs retry attempts
func (l *Logger) LogRetry(attempt int, maxAttempts int, err error) {
	l.Warn("Retry attempt %d/%d failed: %v", attempt, maxAttempts, err)
}

// LogRetrySuccess logs successful retry
func (l *Logger) LogRetrySuccess(attempt int, duration time.Duration) {
	l.Info("Retry succeeded on attempt %d after %v", attempt, duration)
}

// Package-level logging functions

// Debug logs a debug message using the default logger
func Debug(format string, args ...any) {
	DefaultLogger.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...any) {
	DefaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...any) {
	DefaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...any) {
	DefaultLogger.Error(format, args...)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(format string, args ...any) {
	DefaultLogger.Fatal(format, args...)
}

// LogError logs an error with context using the default logger
func LogError(err error, context map[string]any) {
	DefaultLogger.LogError(err, context)
}

// SetLevel sets the log level for the default logger
func SetLevel(level LogLevel) {
	DefaultLogger.level = level
}
