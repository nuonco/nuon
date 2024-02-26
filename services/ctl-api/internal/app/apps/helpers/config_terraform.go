package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers/config"
	"gopkg.in/yaml.v2"
)

func (h *Helpers) GenerateTerraformJSONFromConfig(ctx context.Context, inputCfg []byte, configFmt app.AppConfigFmt) ([]byte, error) {
	var cfg config.AppConfig

	switch configFmt {
	case app.AppConfigFmtJson:
		if err := json.Unmarshal(inputCfg, &cfg); err != nil {
			return nil, fmt.Errorf("unable to parse json config: %w", err)
		}
	case app.AppConfigFmtYaml:
		if err := yaml.Unmarshal(inputCfg, &cfg); err != nil {
			return nil, fmt.Errorf("unable to parse yaml config: %w", err)
		}
	case app.AppConfigFmtToml:
		if err := toml.Unmarshal(inputCfg, &cfg); err != nil {
			return nil, fmt.Errorf("unable to parse toml config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config type: %s", configFmt)
	}

	if err := cfg.Validate(h.v); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	tfJSON, err := cfg.ToTerraformJSON(h.cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("unable to generate terraform json: %w", err)
	}

	return tfJSON, nil
}
