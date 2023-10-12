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
}

type Config struct {
	worker.Config `config:",squash"`

	// configs for starting and introspecting service
	GitRef           string `config:"git_ref" validate:"required"`
	ServiceName      string `config:"service_name" validate:"required"`
	HTTPPort         string `config:"http_port" validate:"required"`
	InternalHTTPPort string `config:"internal_http_port" validate:"required"`

	// database connection parameters
	DBName     string `config:"db_name"`
	DBHost     string `config:"db_host"`
	DBPort     string `config:"db_port"`
	DBSSLMode  string `config:"db_ssl_mode"`
	DBPassword string `config:"db_password"`
	DBUser     string `config:"db_user"`
	DBZapLog   bool   `config:"db_use_zap"`
	DBUseIAM   bool   `config:"db_use_iam"`
	DBRegion   string `config:"db_region"`

	// temporal configuration
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	// github configuration
	GithubAppID            string `config:"github_app_id"`
	GithubAppKey           string `config:"github_app_key"`
	GithubAppKeySecretName string `config:"github_app_key_secret_name"`

	// sandbox artifacts
	SandboxArtifactsBaseURL string `config:"sandbox_artifacts_base_url"`

	// middleware configuration
	Middlewares         []string `config:"middlewares"`
	InternalMiddlewares []string `config:"internal_middlewares"`

	// auth 0 config
	Auth0IssuerURL string `config:"auth0_issuer_url"`
	Auth0Audience  string `config:"auth0_audience"`
	Auth0ClientID  string `config:"auth0_client_id"`

	// flags for controlling the background workers
	ForceSandboxMode bool          `config:"force_sandbox_mode"`
	SandboxSleep     time.Duration `config:"sandbox_sleep"`
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
