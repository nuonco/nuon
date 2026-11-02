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
	config.RegisterDefault("service_name", "dashboard-ui")
	config.RegisterDefault("dist_dir", "./dist")
	config.RegisterDefault("public_dir", "./public")
	config.RegisterDefault("middlewares", []string{
		"metrics",
		"cors",
	})
	config.RegisterDefault("nuon_api_url", "https://api.nuon.co")
	config.RegisterDefault("nuon_app_url", "http://localhost:4000")
	config.RegisterDefault("github_app_name", "nuon-connect")
	config.RegisterDefault("disable_metrics", true)
	config.RegisterDefault("service_deployment", "local")
	config.RegisterDefault("cors_allow_loopback_origins", false)
	config.RegisterDefault("nuon_admin_dashboard_url", "http://localhost:8087")
	config.RegisterDefault("posthog_host", "https://us.i.posthog.com")
	config.RegisterDefault("posthog_replay_enabled", true)
	// docker-compose's kafka-ui, which serves under the same /admin/kafka context
	// path it does in the cluster
	config.RegisterDefault("nuon_kafka_ui_url", "http://localhost:8092")
	config.RegisterDefault("nuon_grafana_ui_url", "")
}

type Config struct {
	HTTPPort    string   `config:"http_port" validate:"required"`
	LogLevel    string   `config:"log_level"`
	ServiceName string   `config:"service_name"`
	Version     string   `config:"version"`
	GitRef      string   `config:"git_ref"`
	Middlewares []string `config:"middlewares"`
	DistDir     string   `config:"dist_dir"`
	PublicDir   string   `config:"public_dir"`

	APIUrl string `config:"nuon_api_url"`
	// InternalAPIUrl is the ctl-api address this process dials. The browser still
	// receives APIUrl, which has to be a hostname it can resolve. When this is
	// empty, server calls use APIUrl.
	InternalAPIUrl string `config:"nuon_api_internal_url"`
	// RunnerAPIUrl is the runner API install stacks authenticate against. A different
	// host from APIUrl, and the audience a stack's OIDC trust policy names.
	RunnerAPIUrl   string `config:"nuon_runner_api_url"`
	AdminAPIUrl    string `config:"nuon_admin_api_url"`
	TemporalUIUrl  string `config:"nuon_temporal_ui_url"`
	KafkaUIUrl     string `config:"nuon_kafka_ui_url"`
	GrafanaUIUrl   string `config:"nuon_grafana_ui_url"`
	AuthServiceUrl string `config:"nuon_auth_service_url"`
	AppUrl         string `config:"nuon_app_url"`
	GithubAppName  string `config:"github_app_name"`
	// OAuth client ID of the GitHub App. When set, /connect sends the user
	// through GitHub's user-authorization flow so ctl-api can confirm they own
	// the installation they report.
	GithubAppClientID        string `config:"github_app_client_id"`
	PylonAppID               string `config:"pylon_app_id"`
	DatadogEnv               string `config:"datadog_env"`
	DatadogAPIKey            string `config:"datadog_api_key"`
	DatadogApplicationKey    string `config:"datadog_application_key"`
	DatadogTraceDebug        bool   `config:"datadog_trace_debug"`
	DatadogAPIUrl            string `config:"datadog_api_url"`
	DisableMetrics           bool   `config:"disable_metrics"`
	ServiceDeployment        string `config:"service_deployment"`
	CORSAllowLoopbackOrigins bool   `config:"cors_allow_loopback_origins"`
	IsBYOC                   bool   `config:"nuon_byoc"`
	BYOCName                 string `config:"nuon_byoc_name"`
	BYOCIconText             string `config:"nuon_byoc_icon_text"`
	BYOCColor                string `config:"nuon_byoc_color"`
	OnboardingFirstRun       bool   `config:"nuon_onboarding_first_run"`
	DashboardLite            bool   `config:"nuon_dashboard_lite"`
	StatusBarAutoEnabled     bool   `config:"status_bar_auto_enabled"`
	InstallsTabAutoEnabled   bool   `config:"installs_tab_auto_enabled"`
	AdminDashboardUrl        string `config:"nuon_admin_dashboard_url"`
	PostHogKey               string `config:"posthog_key"`
	PostHogHost              string `config:"posthog_host"`
	PostHogReplayEnabled     bool   `config:"posthog_replay_enabled"`

	MailchimpAPIKey       string `config:"mailchimp_api_key"`
	MailchimpServerPrefix string `config:"mailchimp_server_prefix"`
	MailchimpAudienceID   string `config:"mailchimp_audience_id"`
}

const LocalDevDeployment = "dashboard-local"

func (c *Config) IsLocalDev() bool {
	return c.ServiceDeployment == LocalDevDeployment
}

func (c *Config) ServerAPIURL() string {
	if c.InternalAPIUrl != "" {
		return c.InternalAPIUrl
	}
	return c.APIUrl
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

	if cfg.CORSAllowLoopbackOrigins && !cfg.IsLocalDev() {
		return nil, fmt.Errorf("cors_allow_loopback_origins is only allowed when service_deployment=%s", LocalDevDeployment)
	}

	return &cfg, nil
}
