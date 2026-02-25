package internal

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/services/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_port", "4040")
	config.RegisterDefault("nuon_api_url", "https://api.stage.nuon.co")
	config.RegisterDefault("log_level", "INFO")
	config.RegisterDefault("dashboard_dev", false)
	config.RegisterDefault("disable_metrics", false)
	config.RegisterDefault("service_name", "dashboard-ui")
	config.RegisterDefault("service_type", "bff")
	config.RegisterDefault("service_deployment", "dashboard")
	config.RegisterDefault("admin_api_url", "http://localhost:8082")
	config.RegisterDefault("temporal_ui_url", "http://temporal-web.temporal.svc.cluster.local:8080")
	config.RegisterDefault("dist_dir", "./dist")
	config.RegisterDefault("public_dir", "./public")
	config.RegisterDefault("middlewares", []string{
		"panicker",
		"metrics",
		"tracer",
		"cors",
		"log",
		"auth",
		"apiclient",
	})
}

type Config struct {
	HTTPPort          string   `config:"http_port" validate:"required"`
	NuonAPIURL        string   `config:"nuon_api_url" validate:"required"`
	LogLevel          string   `config:"log_level"`
	DashboardDev      bool     `config:"dashboard_dev"`
	DisableMetrics    bool     `config:"disable_metrics"`
	ServiceName       string   `config:"service_name"`
	ServiceType       string   `config:"service_type"`
	ServiceDeployment string   `config:"service_deployment"`
	Version           string   `config:"version"`
	GitRef            string   `config:"git_ref"`
	Middlewares       []string `config:"middlewares"`
	AdminAPIURL       string   `config:"admin_api_url"`
	TemporalUIURL     string   `config:"temporal_ui_url"`
	DistDir           string   `config:"dist_dir"`
	PublicDir         string   `config:"public_dir"`
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
