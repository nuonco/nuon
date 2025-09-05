package config

import (
	"github.com/mitchellh/mapstructure"
)

func DecoderConfig() *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(DecodeSource, DecodeComponent, DecodeInstallInputs),
	}
}
