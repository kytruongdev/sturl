package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	once   sync.Once
	global *zerolog.Logger
)

// Init sets global Zerolog logger with service name and RFC3339Nano timestamp.
// Goal: uniform JSON format, add "service" attribute in every log line.
// this func must be called exactly once at startup (e.g., in main()).
func Init() {
	once.Do(func() {
		var out io.Writer
		appEnv := os.Getenv("APP_ENV")
		if appEnv == "prod" || appEnv == "qa" {
			out = os.Stdout
		} else {
			out = zerolog.ConsoleWriter{Out: os.Stdout}
		}

		// timestamps + error stacks
		zerolog.TimeFieldFormat = time.RFC3339Nano
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

		lvl, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
		if err != nil {
			lvl = zerolog.InfoLevel
		}

		// attach service name at the base so all child loggers inherit it
		l := zerolog.New(out).
			Level(lvl).
			With().
			Timestamp().
			Str("service", os.Getenv("SERVICE_NAME")).
			Logger()

		global = &l
	})
}

// Get returns the process-wide base logger.
// It panics if Init(...) hasn't been called.
func Get() *zerolog.Logger {
	if global == nil {
		panic("logger not initialized; call logger.Init(...) first")
	}
	return global
}

// TimeTrack logs how long a scope took (use with defer at the call site).
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
