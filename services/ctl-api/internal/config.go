package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_port", "8080")
	config.RegisterDefault("http_address", "0.0.0.0")

	// defaults for database
	config.RegisterDefault("db_region", "us-west-2")
	config.RegisterDefault("db_port", 5432)
	config.RegisterDefault("db_user", "ctl_api")
	config.RegisterDefault("db_name", "ctl_api")

	// defaults for app
	config.RegisterDefault("temporal_namespace", "default")
	config.RegisterDefault("github_app_key_secret_name", "ctl-api-github-app-key")
	config.RegisterDefault("sandbox_artifacts_base_url", "https://nuon-artifacts.s3.us-west-2.amazonaws.com/sandbox")

	// defaults for sandbox mode
	config.RegisterDefault("sandbox_sleep", "5s")
}

type Config struct {
	worker.Config `config:",squash"`

	// configs for starting and introspecting service
	GitRef           string `config:"git_ref" validate:"required"`
	ServiceName      string `config:"service_name" validate:"required"`
	HTTPPort         string `config:"http_port" validate:"required"`
	InternalHTTPPort string `config:"internal_http_port" validate:"required"`

	// database connection parameters
	DBName     string `config:"db_name" validate:"required"`
	DBHost     string `config:"db_host" validate:"required"`
	DBPort     string `config:"db_port" validate:"required"`
	DBSSLMode  string `config:"db_ssl_mode" validate:"required"`
	DBPassword string `config:"db_password"`
	DBUser     string `config:"db_user" validate:"required"`
	DBZapLog   bool   `config:"db_use_zap"`
	DBUseIAM   bool   `config:"db_use_iam"`
	DBRegion   string `config:"db_region" validate:"required"`

	// temporal configuration
	TemporalHost      string `config:"temporal_host"  validate:"required"`
	TemporalNamespace string `config:"temporal_namespace" validate:"required"`

	// github configuration
	GithubAppID            string `config:"github_app_id" validate:"required"`
	GithubAppKey           string `config:"github_app_key" validate:"required"`
	GithubAppKeySecretName string `config:"github_app_key_secret_name" validate:"required"`

	// sandbox artifacts
	SandboxArtifactsBaseURL string `config:"sandbox_artifacts_base_url" validate:"required"`

	// middleware configuration
	Middlewares         []string `config:"middlewares"`
	InternalMiddlewares []string `config:"internal_middlewares"`

	// auth 0 config
	Auth0IssuerURL string `config:"auth0_issuer_url" validate:"required"`
	Auth0Audience  string `config:"auth0_audience" validate:"required"`
	Auth0ClientID  string `config:"auth0_client_id" validate:"required"`

	// flags for controlling the background workers
	ForceSandboxMode   bool          `config:"force_sandbox_mode"`
	SandboxSleep       time.Duration `config:"sandbox_sleep" validate:"required"`
	TFEToken           string        `config:"tfe_token" validate:"required"`
	TFEOrgsWorkspaceID string        `config:"tfe_orgs_workspace_id" validate:"required"`

	// flags for controlling creation of integration users
	IntegrationGithubInstallID string `config:"integration_github_install_id" validate:"required"`
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
