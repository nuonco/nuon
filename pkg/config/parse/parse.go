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
}

func ToTerraformJSON(parseCfg ParseConfig) ([]byte, error) {
	var cfg config.AppConfig

	byts, err := Template(parseCfg.Bytes)
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(byts, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse toml config: %w", err)
	}

	if err := cfg.Parse(); err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	if err := cfg.Validate(parseCfg.V); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	tfJSON, err := cfg.ToTerraformJSON(parseCfg.BackendType)
	if err != nil {
		return nil, fmt.Errorf("unable to generate terraform json: %w", err)
	}

	return tfJSON, nil
}
