package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger wraps zerolog.Logger for context-safe usage.
type Logger struct {
	zLog *zerolog.Logger
}

// Config stores required fields to init Logger
type Config struct {
	serviceName string
	logLevel    string
	appEnv      string
}

// FromEnv loads config from environment variables
func FromEnv() Config {
	return Config{
		serviceName: os.Getenv("SERVICE_NAME"),
		logLevel:    os.Getenv("LOG_LEVEL"),
		appEnv:      os.Getenv("APP_ENV"),
	}
}

// New creates a new zerolog instance with standard fields
func New(cfg Config) Logger {
	var out io.Writer
	if cfg.appEnv == "prod" || cfg.appEnv == "qa" {
		out = os.Stdout
	} else {
		out = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	lvl, err := zerolog.ParseLevel(cfg.logLevel)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	core := zerolog.New(out).
		Level(lvl).
		With().
		Timestamp().
		Str("service", cfg.serviceName).
		Logger()

	return Logger{zLog: &core}
}

var nop = zerolog.Nop()

// Z returns usable zerolog.Logger (fallback to nop if nil)
func (l Logger) Z() *zerolog.Logger {
	if l.zLog == nil {
		return &nop
	}
	return l.zLog
}

// TimeTrack logs how long a scope took (use with defer at call site)
// Example:
//
//	defer logging.TimeTrack(logging.FromContext(ctx), time.Now(), "db.insert")
func TimeTrack(l *zerolog.Logger, start time.Time, scope string) {
	if l == nil {
		return
	}
	elapsed := time.Since(start)
	l.Debug().
		Str("scope", scope).
		Dur("elapsed_ms", elapsed).
		Msg("timing")
}
