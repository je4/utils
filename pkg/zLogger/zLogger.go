package zLogger

import (
	"github.com/rs/zerolog"
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
