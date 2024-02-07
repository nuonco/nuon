package worker

import (
	"fmt"

	"go.uber.org/zap"
)

func (w *worker) getLogger() (*zap.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	l, err = zap.NewProduction()
	if w.Config.LogLevel == "DEBUG" {
		l, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create logger: %w", err)
	}

	zap.ReplaceGlobals(l)
	return l, nil
}
