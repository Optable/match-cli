package cli

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// LevelFromVerbosity takes a command-line `-v` stackable flag count, e.g.
// `-vv`, `-vvv` and transforms it into a sensible loglevel.
// The mapping is:
//   ``:     Warn
//   `-v`:   Info
//   `-vv`:  Debug
//   `-vvv`: Trace
func LevelFromVerbosity(v int) zerolog.Level {
	switch v {
	case 0:
		return zerolog.WarnLevel
	case 1:
		return zerolog.InfoLevel
	case 2:
		return zerolog.DebugLevel
	default:
		return zerolog.TraceLevel
	}
}

func NewLogger(cliName string, verbosity int) *zerolog.Logger {
	logger := zerolog.
		New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(LevelFromVerbosity(verbosity)).With().Timestamp().
		Str("cli", cliName).
		Logger()
	return &logger
}

func withInfoLogger(ctx context.Context) context.Context {
	logger := *zerolog.Ctx(ctx)
	if logger.GetLevel() > zerolog.InfoLevel {
		logger = logger.Level(zerolog.InfoLevel)
	}
	return logger.WithContext(ctx)
}

func info(ctx context.Context) *zerolog.Event {
	return zerolog.Ctx(ctx).Info()
}

func debug(ctx context.Context) *zerolog.Event {
	return zerolog.Ctx(ctx).Debug()
}
