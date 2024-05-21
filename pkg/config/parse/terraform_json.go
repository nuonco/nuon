package parse

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/powertoolsdev/mono/pkg/config"
)

func ToTerraformJSON(parseCfg ParseConfig) ([]byte, error) {
	if parseCfg.Context != config.ConfigContextSource {
		return nil, fmt.Errorf("Must use source context for terraform")
	}

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

	if err := cfg.Validate(parseCfg.V); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	tfJSON, err := cfg.ToTerraformJSON(parseCfg.BackendType)
	if err != nil {
		return nil, fmt.Errorf("unable to generate terraform json: %w", err)
	}

	return tfJSON, nil
}
