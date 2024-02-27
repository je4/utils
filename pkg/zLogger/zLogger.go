package zLogger

import (
	"github.com/rs/zerolog"
	"strings"
)

type ZLogger interface {
	Trace() *zerolog.Event

	Debug() *zerolog.Event

	Info() *zerolog.Event

	Warn() *zerolog.Event

	Error() *zerolog.Event

	Err(err error) *zerolog.Event

	Fatal() *zerolog.Event

	Panic() *zerolog.Event
}

func LogLevel(str string) zerolog.Level {
	switch strings.ToUpper(str) {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	case "PANIC":
		return zerolog.PanicLevel
	default:
		return zerolog.DebugLevel
	}
}
