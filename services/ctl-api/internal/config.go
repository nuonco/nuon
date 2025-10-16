package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_address", "0.0.0.0")

	// defaults for psql database
	config.RegisterDefault("db_region", "us-west-2")
	config.RegisterDefault("db_port", 5432)
	config.RegisterDefault("db_user", "ctl_api")
	config.RegisterDefault("db_name", "ctl_api")

	// defaults for clickhouse database
	config.RegisterDefault("clickhouse_db_read_timeout", "10s")
	config.RegisterDefault("clickhouse_db_write_timeout", "10s")
	config.RegisterDefault("clickhouse_db_dial_timeout", "1s")

	// defaults for app
	config.RegisterDefault("github_app_key_secret_name", "ctl-api-github-app-key")
	config.RegisterDefault("sandbox_artifacts_base_url", "https://nuon-artifacts.s3.us-west-2.amazonaws.com/sandbox")

	// defaults for sandbox mode
	config.RegisterDefault("sandbox_mode_sleep", "5s")
	// if sandbox_enable_runners is set to true, all jobs require that you process them via a runner, which means
	// running an org runner during seeding and then install runners, etc.
	config.RegisterDefault("sandbox_mode_enable_runners", false)

	// runner defaults
	config.RegisterDefault("runner_container_image_url", "public.ecr.aws/p7e3r5y0/runner")
	config.RegisterDefault("runner_api_url", "http://localhost:8083")
	config.RegisterDefault("public_api_url", "http://localhost:8081")
	config.RegisterDefault("temporal_url", "https://app.nuon.co")

	// max request sizes to prevent too large of requests
	config.RegisterDefault("max_request_size", 1024*50)
	config.RegisterDefault("max_request_duration", time.Second*30)

	config.RegisterDefault("app_repository_name_template", "%s/%s")
	config.RegisterDefault("app_repository_region", "%s/%s")

	config.RegisterDefault("org_runner_helm_chart_dir", "/bundle/helm")
	config.RegisterDefault("org_runner_instance_type", "t3a.medium")

	config.RegisterDefault("aws_cloudformation_stack_template_bucket_region", "us-east-1")
	config.RegisterDefault("org_creation_email_allow_list", "nuon.co")
	config.RegisterDefault("temporal_dataconverter_large_payload_size", 1024*128)

	config.RegisterDefault("enable_httpbin_debug_endpoints", false)
	config.RegisterDefault("enable_endpoint_auditing", false)
	config.RegisterDefault("org_default_user_journeys_enabled", false)
	config.RegisterDefault("evaluation_journey_enabled", true)

	config.RegisterDefault("temporal_workflow_failure_panic", false)

	config.RegisterDefault("action_crons_enabled", false)
}

type Config struct {
	worker.Config `config:",squash"`

	// configs for starting and introspecting service
	GitRef         string   `config:"git_ref" validate:"required"`
	Version        string   `config:"version" validate:"required"`
	MetricsTags    []string `config:"metrics_tags"`
	DisableMetrics bool     `config:"disable_metrics"`

	ServiceName             string        `config:"service_name" validate:"required"`
	ServiceType             string        `config:"service_type" validate:"required"`
	ServiceDeployment       string        `config:"service_deployment"`
	HTTPPort                string        `config:"http_port" validate:"required"`
	InternalHTTPPort        string        `config:"internal_http_port" validate:"required"`
	RunnerHTTPPort          string        `config:"runner_http_port" validate:"required"`
	GracefulShutdownTimeout time.Duration `config:"graceful_shutdown_timeout" validate:"required"`

	// psql connection parameters
	DBName       string `config:"db_name" validate:"required"`
	DBHost       string `config:"db_host" validate:"required"`
	DBPort       string `config:"db_port" validate:"required"`
	DBSSLMode    string `config:"db_ssl_mode" validate:"required"`
	DBPassword   string `config:"db_password"`
	DBUser       string `config:"db_user" validate:"required"`
	DBZapLog     bool   `config:"db_use_zap"`
	DBUseIAM     bool   `config:"db_use_iam"`
	DBRegion     string `config:"db_region" validate:"required"`
	DBLogQueries bool   `config:"db_log_queries"`

	// clickhouse connection parameters
	ClickhouseDBName         string        `config:"clickhouse_db_name" validate:"required"`
	ClickhouseDBHost         string        `config:"clickhouse_db_host" validate:"required"`
	ClickhouseDBUser         string        `config:"clickhouse_db_user" validate:"required"`
	ClickhouseDBPassword     string        `config:"clickhouse_db_password" validate:"required"`
	ClickhouseDBPort         string        `config:"clickhouse_db_port" validate:"required"`
	ClickhouseDBUseTLS       bool          `config:"clickhouse_db_use_tls"`
	ClickhouseDBReadTimeout  time.Duration `config:"clickhouse_db_read_timeout" validate:"required"`
	ClickhouseDBWriteTimeout time.Duration `config:"clickhouse_db_write_timeout" validate:"required"`
	ClickhouseDBDialTimeout  time.Duration `config:"clickhouse_db_dial_timeout" validate:"required"`

	// temporal configuration
	TemporalHost                          string `config:"temporal_host"  validate:"required"`
	TemporalStickyWorkflowCacheSize       int    `config:"temporal_sticky_workflow_cache_size"`
	TemporalDataConverterLargePayloadSize int    `config:"temporal_dataconverter_large_payload_size"`
	TemporalWorkflowFailurePanic          bool   `config:"temporal_workflow_failure_panic"`

	// github configuration
	GithubAppID            string `config:"github_app_id" validate:"required"`
	GithubAppKey           string `config:"github_app_key" validate:"required"`
	GithubAppKeySecretName string `config:"github_app_key_secret_name" validate:"required"`

	// base urls for filling in various fields on objects
	SandboxArtifactsBaseURL string `config:"sandbox_artifacts_base_url" validate:"required"`

	// middleware configuration
	Middlewares         []string `config:"middlewares"`
	InternalMiddlewares []string `config:"internal_middlewares"`
	RunnerMiddlewares   []string `config:"runner_middlewares"`

	// auth 0 config
	Auth0IssuerURL string `config:"auth0_issuer_url" validate:"required"`
	Auth0Audience  string `config:"auth0_audience" validate:"required"`
	Auth0ClientID  string `config:"auth0_client_id" validate:"required"`

	// links
	AppURL        string `config:"app_url" validate:"required"`
	RunnerAPIURL  string `config:"runner_api_url" validate:"required"`
	PublicAPIURL  string `config:"public_api_url" validate:"required"`
	AdminAPIURL   string `config:"admin_api_url" validate:"required"`
	TemporalUIURL string `config:"temporal_ui_url" validate:"required"`

	// flags for controlling the background workers
	ForceSandboxMode         bool          `config:"force_sandbox_mode"`
	SandboxModeSleep         time.Duration `config:"sandbox_mode_sleep" validate:"required"`
	SandboxModeEnableRunners bool          `config:"sandbox_mode_enable_runners"`

	// flags for controlling creation of integration users
	IntegrationGithubInstallID string `config:"integration_github_install_id" validate:"required"`

	// notifications configuration
	LoopsAPIKey             string `config:"loops_api_key" validate:"required"`
	InternalSlackWebhookURL string `config:"internal_slack_webhook_url" validate:"required"`
	DisableNotifications    bool   `config:"disable_notifications"`

	// configuration for runners
	RunnerContainerImageURL string `config:"runner_container_image_url" validate:"required"`
	RunnerContainerImageTag string `config:"runner_container_image_tag" validate:"required"`
	UseLocalRunners         bool   `config:"use_local_runners"`

	// cloudformation phone home
	AWSCloudFormationStackTemplateBucketRegion string `config:"aws_cloudformation_stack_template_bucket_region"`
	AWSCloudFormationStackTemplateBucket       string `config:"aws_cloudformation_stack_template_bucket"`
	AWSCloudFormationStackTemplateBaseURL      string `config:"aws_cloudformation_stack_template_base_url"`
	RunnerEnableSupport                        bool   `config:"runner_enable_support"`
	RunnerDefaultSupportIAMRole                string `config:"runner_default_support_iam_role_arn" validate:"required"`

	// configuration for managing AWS infra for orgs, apps and installs
	ManagementIAMRoleARN string `config:"management_iam_role_arn" validate:"required"`

	ManagementAccountID      string `config:"management_account_id" validate:"required"`
	ManagementECRRegistryID  string `config:"management_ecr_registry_id" validate:"required"`
	ManagementECRRegistryARN string `config:"management_ecr_registry_arn" validate:"required"`

	// configuration for org runners
	OrgRunnerK8sClusterID       string `config:"org_runner_k8s_cluster_id" validate:"required"`
	OrgRunnerK8sPublicEndpoint  string `config:"org_runner_k8s_public_endpoint" validate:"required"`
	OrgRunnerK8sCAData          string `config:"org_runner_k8s_ca_data" validate:"required"`
	OrgRunnerOIDCProviderURL    string `config:"org_runner_oidc_provider_url" validate:"required"`
	OrgRunnerOIDCProviderARN    string `config:"org_runner_oidc_provider_arn" validate:"required"`
	OrgRunnerRegion             string `config:"org_runner_region" validate:"required"`
	OrgRunnerSupportRoleARN     string `config:"org_runner_support_role_arn" validate:"required"`
	OrgRunnerHelmChartDir       string `config:"org_runner_helm_chart_dir" validate:"required"`
	OrgRunnerK8sIAMRoleARN      string `config:"org_runner_k8s_iam_role_arn" validate:"required"`
	OrgRunnerK8sUseDefaultCreds bool   `config:"org_runner_k8s_use_default_creds"`
	OrgRunnerInstanceType       string `config:"org_runner_instance_type" validate:"required"`

	// configuration for managing the public dns zone
	DNSManagementIAMRoleARN string `config:"dns_management_iam_role_arn" validate:"required"`
	DNSZoneID               string `config:"dns_zone_id" validate:"required"`
	DNSRootDomain           string `config:"dns_root_domain" validate:"required"`

	// analytics configuration
	SegmentWriteKey  string `config:"segment_write_key" validate:"required"`
	DisableAnalytics bool   `config:"disable_analytics"`

	MaxRequestSize     int64         `config:"max_request_size" validate:"required"`
	MaxRequestDuration time.Duration `config:"max_request_duration" validate:"required"`

	// Force debug mode for everything
	ForceDebugMode              bool `config:"force_debug_mode"`
	LogRequestBody              bool `config:"log_request_body"`
	EnableHttpBinDebugEndpoints bool `config:"enable_httpbin_debug_endpoints"`
	EnableEndpointAuditing      bool `config:"enable_endpoint_auditing"`
	EvaluationJourneyEnabled    bool `config:"evaluation_journey_enabled"`

	// chaos configuration
	ChaosRate   int           `config:"chaos_rate"`
	ChaosErrors []string      `config:"chaos_errors"`
	ChaosRoutes []string      `config:"chaos_routes"`
	ChaosSleep  time.Duration `config:"chaos_sleep"`

	// Action crons
	ActionCronsEnabled bool `config:"action_crons_enabled"`

	MinCLIVersion string `config:"min_cli_version"`
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
