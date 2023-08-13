package internal

import (
	"fmt"

	"github.com/go-playground/validator"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_port", "8080")
	config.RegisterDefault("http_address", "0.0.0.0")

	// defaults for database
	config.RegisterDefault("db_user", "api")
	config.RegisterDefault("db_port", "5432")
	config.RegisterDefault("db_name", "api")
	config.RegisterDefault("db_ssl_mode", "disable")
	config.RegisterDefault("db_host", "localhost")
	config.RegisterDefault("db_use_zap", false)
	config.RegisterDefault("db_use_iam", false)
	config.RegisterDefault("db_region", "us-west-2")
	config.RegisterDefault("db_migrations_path", "./migrations")

	config.RegisterDefault("temporal_namespace", "default")

	// default for github
	config.RegisterDefault("github_app_key_secret_name", "graphql-api-github-app-key")

	// default sandbox url
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
	DBName           string `config:"db_name"`
	DBHost           string `config:"db_host"`
	DBPort           string `config:"db_port"`
	DBSSLMode        string `config:"db_ssl_mode"`
	DBPassword       string `config:"db_password"`
	DBUser           string `config:"db_user"`
	DBZapLog         bool   `config:"db_use_zap"`
	DBUseIAM         bool   `config:"db_use_iam"`
	DBRegion         string `config:"db_region"`
	DBMigrationsPath string `config:"db_migrations_path"`

	// temporal configuration
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	// github configuration
	GithubAppID            string `config:"github_app_id"`
	GithubAppKey           string `config:"github_app_key"`
	GithubAppKeySecretName string `config:"github_app_key_secret_name"`

	// sandbox artifacts
	SandboxArtifactsBaseURL string `config:"sandbox_artifacts_base_url"`
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
