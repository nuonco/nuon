package internal

import (
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "canary")

	// configuration for testing
	config.RegisterDefault("install_script_path", "/install-cli.sh")
	config.RegisterDefault("tests_dir", "/tests")

	// configuration for terraform run
	config.RegisterDefault("terraform_module_dir", "/terraform")
	config.RegisterDefault("terraform_state_base_dir", "/tmp/state")
	config.RegisterDefault("default_install_count", 1)
	config.RegisterDefault("sandbox_mode_install_count", 5)
	config.RegisterDefault("canary_debug_period", "4h")
}

type Config struct {
	worker.Config `config:",squash"`

	SlackWebhookURL  string `config:"slack_webhook_url" validate:"required"`
	AWSEKSIAMRoleArn string `config:"aws_eks_iam_role_arn" validate:"required"`
	AWSECSIAMRoleArn string `config:"aws_ecs_iam_role_arn" validate:"required"`

	APIURL         string `config:"api_url" validate:"required"`
	InternalAPIURL string `config:"internal_api_url" validate:"required"`

	TerraformModuleDir    string `config:"terraform_module_dir" validate:"required"`
	TerraformStateBaseDir string `config:"terraform_state_base_dir" validate:"required"`
	InstallScriptPath     string `config:"install_script_path" validate:"required"`
	TestsDir              string `config:"tests_dir" validate:"required"`

	// flags for local development
	DisableNotifications bool `config:"disable_notifications"`
	DisableTests         bool `config:"disable_tests"`

	SandboxModeInstallCount int `config:"sandbox_mode_install_count"`
	DefaultInstallCount     int `config:"default_install_count"`

	StateBucketName   string `config:"state_bucket_name"`
	StateBucketRegion string `config:"state_bucket_region"`

	// the canary will _not_ run deprovision until the period has finished, allowing us time to debug using the CLI.
	CanaryDebugPeriod string `config:"canary_debug_period"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
