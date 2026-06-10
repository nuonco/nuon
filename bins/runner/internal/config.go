package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/services/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("runner_api_url", "https://api.nuon.co")
	config.RegisterDefault("bundle_dir", "/bundle")
	config.RegisterDefault("registry_dir", "/tmp/runner-registry")
	config.RegisterDefault("log_level", "INFO")
	config.RegisterDefault("registry_port", "5001")
	config.RegisterDefault("sandbox_job_duration", "5s")
	config.RegisterDefault("sandbox_control_port", "9095")
	config.RegisterDefault("health_port", "9999")
	config.RegisterDefault("git_ref", "dev")
}

type Config struct {
	GitRef string `config:"git_ref" validate:"required"`

	RunnerAPIURL     string `config:"runner_api_url" validate:"required"`
	RunnerAPIToken   string `config:"runner_api_token"`
	RunnerID         string `config:"runner_id" validate:"required"`
	RunnerPlatform   string `config:"runner_platform"`
	RunnerAuthMethod string `config:"runner_auth_method"` // default if "" or "iid" for aws

	// observability configuration
	HostIP   string `config:"host_ip" validate:"required"`
	LogLevel string `config:"log_level"`

	// some artifacts are bundled into the runner binary, to make loading them easier.
	BundleDir    string `config:"bundle_dir" validate:"required"`
	RegistryDir  string `config:"registry_dir" validate:"required"`
	RegistryPort int    `config:"registry_port" validate:"required"`

	// HealthPort is the port the mng process serves its /healthz endpoint on.
	// Azure's VMSS Application Health extension probes this to drive automatic
	// instance repair (the self-heal analog to the AWS ASG EC2 health check).
	HealthPort int `config:"health_port" validate:"required"`

	// kubernetes pod identity and self-deletion
	PodName             string `config:"pod_name"`
	PodNamespace        string `config:"pod_namespace"`
	DeploymentName      string `config:"deployment_name"`
	DeletePodOnShutdown bool   `config:"delete_pod_on_shutdown"`

	// only for enabling local things
	IsNuonctl                bool          `config:"is_nuonctl"`
	SandboxJobDuration       time.Duration `config:"sandbox_job_duration"`
	SandboxModeFaultsEnabled bool          `config:"sandbox_mode_faults_enabled"`
	SandboxControlPort       int           `config:"sandbox_control_port"`

	// TerraformMirrorPlatforms overrides the `<os>_<arch>` platform set
	// the build runner vendors providers for via `terraform providers
	// mirror`. When empty, the build runner defaults to the runner's own
	// runtime platform (one entry, runtime.GOOS_runtime.GOARCH).
	//
	// Set via env var TERRAFORM_MIRROR_PLATFORMS as a comma-separated
	// list, e.g. `linux_amd64,linux_arm64,darwin_arm64`. Useful when
	// vendoring artifacts that need to be consumed by install runners
	// on a different platform than the build runner (heterogeneous
	// orgs, cross-arch dev/test).
	TerraformMirrorPlatforms []string `config:"terraform_mirror_platforms"`
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
