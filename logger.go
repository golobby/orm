package orm

import (
	"fmt"

	"go.uber.org/zap"
)

type LogLevel int

const (
	LogLevelDev LogLevel = iota
	LogLevelProd
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type zapLogger struct {
	l *zap.SugaredLogger
}

func newZapLogger(env LogLevel) (*zapLogger, error) {
	if env == LogLevelDev {
		l, err := zap.NewDevelopmentConfig().Build()
		if err != nil {
			return nil, err
		}
		return &zapLogger{l.Sugar()}, nil
	} else if env == LogLevelProd {
		l, err := zap.NewProductionConfig().Build()
		if err != nil {
			return nil, err
		}
		return &zapLogger{l.Sugar()}, nil
	} else {
		return nil, fmt.Errorf("log level should be either LogLevelDev or LogLevelProd")
	}
}

func (z *zapLogger) Debugf(format string, args ...any) {
	format = fmt.Sprintf("[DEBUG] %s", format)
	z.l.Debugf(format, args...)
}
func (z *zapLogger) Warnf(format string, args ...any) {
	format = fmt.Sprintf("[WARNF] %s", format)
	z.l.Warnf(format, args...)

}
func (z *zapLogger) Errorf(format string, args ...any) {
	format = fmt.Sprintf("[ERROR] %s", format)
	z.l.Errorf(format, args...)
}

func (z *zapLogger) Infof(format string, args ...any) {
	format = fmt.Sprintf("[INFO] %s", format)
	z.l.Infof(format, args...)
}
