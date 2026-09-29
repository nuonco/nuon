package preflight

import (
	"fmt"

	svcconfig "github.com/nuonco/nuon/pkg/services/config"
	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

func LoadConfig() (*internal.Config, error) {
	var cfg internal.Config
	if err := svcconfig.LoadInto(nil, &cfg); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	return &cfg, nil
}
