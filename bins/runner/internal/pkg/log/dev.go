package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/bins/runner/internal"
)

func NewDev(cfg *internal.Config) (*zap.Logger, error) {
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
