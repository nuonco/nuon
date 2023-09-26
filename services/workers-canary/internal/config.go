package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_namespace", "canary")
	config.RegisterDefault("terraform_state_path", "/tmp/nuonctl-canary.tfstate")
	config.RegisterDefault("install_script_path", "/install-cli.sh")
	config.RegisterDefault("terraform_module_dir", "/terraform")
}

type Config struct {
	worker.Config `config:",squash"`

	SlackWebhookURL   string `config:"slack_webhook_url"	 validate:"required"`
	InstallIamRoleArn string `config:"install_iam_role_arn"  validate:"required"`

	APIURL          string `config:"api_url" validate:"required"`
	APIToken        string `config:"nuon_api_token" validate:"required"`
	GithubInstallID string `config:"github_install_id" validate:"required"`

	TerraformModuleDir string `config:"terraform_module_dir" validate:"required"`
	TerraformStatePath string `config:"terraform_state_path" validate:"required"`
	InstallScriptPath  string `config:"install_script_path" validate:"required"`

	// flags for local development
	DisableNotifications bool `config:"disable_notifications"`
	DisableCLICommands   bool `config:"disable_cli_commands"`
	DisableIntrospection bool `config:"disable_introspection"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
