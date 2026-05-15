package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/pkg/workflows/worker"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_address", "0.0.0.0")

	// ports
	config.RegisterDefault("http_port", "8081")
	config.RegisterDefault("internal_http_port", "8082")
	config.RegisterDefault("runner_http_port", "8083")
	config.RegisterDefault("auth_http_port", "8084")
	config.RegisterDefault("admin_dashboard_http_port", "8087")
	config.RegisterDefault("slack_http_port", "8089")
	// Slack secrets: dev-only insecure defaults so the slack-libs FX module
	// (statejwt.New) and signing.Middleware construction don't fail boot
	// when no SLACK_* env is set. Prod overrides via env. Same pattern as
	// nuon_auth_session_key. Other slack_* keys (client_id, client_secret,
	// oauth_redirect_url) are only consumed inside handlers, so an unset
	// value there fails the OAuth request — not boot.
	config.RegisterDefault("slack_signing_secret", "insecure-slack-signing-secret-for-dev-only")
	config.RegisterDefault("slack_state_jwt_secret", "insecure-slack-state-jwt-secret-for-dev-only")
	config.RegisterDefault("worker_healthcheck_port", "8086")
	config.RegisterDefault("worker_healthcheck_enabled", true)

	// defaults for psql database
	config.RegisterDefault("db_region", "us-west-2")
	config.RegisterDefault("db_port", 5432)
	config.RegisterDefault("db_user", "ctl_api")
	config.RegisterDefault("db_name", "ctl_api")
	config.RegisterDefault("db_max_connections", 12)

	// defaults for clickhouse database
	config.RegisterDefault("clickhouse_db_read_timeout", "10s")
	config.RegisterDefault("clickhouse_db_write_timeout", "10s")
	config.RegisterDefault("clickhouse_db_dial_timeout", "1s")

	// defaults for app
	config.RegisterDefault("github_app_key_secret_name", "ctl-api-github-app-key")
	config.RegisterDefault("sandbox_artifacts_base_url", "https://nuon-artifacts.s3.us-west-2.amazonaws.com/sandbox")

	// debug options
	config.RegisterDefault("debug_enable_query_collector", false)
	config.RegisterDefault("query_collector_disabled_tables", "")

	// defaults for sandbox mode
	config.RegisterDefault("sandbox_mode_sleep", "5s")
	// if sandbox_enable_runners is set to true, all jobs require that you process them via a runner, which means
	// running an org runner during seeding and then install runners, etc.
	config.RegisterDefault("sandbox_mode_enable_runners", false)

	// runner defaults; per-cloud overrides avoid cross-cloud egress against AWS ECR's pull quota.
	config.RegisterDefault("runner_container_image_url", "public.ecr.aws/p7e3r5y0/runner")
	config.RegisterDefault("runner_container_image_url_gcp", "us-west1-docker.pkg.dev/nuon-public/runner/runner")
	config.RegisterDefault("runner_container_image_url_azure", "")
	config.RegisterDefault("runner_api_url", "http://localhost:8083")
	config.RegisterDefault("public_api_url", "http://localhost:8081")
	config.RegisterDefault("temporal_url", "https://app.nuon.co")

	// max request sizes to prevent too large of requests
	config.RegisterDefault("max_request_size", 1024*50)
	config.RegisterDefault("max_request_duration", time.Second*30)

	config.RegisterDefault("app_repository_name_template", "%s/%s")
	config.RegisterDefault("app_region", "us-west-2")

	config.RegisterDefault("org_runner_helm_chart_dir", "/bundle/helm")
	config.RegisterDefault("org_runner_instance_type", "t3a.medium")

	config.RegisterDefault("aws_cloudformation_stack_template_bucket_region", "us-east-1")
	config.RegisterDefault("gcp_stack_template_bucket", "nuon-install-templates-gcp")
	config.RegisterDefault("gcp_stack_template_base_url", "https://storage.googleapis.com/nuon-install-templates-gcp")
	config.RegisterDefault("org_creation_email_allow_list", "nuon.co")
	config.RegisterDefault("temporal_dataconverter_large_payload_size", 1024*128)

	config.RegisterDefault("enable_httpbin_debug_endpoints", false)
	config.RegisterDefault("enable_endpoint_auditing", false)
	config.RegisterDefault("org_default_user_journeys_enabled", false)
	config.RegisterDefault("evaluation_journey_enabled", true)
	config.RegisterDefault("webhook_urls", []string{})
	config.RegisterDefault("webhook_timeout", "5s")

	config.RegisterDefault("temporal_workflow_failure_panic", false)
	config.RegisterDefault("temporal_disable_registration_aliasing", false)
	config.RegisterDefault("temporal_sticky_workflow_cache_size", 40000)

	config.RegisterDefault("action_crons_enabled", false)

	// queue handler grace period: how long a finished handler stays alive before auto-terminating
	// short for local dev; prod overrides via config
	config.RegisterDefault("queue_handler_grace_period", "1m")

	// queue idle timeout: how long a queue workflow stays alive with no activity before terminating
	// short for local dev; prod overrides via config
	config.RegisterDefault("queue_idle_timeout", "10m")

	// queue continue-as-new hint period: how often the CAN listener checks for restart hints
	config.RegisterDefault("queue_continue_as_new_hint_period", "1m")

	// runner process uptime thresholds: how long before auto-shutdown
	// defaults are short for local dev; prod overrides via config
	config.RegisterDefault("process_install_uptime_threshold", "8h")
	config.RegisterDefault("process_mng_uptime_threshold", "168h")
	config.RegisterDefault("process_build_uptime_threshold", "8h")

	config.RegisterDefault("event_loop_general_purge_stale_data_cron", "0 6 * * *")
	config.RegisterDefault("event_loop_general_purge_stale_data_duration_ago", "168h")

	// Slack auto-link: empty TeamID or empty OrgLabelKey disables the feature.
	config.RegisterDefault("slack_auto_link_team_id", "")
	config.RegisterDefault("slack_auto_link_channel_id", "")
	config.RegisterDefault("slack_auto_link_org_label_key", "")
	config.RegisterDefault("slack_auto_link_org_label_value", "")

	config.RegisterDefault("internal_email_domains", []string{})

	// Nuon Auth Service Configs
	config.RegisterDefault("nuon_auth_session_key", "insecure-session-key-for-dev-giqi8x82Ti2+qTQ5ofpazomHkQPSnMY")
	config.RegisterDefault("nuon_auth_allow_all_users", false)
	config.RegisterDefault("nuon_auth_session_ttl", 24*60)
	config.RegisterDefault("nuon_auth_token_ttl", 24*60)
	config.RegisterDefault("nuon_auth_allowed_domains", []string{}) // defaults to an empty list so the empty string doesn't raise errors

	// Blob storage configuration
	config.RegisterDefault("blob_storage_bucket", "nuon-dev")
	config.RegisterDefault("blob_storage_region", "us-west-2")

	// Flow check thresholds
	config.RegisterDefault("stale_plan_threshold", "72h") // override with STALE_PLAN_THRESHOLD env var
}

type Config struct {
	worker.Config `config:",squash"`

	// configs for starting and introspecting service
	GitRef         string   `config:"git_ref" validate:"required"`
	Version        string   `config:"version" validate:"required"`
	MetricsTags    []string `config:"metrics_tags"`
	DisableMetrics bool     `config:"disable_metrics"`

	ServiceName       string `config:"service_name" validate:"required"`
	ServiceType       string `config:"service_type" validate:"required"`
	ServiceDeployment string `config:"service_deployment"`

	RootDomain string `config:"root_domain"` // for all services

	HTTPPort               string `config:"http_port" validate:"required"`
	InternalHTTPPort       string `config:"internal_http_port" validate:"required"`
	RunnerHTTPPort         string `config:"runner_http_port" validate:"required"`
	AuthHTTPPort           string `config:"auth_http_port" validate:"required"`
	AdminDashboardHTTPPort string `config:"admin_dashboard_http_port" validate:"required"`
	AdminDashboardDistDir  string `config:"admin_dashboard_dist_dir"`
	SlackHTTPPort          string `config:"slack_http_port" validate:"required"`

	WorkerHealthcheckPort    string `config:"worker_healthcheck_port"`
	WorkerHealthcheckEnabled bool   `config:"worker_healthcheck_enabled"`

	GracefulShutdownTimeout time.Duration `config:"graceful_shutdown_timeout" validate:"required"`

	// psql connection parameters
	DBName                       string `config:"db_name" validate:"required"`
	DBHost                       string `config:"db_host" validate:"required"`
	DBReplicaHost                string `config:"db_replica_host"`
	DBPort                       string `config:"db_port" validate:"required"`
	DBSSLMode                    string `config:"db_ssl_mode" validate:"required"`
	DBPassword                   string `config:"db_password"`
	DBUser                       string `config:"db_user" validate:"required"`
	DBZapLog                     bool   `config:"db_use_zap"`
	DBUseIAM                     bool   `config:"db_use_iam"`
	DBRegion                     string `config:"db_region" validate:"required"`
	CloudProvider                string `config:"cloud_provider"`
	DBLogQueries                 bool   `config:"db_log_queries"`
	DebugEnableQueryCollector    bool   `config:"debug_enable_query_collector"`
	QueryCollectorDisabledTables string `config:"query_collector_disabled_tables"`
	DBMaxConnections             int32  `config:"db_max_connections"`

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
	TemporalDisableRegistrationAliasing   bool   `config:"temporal_disable_registration_aliasing"`

	// github configuration
	GithubAppID            string `config:"github_app_id" validate:"required"`
	GithubAppKey           string `config:"github_app_key" validate:"required"`
	GithubAppKeySecretName string `config:"github_app_key_secret_name" validate:"required"`

	// base urls for filling in various fields on objects
	SandboxArtifactsBaseURL string `config:"sandbox_artifacts_base_url" validate:"required"`

	// middleware configuration
	Middlewares               []string `config:"middlewares"`
	InternalMiddlewares       []string `config:"internal_middlewares"`
	RunnerMiddlewares         []string `config:"runner_middlewares"`
	AuthMiddlewares           []string `config:"auth_middlewares"`
	AdminDashboardMiddlewares []string `config:"admin_dashboard_middlewares"`
	SlackMiddlewares          []string `config:"slack_middlewares"`

	// Slack app configuration (Phase 0–4 of slackbot integration).
	// Tokens are stored plaintext in DB; these env-var-driven values back
	// OAuth + signed-webhook verification at the listener layer.
	SlackClientID         string `config:"slack_client_id"`
	SlackClientSecret     string `config:"slack_client_secret"`
	SlackSigningSecret    string `config:"slack_signing_secret"`
	SlackStateJWTSecret   string `config:"slack_state_jwt_secret"`
	SlackOAuthRedirectURL string `config:"slack_oauth_redirect_url"`

	// Nuon Auth Config
	NuonAuthSessionKey     string   `config:"nuon_auth_session_key"`
	NuonAuthSessionTTL     int      `config:"nuon_auth_session_ttl"`
	NuonAuthTokenTTL       int      `config:"nuon_auth_token_ttl"`
	NuonAuthAllowedDomains []string `config:"nuon_auth_allowed_domains"` // domains from which emails can register
	NuonAuthAllowAllUsers  bool     `config:"nuon_auth_allow_all_users"` // if true, any user with an allowedDomain can sign in

	// Nuon Auth: Default Provider ConfigS
	NuonAuthProviderType string `config:"nuon_auth_provider_type"` // NOTE: becomes required after auth is in GA
	NuonAuthClientID     string `config:"nuon_auth_client_id"`
	NuonAuthClientSecret string `config:"nuon_auth_client_secret"`
	NuonAuthIssuerURL    string `config:"nuon_auth_issuer_url"`
	NuonAuthRedirectURL  string `config:"nuon_auth_redirect_url"`

	// links
	AppURL        string `config:"app_url" validate:"required"`
	RunnerAPIURL  string `config:"runner_api_url" validate:"required"`
	PublicAPIURL  string `config:"public_api_url" validate:"required"`
	AdminAPIURL   string `config:"admin_api_url" validate:"required"`
	TemporalUIURL string `config:"temporal_ui_url" validate:"required"`

	// flags for controlling the background workers
	ForceSandboxMode           bool          `config:"force_sandbox_mode"`
	ForceOnboardingSandboxMode bool          `config:"force_onboarding_sandbox_mode"`
	SandboxModeSleep           time.Duration `config:"sandbox_mode_sleep" validate:"required"`
	SandboxModeEnableRunners   bool          `config:"sandbox_mode_enable_runners"`

	// flags for controlling creation of integration users
	IntegrationGithubInstallID string `config:"integration_github_install_id" validate:"required"`

	// notifications configuration
	LoopsAPIKey             string `config:"loops_api_key" validate:"required"`
	InternalSlackWebhookURL string `config:"internal_slack_webhook_url" validate:"required"`
	DisableNotifications    bool   `config:"disable_notifications"`

	// webhook configuration
	WebhookURLs    []string      `config:"webhook_urls"`
	WebhookTimeout time.Duration `config:"webhook_timeout"`

	// configuration for runners
	RunnerContainerImageURL      string `config:"runner_container_image_url" validate:"required"`
	RunnerContainerImageURLGCP   string `config:"runner_container_image_url_gcp"`
	RunnerContainerImageURLAzure string `config:"runner_container_image_url_azure"`
	RunnerContainerImageTag      string `config:"runner_container_image_tag" validate:"required"`
	UseLocalRunners              bool   `config:"use_local_runners"`

	// AWS IID auth
	AWSIIDCertsDir string `config:"aws_iid_certs_dir"`

	// cloudformation phone home
	AWSCloudFormationStackTemplateBucketRegion string `config:"aws_cloudformation_stack_template_bucket_region"`
	AWSCloudFormationStackTemplateBucket       string `config:"aws_cloudformation_stack_template_bucket"`
	AWSCloudFormationStackTemplateBaseURL      string `config:"aws_cloudformation_stack_template_base_url"`
	RunnerEnableSupport                        bool   `config:"runner_enable_support"`
	RunnerDefaultSupportIAMRole                string `config:"runner_default_support_iam_role_arn"`

	// configuration for managing cloud infra for orgs, apps and installs
	ManagementAccountID string `config:"management_account_id" validate:"required"`

	// AWS management (not required for GCP)
	ManagementIAMRoleARN     string `config:"management_iam_role_arn"`
	ManagementECRRegistryID  string `config:"management_ecr_registry_id"`
	ManagementECRRegistryARN string `config:"management_ecr_registry_arn"`

	// GCP management (not required for AWS)
	ManagementGARRepositoryURL string `config:"management_gar_repository_url"`
	GCSInstallTemplateBucket   string `config:"gcs_install_template_bucket"`

	// Azure management (not required for AWS/GCP)
	ManagementACRRegistryURL      string `config:"management_acr_registry_url"`
	ManagementAzureTenantID       string `config:"management_azure_tenant_id"`
	ManagementAzureClientID       string `config:"management_azure_client_id"`
	ManagementAzureSubscriptionID string `config:"management_azure_subscription_id"`
	ManagementAzureResourceGroup  string `config:"management_azure_resource_group"`
	ManagementAzureOIDCIssuerURL  string `config:"management_azure_oidc_issuer_url"`

	// configuration for org runners (shared across cloud providers)
	OrgRunnerK8sClusterID      string `config:"org_runner_k8s_cluster_id" validate:"required"`
	OrgRunnerK8sPublicEndpoint string `config:"org_runner_k8s_public_endpoint" validate:"required"`
	OrgRunnerK8sCAData         string `config:"org_runner_k8s_ca_data" validate:"required"`
	OrgRunnerRegion            string `config:"org_runner_region" validate:"required"`
	OrgRunnerHelmChartDir      string `config:"org_runner_helm_chart_dir" validate:"required"`
	OrgRunnerInstanceType      string `config:"org_runner_instance_type" validate:"required"`

	// configuration for org runners (AWS-only, not required for GCP)
	OrgRunnerOIDCProviderURL    string `config:"org_runner_oidc_provider_url"`
	OrgRunnerOIDCProviderARN    string `config:"org_runner_oidc_provider_arn"`
	OrgRunnerSupportRoleARN     string `config:"org_runner_support_role_arn"`
	OrgRunnerK8sIAMRoleARN      string `config:"org_runner_k8s_iam_role_arn"`
	OrgRunnerK8sUseDefaultCreds bool   `config:"org_runner_k8s_use_default_creds"`

	// configuration for apps
	AppRegion string `config:"app_region" validate:"required"`

	// configuration for managing the public dns zone
	DNSManagementIAMRoleARN string `config:"dns_management_iam_role_arn"` // AWS-only
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

	// Runner process uptime thresholds
	ProcessInstallUptimeThreshold time.Duration `config:"process_install_uptime_threshold"`
	ProcessMngUptimeThreshold     time.Duration `config:"process_mng_uptime_threshold"`
	ProcessBuildUptimeThreshold   time.Duration `config:"process_build_uptime_threshold"`

	// Queue handler grace period
	QueueHandlerGracePeriod time.Duration `config:"queue_handler_grace_period"`

	// Queue idle timeout: how long before an idle queue workflow terminates
	QueueIdleTimeout time.Duration `config:"queue_idle_timeout"`

	// Queue continue-as-new hint period: how often the CAN listener checks for restart hints
	QueueContinueAsNewHintPeriod time.Duration `config:"queue_continue_as_new_hint_period"`

	// Queue continue-as-new history max: trigger CAN when workflow history exceeds this length
	QueueContinueAsNewHistoryMax int `config:"queue_continue_as_new_history_max"`

	// Action crons
	ActionCronsEnabled bool `config:"action_crons_enabled"`

	MinCLIVersion string `config:"min_cli_version"`

	EventLoopGeneralPurgeStaleDataCron        string        `config:"event_loop_general_purge_stale_data_cron"`
	EventLoopGeneralPurgeStaleDataDurationAgo time.Duration `config:"event_loop_general_purge_stale_data_duration_ago" validate:"required"`

	// Slack auto-link reconciler. TeamID + OrgLabelKey must both be set;
	// ChannelID is optional and seeds a default org-wide subscription per link.
	SlackAutoLinkTeamID        string `config:"slack_auto_link_team_id"`
	SlackAutoLinkChannelID     string `config:"slack_auto_link_channel_id"`
	SlackAutoLinkOrgLabelKey   string `config:"slack_auto_link_org_label_key"`
	SlackAutoLinkOrgLabelValue string `config:"slack_auto_link_org_label_value"`

	// InternalEmailDomains: creator emails matching these skip the default
	// slack-auto-link label seeding in CreateOrg.
	InternalEmailDomains []string `config:"internal_email_domains"`

	// Blob storage configuration
	BlobStorageBucket string `config:"blob_storage_bucket" validate:"required"`
	BlobStorageRegion string `config:"blob_storage_region" validate:"required"`

	// Flow check thresholds
	StalePlanThreshold string `config:"stale_plan_threshold"`
}

func (c *Config) IsAWS() bool {
	return c.CloudProvider != "gcp" && c.CloudProvider != "azure"
}

func (c *Config) IsGCP() bool {
	return c.CloudProvider == "gcp"
}

func (c *Config) IsAzure() bool {
	return c.CloudProvider == "azure"
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
