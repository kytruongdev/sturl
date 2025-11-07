package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Field represents a typed key-value pair used to enrich structured logs
type Field func(zerolog.Context) zerolog.Context

// String returns a string Field for structured logging
func String(k, v string) Field { return func(c zerolog.Context) zerolog.Context { return c.Str(k, v) } }

// Int returns an int Field for structured logging.
func Int(k string, v int) Field {
	return func(c zerolog.Context) zerolog.Context { return c.Int(k, v) }
}

// Int64 returns an int64 Field for structured logging
func Int64(k string, v int64) Field {
	return func(c zerolog.Context) zerolog.Context { return c.Int64(k, v) }
}

// Dur returns a duration Field for structured logging
func Dur(k string, v time.Duration) Field {
	return func(c zerolog.Context) zerolog.Context { return c.Dur(k, v) }
}

// Config holds logging options such as level, pretty mode, service name, and environment
type Config struct {
	ServiceName string
	LogLevel    string
	AppEnv      string
}

// FromEnv loads logging configuration from environment variables
func FromEnv() Config {
	return Config{
		ServiceName: os.Getenv("SERVICE_NAME"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		AppEnv:      os.Getenv("APP_ENV"),
	}
}

// Logger wraps a zerolog.LoggerWithSpan for structured, contextual logging
type Logger struct{ z *zerolog.Logger }

// New creates a new Logger from the provided Config
func New(cfg Config) Logger {
	var out io.Writer
	if cfg.AppEnv == "prod" || cfg.AppEnv == "qa" {
		out = os.Stdout
	} else {
		out = zerolog.ConsoleWriter{Out: os.Stdout}
	}
	lvl, _ := zerolog.ParseLevel(cfg.LogLevel)
	core := zerolog.New(out).Level(lvl).With().Timestamp().Str("service", cfg.ServiceName).Logger()
	return Logger{z: &core}
}

// With creates a new Logger derived from the current one and enriched
// with the provided structured fields. Each Field function appends
// a typed key-value pair to the log context.
func (l Logger) With(fields ...Field) Logger {
	b := l.z.With()
	for _, f := range fields {
		b = f(b)
	}
	child := b.Logger()
	return Logger{z: &child}
}

// Info returns a new log event at the Info level
func (l Logger) Info() *zerolog.Event { return l.z.Info() }

// Error returns a new log event at the Error level
func (l Logger) Error() *zerolog.Event { return l.z.Error() }

// Debug returns a new log event at the Debug level
func (l Logger) Debug() *zerolog.Event { return l.z.Debug() }

// Warn returns a new log event at the Warn level
func (l Logger) Warn() *zerolog.Event { return l.z.Warn() }

// TimeTrack logs the duration elapsed since the provided start time
func (l Logger) TimeTrack(start time.Time, label string, fields ...Field) {
	elapsed := time.Since(start)
	l.With(fields...).z.Info().Dur("elapsed", elapsed).Msg(label)
}
