package internal

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/services/config"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_port", "4000")
	config.RegisterDefault("log_level", "INFO")
	config.RegisterDefault("dashboard_dev", false)
	config.RegisterDefault("service_name", "dashboard-ui")
	config.RegisterDefault("dist_dir", "./dist")
	config.RegisterDefault("public_dir", "./public")
	config.RegisterDefault("middlewares", []string{
		"cors",
	})
	config.RegisterDefault("nuon_api_url", "https://api.nuon.co")
	config.RegisterDefault("nuon_app_url", "http://localhost:4000")
	config.RegisterDefault("github_app_name", "nuon-connect")
}

type Config struct {
	HTTPPort     string   `config:"http_port" validate:"required"`
	LogLevel     string   `config:"log_level"`
	DashboardDev bool     `config:"dashboard_dev"`
	ServiceName  string   `config:"service_name"`
	Version      string   `config:"version"`
	GitRef       string   `config:"git_ref"`
	Middlewares  []string `config:"middlewares"`
	DistDir      string   `config:"dist_dir"`
	PublicDir    string   `config:"public_dir"`

	APIUrl                string `config:"nuon_api_url"`
	AdminAPIUrl           string `config:"nuon_admin_api_url"`
	TemporalUIUrl         string `config:"nuon_temporal_ui_url"`
	AuthServiceUrl        string `config:"nuon_auth_service_url"`
	AppUrl                string `config:"nuon_app_url"`
	GithubAppName         string `config:"github_app_name"`
	PylonAppID            string `config:"pylon_app_id"`
	DatadogEnv            string `config:"datadog_env"`
	DatadogAPIKey         string `config:"datadog_api_key"`
	DatadogApplicationKey string `config:"datadog_application_key"`
	DatadogTraceDebug     bool   `config:"datadog_trace_debug"`
	DatadogAPIUrl         string `config:"datadog_api_url"`
	IsBYOC                bool   `config:"nuon_byoc"`
	SFTrialEndpoint       string `config:"sf_trial_access_endpoint"`
	OnboardingV2          bool   `config:"nuon_onboarding_v2"`
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
