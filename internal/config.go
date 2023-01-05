package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_port", "8080")
	config.RegisterDefault("http_address", "0.0.0.0")
}

type Config struct {
	config.Base `config:",squash"`

	// configs for starting and introspecting service
	GitRef      string `config:"git_ref" validate:"required"`
	HTTPPort    string `config:"http_port" validate:"required"`
	HTTPAddress string `config:"http_address" validate:"required"`

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
	GithubAppID  string `config:"github_app_id"`
	GithubAppKey string `config:"github_app_key"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
