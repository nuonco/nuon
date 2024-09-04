package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/services/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("runner_api_url", "https://api.nuon.co")
}

type Config struct {
	GitRef  string `config:"git_ref" validate:"required"`
	Version string `config:"version" validate:"required"`

	RunnerAPIURL   string `config:"runner_api_url" validate:"required"`
	RunnerAPIToken string `config:"runner_api_token" validate:"required"`
	RunnerID       string `config:"runner_id" validate:"required"`

	SettingsRefreshTimeout time.Duration `config:"settings_refresh_timeout" validate:"required"`

	// observability configuration
	HostIP   string `config:"host_ip" validate:"required"`
	LogLevel string `config:"log_level"`

	// some artifacts are bundled into the runner binary, to make loading them easier.
	BundleDir string `config:"bundle_dir"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := config.LoadInto(nil, &cfg); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate config: %w", err)
	}

	return &cfg, nil
}
