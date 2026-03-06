package log

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewFXLog() (fxevent.Logger, error) {
	zl, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	fxLog := &fxevent.ZapLogger{Logger: zl}
	return fxLog, nil
}
