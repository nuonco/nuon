package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
	"go.uber.org/zap"
)

func (w *worker) getLogger() (*zap.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	switch w.Config.Env {
	case config.Local, config.Development:
		l, err = zap.NewDevelopment()
	default:
		l, err = zap.NewProduction()
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create logger: %w", err)
	}

	zap.ReplaceGlobals(l)
	return l, nil
}
