package log

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
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
