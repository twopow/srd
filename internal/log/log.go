package log

import (
	"context"
	"io"
	"log/slog"
	"os"
)

var (
	// Global logger instance
	globalLogger *slog.Logger
	// Default handler for console output
	defaultHandler *slog.JSONHandler
)

// Level represents log levels
type Level = slog.Level

const (
	DebugLevel = slog.LevelDebug
	InfoLevel  = slog.LevelInfo
	WarnLevel  = slog.LevelWarn
	ErrorLevel = slog.LevelError
	FatalLevel = slog.LevelError + 1
)

// Event represents a log event
type Event struct {
	logger *slog.Logger
	attrs  []any
	level  slog.Level
}

// Logger represents a logger instance
type Logger struct {
	*slog.Logger
}

// init initializes the default logger
func init() {
	defaultHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	globalLogger = slog.New(defaultHandler)
}

// SetGlobalLevel sets the global log level
func SetGlobalLevel(level slog.Level) {
	defaultHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	globalLogger = slog.New(defaultHandler)
}

// SetPrettyOutput enables pretty console output
func SetPrettyOutput(w io.Writer) {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// if a.Key == slog.TimeKey {
			// 	return slog.Attr{
			// 		Key:   slog.TimeKey,
			// 		Value: slog.StringValue(time.Unix(a.Value.Int64(), 0).Format("2006-01-02T15:04:05.000Z07:00")),
			// 	}
			// }
			return a
		},
	})
	globalLogger = slog.New(handler)
}

// With creates a new logger with additional attributes
func With(attrs ...any) *Logger {
	return &Logger{globalLogger.With(attrs...)}
}

// Debug creates a debug level event
func Debug() *Event {
	return &Event{
		logger: globalLogger,
		level:  DebugLevel,
	}
}

// Info creates an info level event
func Info() *Event {
	return &Event{
		logger: globalLogger,
		level:  InfoLevel,
	}
}

// Warn creates a warn level event
func Warn() *Event {
	return &Event{
		logger: globalLogger,
		level:  WarnLevel,
	}
}

// Error creates an error level event
func Error() *Event {
	return &Event{
		logger: globalLogger,
		level:  ErrorLevel,
	}
}

// Fatal creates a fatal level event
func Fatal() *Event {
	return &Event{
		logger: globalLogger,
		level:  FatalLevel,
	}
}

// Str adds a string attribute to the event
func (e *Event) Str(key, value string) *Event {
	e.attrs = append(e.attrs, key, value)
	return e
}

// Int64 adds an int64 attribute to the event
func (e *Event) Int64(key string, value int64) *Event {
	e.attrs = append(e.attrs, key, value)
	return e
}

// Int adds an int attribute to the event
func (e *Event) Int(key string, value int) *Event {
	e.attrs = append(e.attrs, key, value)
	return e
}

// Err adds an error attribute to the event
func (e *Event) Err(err error) *Event {
	e.attrs = append(e.attrs, "error", err.Error())
	return e
}

// Msg logs the event with the given message
func (e *Event) Msg(msg string) {
	if e.level == FatalLevel {
		e.logger.Error(msg, e.attrs...)
		os.Exit(1)
	} else {
		e.logger.Log(context.Background(), e.level, msg, e.attrs...)
	}
}

// Logger methods for compatibility
func (l *Logger) Debug() *Event {
	return &Event{
		logger: l.Logger,
		level:  DebugLevel,
	}
}

func (l *Logger) Info() *Event {
	return &Event{
		logger: l.Logger,
		level:  InfoLevel,
	}
}

func (l *Logger) Warn() *Event {
	return &Event{
		logger: l.Logger,
		level:  WarnLevel,
	}
}

func (l *Logger) Error() *Event {
	return &Event{
		logger: l.Logger,
		level:  ErrorLevel,
	}
}

func (l *Logger) Fatal() *Event {
	return &Event{
		logger: l.Logger,
		level:  FatalLevel,
	}
}
