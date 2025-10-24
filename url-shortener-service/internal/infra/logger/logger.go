package logger

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

// New creates a new zerolog instance with standard fields
func New(service, level, appEnv string) Logger {
	var out io.Writer
	if appEnv == "prod" || appEnv == "qa" {
		out = os.Stdout
	} else {
		out = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	core := zerolog.New(out).
		Level(lvl).
		With().
		Timestamp().
		Str("service", service).
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
//	defer logger.TimeTrack(logger.FromContext(ctx), time.Now(), "db.insert")
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
