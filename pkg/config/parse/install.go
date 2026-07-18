package parse

import (
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/pkg/config"
)

func ParseInstallConfig(r io.Reader, v *validator.Validate) (*config.Install, error) {
	decoder := toml.NewDecoder(r)

	obj := make(map[string]interface{})
	if err := decoder.Decode(&obj); err != nil {
		return nil, fmt.Errorf("error decoding TOML: %w", err)
	}

	var cfg config.Install
	decCfg := config.DecoderConfig()
	decCfg.Result = &cfg
	dec, err := mapstructure.NewDecoder(decCfg)
	if err != nil {
		return nil, err
	}

	if err := dec.Decode(obj); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	if err := cfg.Parse(); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return &cfg, nil
}
