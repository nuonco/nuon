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
	config.RegisterDefault("bundle_dir", "/bundle")
	config.RegisterDefault("registry_dir", "/tmp/runner-registry")
	config.RegisterDefault("log_level", "INFO")
	config.RegisterDefault("registry_port", "5001")
	config.RegisterDefault("sandbox_job_duration", "5s")
}

type Config struct {
	GitRef string `config:"git_ref" validate:"required"`

	RunnerAPIURL   string `config:"runner_api_url" validate:"required"`
	RunnerAPIToken string `config:"runner_api_token" validate:"required"`
	RunnerID       string `config:"runner_id" validate:"required"`

	// observability configuration
	HostIP   string `config:"host_ip" validate:"required"`
	LogLevel string `config:"log_level"`

	// some artifacts are bundled into the runner binary, to make loading them easier.
	BundleDir    string `config:"bundle_dir" validate:"required"`
	RegistryDir  string `config:"registry_dir" validate:"required"`
	RegistryPort int    `config:"registry_port" validate:"required"`

	// only for enabling local things
	IsNuonctl                bool          `config:"is_nuonctl"`
	SandboxJobDuration       time.Duration `config:"sandbox_job_duration"`
	SandboxModeFaultsEnabled bool          `config:"sandbox_mode_faults_enabled"`
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
