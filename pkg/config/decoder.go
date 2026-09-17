package config

import (
	"github.com/mitchellh/mapstructure"
)

func DecoderConfig(opts ...ParseOption) *mapstructure.DecoderConfig {
	cfg := parseOptions(opts...)
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(DecodeSource, DecodeComponent(cfg.RootDir), DecodeInstallInputs),
	}
}
