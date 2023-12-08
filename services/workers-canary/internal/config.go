package internal

import (
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "canary")
	config.RegisterDefault("terraform_state_base_dir", "/tmp/state")
	config.RegisterDefault("install_script_path", "/install-cli.sh")
	config.RegisterDefault("terraform_module_dir", "/terraform")
	config.RegisterDefault("tests_dir", "/tests")
	config.RegisterDefault("disable_cli_commands", false)
}

type Config struct {
	worker.Config `config:",squash"`

	SlackWebhookURL   string `config:"slack_webhook_url"	 validate:"required"`
	InstallIamRoleArn string `config:"install_iam_role_arn"  validate:"required"`

	APIURL         string `config:"api_url" validate:"required"`
	InternalAPIURL string `config:"internal_api_url" validate:"required"`

	TerraformModuleDir    string `config:"terraform_module_dir" validate:"required"`
	TerraformStateBaseDir string `config:"terraform_state_base_dir" validate:"required"`
	InstallScriptPath     string `config:"install_script_path" validate:"required"`
	TestsDir              string `config:"tests_dir" validate:"required"`

	// flags for local development
	DisableNotifications bool `config:"disable_notifications"`
	DisableTests         bool `config:"disable_tests"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
