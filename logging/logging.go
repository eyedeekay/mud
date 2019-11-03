package logging

import "go.uber.org/zap"

type Logger interface {
	Info(v ...interface{})
	Warn(v ...interface{})
	Err(v ...interface{})
	Fatal(v ...interface{})
}

type LogLevel int

const (
	None LogLevel = 0 + iota
	Dev
	Prod
)

func NewLogger(level LogLevel) (Logger, error) {
	var logger *zap.Logger
	var err error
	switch level {
	case Dev:
		logger, err = zap.NewDevelopment()
	case Prod:
		logger, err = zap.NewProduction()
	case None:
		fallthrough
	default:
		logger = zap.NewNop()
	}
	if err != nil {
		return nil, err
	}

	return &loggerImpl{logger.Sugar()}, nil
}

type loggerImpl struct {
	*zap.SugaredLogger
}

type logFunc func(msg string, keysAndValues ...interface{})

func log(f logFunc, v ...interface{}) {
	if len(v) == 0 {
		return
	}
	msg := v[0].(string)
	args := v[1:]
	f(msg, args...)
}

func (l loggerImpl) Info(v ...interface{}) {
	log(l.Infow, v...)
}

func (l loggerImpl) Warn(v ...interface{}) {
	log(l.Warnw, v...)
}

func (l loggerImpl) Err(v ...interface{}) {
	log(l.Errorw, v...)
}

func (l loggerImpl) Fatal(v ...interface{}) {
	log(l.Fatalw, v...)
}
