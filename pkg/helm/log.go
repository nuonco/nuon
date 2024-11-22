package helm

import (
	"fmt"

	"go.uber.org/zap"
)

func Logger(l *zap.Logger) func(string, ...interface{}) {
	return func(format string, vs ...interface{}) {
		msg := fmt.Sprintf(format, vs...)
		l.Info(msg)
	}
}
