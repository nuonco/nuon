package parse

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
)

type ParseConfig struct {
	Bytes       []byte
	BackendType config.BackendType
	V           *validator.Validate
	Template    bool
	Context     config.ConfigContext
}

func Parse(parseCfg ParseConfig) (*config.AppConfig, error) {
	var cfg config.AppConfig

	byts, err := Template(parseCfg.Bytes)
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(byts, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse toml config: %w", err)
	}

	if err := cfg.Parse(parseCfg.Context); err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	return &cfg, nil
}
