package zLogger

import "fmt"

type ZWrapper interface {
	Error(args ...any)
	Errorf(msg string, args ...any)
	Warning(args ...any)
	Warningf(msg string, args ...any)
	Info(args ...any)
	Infof(msg string, args ...any)
	Debug(args ...any)
	Debugf(msg string, args ...any)
	Fatal(args ...any)
	Fatalf(msg string, args ...any)
	Panic(args ...any)
	Panicf(msg string, args ...any)
}

func NewZWrapper(z ZLogger) ZWrapper {
	return &zWrapper{z}
}

type zWrapper struct{ ZLogger }

func (z *zWrapper) Error(args ...any) {
	z.ZLogger.Error().Msg(fmt.Sprint(args))
}
func (z *zWrapper) Errorf(msg string, args ...any) {
	z.ZLogger.Error().Msgf(msg, args...)
}

func (z *zWrapper) Warning(args ...any) {
	z.ZLogger.Warn().Msg(fmt.Sprint(args))
}
func (z *zWrapper) Warningf(msg string, args ...any) {
	z.ZLogger.Warn().Msgf(msg, args...)
}

func (z *zWrapper) Info(args ...any) {
	z.ZLogger.Info().Msg(fmt.Sprint(args))
}
func (z *zWrapper) Infof(msg string, args ...any) {
	z.ZLogger.Info().Msgf(msg, args...)
}

func (z *zWrapper) Debug(args ...any) {
	z.ZLogger.Debug().Msg(fmt.Sprint(args))
}
func (z *zWrapper) Debugf(msg string, args ...any) {
	z.ZLogger.Debug().Msgf(msg, args...)
}

func (z *zWrapper) Trace(args ...any) {
	z.ZLogger.Trace().Msg(fmt.Sprint(args))
}
func (z *zWrapper) Tracef(msg string, args ...any) {
	z.ZLogger.Trace().Msgf(msg, args...)
}

func (z *zWrapper) Fatal(args ...any) {
	z.ZLogger.Fatal().Msg(fmt.Sprint(args))
}
func (z *zWrapper) Fatalf(msg string, args ...any) {
	z.ZLogger.Fatal().Msgf(msg, args...)
}

func (z *zWrapper) Panic(args ...any) {
	z.ZLogger.Panic().Msg(fmt.Sprint(args))
}
func (z *zWrapper) Panicf(msg string, args ...any) {
	z.ZLogger.Panic().Msgf(msg, args...)
}

var _ ZWrapper = (*zWrapper)(nil)
