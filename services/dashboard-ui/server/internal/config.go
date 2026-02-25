package internal

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/services/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_port", "4040")
	config.RegisterDefault("log_level", "INFO")
	config.RegisterDefault("dashboard_dev", false)
	config.RegisterDefault("service_name", "dashboard-ui")
	config.RegisterDefault("dist_dir", "./dist")
	config.RegisterDefault("public_dir", "./public")
	config.RegisterDefault("middlewares", []string{
		"cors",
	})
}

type Config struct {
	HTTPPort    string   `config:"http_port" validate:"required"`
	LogLevel    string   `config:"log_level"`
	DashboardDev bool    `config:"dashboard_dev"`
	ServiceName string   `config:"service_name"`
	Version     string   `config:"version"`
	GitRef      string   `config:"git_ref"`
	Middlewares []string `config:"middlewares"`
	DistDir     string   `config:"dist_dir"`
	PublicDir   string   `config:"public_dir"`
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
