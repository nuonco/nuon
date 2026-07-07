package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pkg/errors"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
)

func NewDev(cfg *runnerconfig.Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}

	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)

	dev, err := config.Build()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get zap development")
	}

	return dev, nil
}
