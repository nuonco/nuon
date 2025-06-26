package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// flagger represents anything that can return a pointer to a pflag FlagSet
// typically, this would be a *cobra.Command
type flagger interface {
	Flags() *pflag.FlagSet
}

const configureServiceErrTemplate = `{"level":"error","ts":%d,"msg":"failed to setup service", "error": "%s"}\n`

func loadConfig(cmd flagger) (*Base, error) {
	var cfg Base

	if err := LoadInto(cmd.Flags(), &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}

func configureLogger(cfg *Base) (*zap.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	switch cfg.Env {
	case Development:
		l, err = zap.NewDevelopment()
	default:
		zCfg := zap.NewProductionConfig()

		var lvl zapcore.Level
		lvl, err = zapcore.ParseLevel(cfg.LogLevel)
		if err == nil {
			// only set the level if it was set correctly on the config
			zCfg.Level.SetLevel(lvl)
		}

		l, err = zCfg.Build()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate logger: %w", err)
	}

	zap.ReplaceGlobals(l)

	return l, nil
}
