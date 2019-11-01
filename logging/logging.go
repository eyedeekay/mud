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

func (l loggerImpl) Info(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	msg := v[0].(string)
	args := v[1:]
	l.Infow(msg, args...)
}

func (l loggerImpl) Warn(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	msg := v[0].(string)
	args := v[1:]
	l.Warnw(msg, args...)
}

func (l loggerImpl) Err(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	msg := v[0].(string)
	args := v[1:]
	l.Errorw(msg, args...)
}

func (l loggerImpl) Fatal(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	msg := v[0].(string)
	args := v[1:]
	l.Fatalw(msg, args...)
}
