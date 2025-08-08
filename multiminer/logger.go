package multiminer

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents logging severity levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger interface for structured logging
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	WithFields(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// F creates a logging field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// SimpleLogger is a basic implementation of the Logger interface
type SimpleLogger struct {
	level  LogLevel
	logger *log.Logger
	fields []Field
}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger(level LogLevel) *SimpleLogger {
	return &SimpleLogger{
		level:  level,
		logger: log.New(os.Stdout, "[MultiMiner] ", log.LstdFlags|log.Lshortfile),
	}
}

func (l *SimpleLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, fields...)
	}
}

func (l *SimpleLogger) Info(ctx context.Context, msg string, fields ...Field) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, fields...)
	}
}

func (l *SimpleLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, fields...)
	}
}

func (l *SimpleLogger) Error(ctx context.Context, msg string, fields ...Field) {
	if l.level <= LogLevelError {
		l.log("ERROR", msg, fields...)
	}
}

func (l *SimpleLogger) WithFields(fields ...Field) Logger {
	newFields := make([]Field, 0, len(l.fields)+len(fields))
	newFields = append(newFields, l.fields...)
	newFields = append(newFields, fields...)

	return &SimpleLogger{
		level:  l.level,
		logger: l.logger,
		fields: newFields,
	}
}

func (l *SimpleLogger) log(level, msg string, fields ...Field) {
	allFields := make([]Field, 0, len(l.fields)+len(fields))
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	logMsg := level + " " + msg
	if len(allFields) > 0 {
		logMsg += " |"
		for _, field := range allFields {
			logMsg += " " + field.Key + "=" + formatValue(field.Value)
		}
	}

	l.logger.Output(3, logMsg)
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	case int, int64:
		return fmt.Sprintf("%v", val)
	case float64:
		return fmt.Sprintf("%.2f", val)
	case time.Duration:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Global logger instance (optional for library consumers)
var globalLogger Logger

func init() {
	// Default to a no-op logger to avoid conflicts with host applications
	globalLogger = &NoOpLogger{}
}

// SetLogger sets the global logger (optional - library consumers can manage their own logging)
func SetLogger(logger Logger) {
	globalLogger = logger
}

// GetLogger returns the global logger
func GetLogger() Logger {
	return globalLogger
}

// Convenience functions using the global logger (only logs if a logger is set)
func LogDebug(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Debug(ctx, msg, fields...)
	}
}

func LogInfo(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Info(ctx, msg, fields...)
	}
}

func LogWarn(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Warn(ctx, msg, fields...)
	}
}

func LogError(ctx context.Context, msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Error(ctx, msg, fields...)
	}
}

// NoOpLogger is a logger that does nothing (safe default for library usage)
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(ctx context.Context, msg string, fields ...Field) {}
func (l *NoOpLogger) Info(ctx context.Context, msg string, fields ...Field)  {}
func (l *NoOpLogger) Warn(ctx context.Context, msg string, fields ...Field)  {}
func (l *NoOpLogger) Error(ctx context.Context, msg string, fields ...Field) {}
func (l *NoOpLogger) WithFields(fields ...Field) Logger                      { return l }
