package preflight

import (
	"fmt"

	svcconfig "github.com/nuonco/nuon/pkg/services/config"
	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

// LoadConfig loads configuration without running struct validation.
//
// internal.NewConfig would reject a partial config before any check could run,
// which is the opposite of what preflight is for: each check validates only the
// fields it needs and reports them by name.
func LoadConfig() (*internal.Config, error) {
	var cfg internal.Config
	if err := svcconfig.LoadInto(nil, &cfg); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	return &cfg, nil
}
