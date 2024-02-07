package log

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
)

func New(cfg *internal.Config) (*zap.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	l, err = zap.NewProduction()
	if cfg.LogLevel == "DEBUG" {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logger: %w", err)
	}
	zap.ReplaceGlobals(l)
	return l, nil
}
