package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(level string) *zerolog.Logger {
	out := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	l := zerolog.New(out).With().Timestamp().Logger()

	switch strings.ToLower(level) {
	case "debug":
		l = l.Level(zerolog.DebugLevel)
	case "info":
		l = l.Level(zerolog.InfoLevel)
	case "warn", "warning":
		l = l.Level(zerolog.WarnLevel)
	case "error":
		l = l.Level(zerolog.ErrorLevel)
	default:
		// если неизвестный — остаёмся на info
		l = l.Level(zerolog.InfoLevel)
	}
	// return pointer for easy use in other packages
	logger := l
	return &logger
}

// For tests or redirecting logs one can use NewWithWriter
func NewWithWriter(level string, w io.Writer) *zerolog.Logger {
	out := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339}
	l := zerolog.New(out).With().Timestamp().Logger()
	switch strings.ToLower(level) {
	case "debug":
		l = l.Level(zerolog.DebugLevel)
	case "info":
		l = l.Level(zerolog.InfoLevel)
	case "warn", "warning":
		l = l.Level(zerolog.WarnLevel)
	case "error":
		l = l.Level(zerolog.ErrorLevel)
	default:
		l = l.Level(zerolog.InfoLevel)
	}
	logger := l
	return &logger
}
