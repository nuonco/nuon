package log

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewFXLog() (fxevent.Logger, error) {
	zl, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &fxevent.ZapLogger{
		Logger: zl,
	}, nil
}
