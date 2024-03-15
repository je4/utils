package zLogger

type ZWrapper interface {
	Error(msg string)
	Errorf(msg string, args ...any)
	Warning(msg string)
	Warningf(msg string, args ...any)
	Info(msg string)
	Infof(msg string, args ...any)
	Debug(msg string)
	Debugf(msg string, args ...any)
	Trace(msg string)
	Tracef(msg string, args ...any)
	Fatal(msg string)
	Fatalf(msg string, args ...any)
	Panic(msg string)
	Panicf(msg string, args ...any)
}

func NewZWrapper(z ZLogger) ZWrapper {
	return &zWrapper{z}
}

type zWrapper struct{ ZLogger }

func (z *zWrapper) Error(msg string) {
	z.ZLogger.Error().Msg(msg)
}
func (z *zWrapper) Errorf(msg string, args ...any) {
	z.ZLogger.Error().Msgf(msg, args...)
}

func (z *zWrapper) Warning(msg string) {
	z.ZLogger.Warn().Msg(msg)
}
func (z *zWrapper) Warningf(msg string, args ...any) {
	z.ZLogger.Warn().Msgf(msg, args...)
}

func (z *zWrapper) Info(msg string) {
	z.ZLogger.Info().Msg(msg)
}
func (z *zWrapper) Infof(msg string, args ...any) {
	z.ZLogger.Info().Msgf(msg, args...)
}

func (z *zWrapper) Debug(msg string) {
	z.ZLogger.Debug().Msg(msg)
}
func (z *zWrapper) Debugf(msg string, args ...any) {
	z.ZLogger.Debug().Msgf(msg, args...)
}

func (z *zWrapper) Trace(msg string) {
	z.ZLogger.Trace().Msg(msg)
}
func (z *zWrapper) Tracef(msg string, args ...any) {
	z.ZLogger.Trace().Msgf(msg, args...)
}

func (z *zWrapper) Fatal(msg string) {
	z.ZLogger.Fatal().Msg(msg)
}
func (z *zWrapper) Fatalf(msg string, args ...any) {
	z.ZLogger.Fatal().Msgf(msg, args...)
}

func (z *zWrapper) Panic(msg string) {
	z.ZLogger.Panic().Msg(msg)
}
func (z *zWrapper) Panicf(msg string, args ...any) {
	z.ZLogger.Panic().Msgf(msg, args...)
}

var _ ZWrapper = (*zWrapper)(nil)
