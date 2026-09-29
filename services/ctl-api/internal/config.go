package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/orgfeatures"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("http_address", "0.0.0.0")

	config.RegisterDefault("http_port", "8081")
	config.RegisterDefault("internal_http_port", "8082")
	config.RegisterDefault("runner_http_port", "8083")
	config.RegisterDefault("auth_http_port", "8084")
	config.RegisterDefault("admin_dashboard_http_port", "8087")
	config.RegisterDefault("slack_http_port", "8089")
	config.RegisterDefault("mcp_http_port", "8088")
	config.RegisterDefault("nuonctl_mcp_http_port", "8091")
	config.RegisterDefault("slack_signing_secret", "insecure-slack-signing-secret-for-dev-only")
	config.RegisterDefault("slack_state_jwt_secret", "insecure-slack-state-jwt-secret-for-dev-only")
	config.RegisterDefault("worker_healthcheck_port", "8086")
	config.RegisterDefault("worker_healthcheck_enabled", true)

	config.RegisterDefault("db_region", "us-west-2")
	config.RegisterDefault("db_port", 5432)
	config.RegisterDefault("db_user", "ctl_api")
	config.RegisterDefault("db_name", "ctl_api")
	config.RegisterDefault("db_max_connections", 12)

	config.RegisterDefault("clickhouse_db_read_timeout", "10s")
	config.RegisterDefault("clickhouse_db_write_timeout", "10s")
	config.RegisterDefault("clickhouse_db_dial_timeout", "1s")

	config.RegisterDefault("kafka_enabled", false)
	config.RegisterDefault("kafka_brokers", "localhost:9092")
	config.RegisterDefault("kafka_security_protocol", "PLAINTEXT")
	config.RegisterDefault("kafka_produce_timeout", "5s")
	// why: kafka_client_id is deliberately not defaulted: it is derived per-process
	// from service_name/service_type/service_deployment unless set explicitly.
	// Group names must keep the ctl-api prefix — the KafkaUser ACL grants group
	// access by prefix, so a name outside it fails authorization at join, which
	// presents as a hang rather than an error because the client retries.
	config.RegisterDefault("kafka_consumer_group_prefix", "ctl-api-consumer")
	config.RegisterDefault("kafka_consumer_fetch_max_wait", "5s")
	config.RegisterDefault("kafka_consumer_fetch_min_bytes", 256*1024)
	config.RegisterDefault("kafka_consumer_fetch_max_bytes", 8*1024*1024)
	config.RegisterDefault("kafka_consumer_fetch_max_partition_bytes", 2*1024*1024)
	config.RegisterDefault("kafka_consumer_max_concurrent_fetches", 2)
	config.RegisterDefault("kafka_consumer_liveness_timeout", "60s")
	config.RegisterDefault("consumer_healthcheck_port", "8090")

	config.RegisterDefault("github_app_key_secret_name", "ctl-api-github-app-key")
	config.RegisterDefault("sandbox_artifacts_base_url", "https://nuon-artifacts.s3.us-west-2.amazonaws.com/sandbox")

	config.RegisterDefault("debug_enable_query_collector", false)
	config.RegisterDefault("query_collector_disabled_tables", "")

	config.RegisterDefault("sandbox_mode_sleep", "5s")
	config.RegisterDefault("sandbox_mode_enable_runners", false)

	// why: runner defaults; per-cloud overrides avoid cross-cloud egress against AWS ECR's pull quota.
	config.RegisterDefault("runner_container_image_url", "public.ecr.aws/p7e3r5y0/runner")
	config.RegisterDefault("runner_container_image_url_gcp", "us-west1-docker.pkg.dev/nuon-public/runner/runner")
	config.RegisterDefault("runner_container_image_url_azure", "")
	config.RegisterDefault("runner_api_url", "http://localhost:8083")
	config.RegisterDefault("public_api_url", "http://localhost:8081")
	config.RegisterDefault("temporal_url", "https://app.nuon.co")
	config.RegisterDefault("telemetry_jwks", "")
	config.RegisterDefault("telemetry_jwt_issuer", "")
	config.RegisterDefault("telemetry_relay_endpoint", "")

	config.RegisterDefault("max_request_size", 1024*50)
	config.RegisterDefault("max_request_duration", time.Second*30)

	config.RegisterDefault("app_repository_name_template", "%s/%s")
	config.RegisterDefault("app_region", "us-west-2")

	config.RegisterDefault("aws_cloudformation_stack_template_bucket_region", "us-east-1")
	config.RegisterDefault("blob_storage_provider", "s3")
	config.RegisterDefault("org_creation_email_allow_list", "nuon.co")
	config.RegisterDefault("server_side_sync_min_cli_version", "0.19.1102")
	config.RegisterDefault("temporal_dataconverter_large_payload_size", 1024*128)
	config.RegisterDefault("large_payload_type", "blob")

	config.RegisterDefault("temporal_blob_s3_prefix", "temporal-blobs/")
	config.RegisterDefault("temporal_blob_cache_dir", "/tmp/temporal-blobs")
	config.RegisterDefault("temporal_blob_cache_max_count", 10000)
	config.RegisterDefault("temporal_blob_cache_max_size_mb", 1024)
	config.RegisterDefault("temporal_blob_s3_timeout", "30s")

	config.RegisterDefault("forced_enabled_features", "")
	config.RegisterDefault("enable_httpbin_debug_endpoints", false)
	config.RegisterDefault("enable_endpoint_auditing", false)
	config.RegisterDefault("org_default_user_journeys_enabled", false)
	config.RegisterDefault("evaluation_journey_enabled", true)
	config.RegisterDefault("webhook_urls", []string{})
	config.RegisterDefault("webhook_timeout", "5s")

	config.RegisterDefault("temporal_workflow_failure_panic", false)
	config.RegisterDefault("temporal_disable_registration_aliasing", false)
	config.RegisterDefault("temporal_sticky_workflow_cache_size", 40000)
	config.RegisterDefault("temporal_sticky_schedule_to_start_timeout", "5s")
	config.RegisterDefault("temporal_deadlock_detection_timeout", "2s")

	config.RegisterDefault("action_crons_enabled", false)

	config.RegisterDefault("queue_handler_grace_period", "1m")

	config.RegisterDefault("queue_idle_timeout", "10m")

	config.RegisterDefault("queue_continue_as_new_hint_period", "1m")

	config.RegisterDefault("queue_drain_timeout", "5m")

	config.RegisterDefault("process_install_uptime_threshold", "8h")
	config.RegisterDefault("process_mng_uptime_threshold", "168h")
	config.RegisterDefault("process_build_uptime_threshold", "8h")

	config.RegisterDefault("general_purge_stale_data_cron", "0 6 * * *")
	config.RegisterDefault("general_purge_stale_data_duration_ago", "168h")
	config.RegisterDefault("queue_signal_cleanup_enabled", true)

	config.RegisterDefault("slack_auto_link_team_id", "")
	config.RegisterDefault("slack_auto_link_channel_id", "")
	config.RegisterDefault("slack_auto_link_org_label_key", "")
	config.RegisterDefault("slack_auto_link_org_label_value", "")

	config.RegisterDefault("internal_email_domains", []string{})

	config.RegisterDefault("posthog_host", "https://us.i.posthog.com")

	config.RegisterDefault("nuon_auth_session_key", "insecure-session-key-for-dev-giqi8x82Ti2+qTQ5ofpazomHkQPSnMY")
	config.RegisterDefault("nuon_auth_allow_all_users", false)
	config.RegisterDefault("nuon_auth_session_ttl", 24*60)
	config.RegisterDefault("nuon_auth_token_ttl", 24*60)
	config.RegisterDefault("nuon_auth_allowed_domains", []string{})

	config.RegisterDefault("oauth_dcr_enabled", true)
	config.RegisterDefault("oauth_access_token_ttl", 60)
	config.RegisterDefault("oauth_refresh_token_ttl", 30*24*60)

	config.RegisterDefault("blob_storage_bucket", "nuon-dev")
	config.RegisterDefault("blob_storage_region", "us-west-2")

	config.RegisterDefault("stale_plan_threshold", "72h")

	config.RegisterDefault("use_legacy_maintenance_role_default", false)
}

type Config struct {
	worker.Config `config:",squash"`

	GitRef                   string   `config:"git_ref" validate:"required"`
	Version                  string   `config:"version" validate:"required"`
	MetricsTags              []string `config:"metrics_tags"`
	DisableMetrics           bool     `config:"disable_metrics"`
	OTELExporterOTLPEndpoint string   `config:"otel_exporter_otlp_endpoint"`
	OTELExporterOTLPProtocol string   `config:"otel_exporter_otlp_protocol"`

	ServiceName       string `config:"service_name" validate:"required"`
	ServiceType       string `config:"service_type" validate:"required"`
	ServiceDeployment string `config:"service_deployment"`

	RootDomain string `config:"root_domain"`

	HTTPPort               string `config:"http_port" validate:"required"`
	InternalHTTPPort       string `config:"internal_http_port" validate:"required"`
	RunnerHTTPPort         string `config:"runner_http_port" validate:"required"`
	AuthHTTPPort           string `config:"auth_http_port" validate:"required"`
	AdminDashboardHTTPPort string `config:"admin_dashboard_http_port" validate:"required"`
	AdminDashboardDistDir  string `config:"admin_dashboard_dist_dir"`
	SlackHTTPPort          string `config:"slack_http_port" validate:"required"`
	MCPHTTPPort            string `config:"mcp_http_port"`
	NuonctlMCPHTTPPort     string `config:"nuonctl_mcp_http_port"`

	WorkerHealthcheckPort    string `config:"worker_healthcheck_port"`
	WorkerHealthcheckEnabled bool   `config:"worker_healthcheck_enabled"`

	GracefulShutdownTimeout time.Duration `config:"graceful_shutdown_timeout" validate:"required"`

	DBName                       string `config:"db_name" validate:"required"`
	DBHost                       string `config:"db_host" validate:"required"`
	DBReplicaHost                string `config:"db_replica_host"`
	DBGormReplicaHost            string `config:"db_gorm_replica_host"`
	DBReplicaEnabled             bool   `config:"db_replica_enabled"`
	DBReplicaBypassOptIn         bool   `config:"db_replica_bypass_opt_in"`
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

	ClickhouseDBName         string        `config:"clickhouse_db_name" validate:"required"`
	ClickhouseDBHost         string        `config:"clickhouse_db_host" validate:"required"`
	ClickhouseDBUser         string        `config:"clickhouse_db_user" validate:"required"`
	ClickhouseDBPassword     string        `config:"clickhouse_db_password" validate:"required"`
	ClickhouseDBPort         string        `config:"clickhouse_db_port" validate:"required"`
	ClickhouseDBUseTLS       bool          `config:"clickhouse_db_use_tls"`
	ClickhouseDBReadTimeout  time.Duration `config:"clickhouse_db_read_timeout" validate:"required"`
	ClickhouseDBWriteTimeout time.Duration `config:"clickhouse_db_write_timeout" validate:"required"`
	ClickhouseDBDialTimeout  time.Duration `config:"clickhouse_db_dial_timeout" validate:"required"`

	KafkaEnabled                        bool          `config:"kafka_enabled"`
	KafkaBrokers                        string        `config:"kafka_brokers"`
	KafkaSecurityProtocol               string        `config:"kafka_security_protocol"`
	KafkaTLSCAPath                      string        `config:"kafka_tls_ca_path"`
	KafkaTLSCertPath                    string        `config:"kafka_tls_cert_path"`
	KafkaTLSKeyPath                     string        `config:"kafka_tls_key_path"`
	KafkaClientID                       string        `config:"kafka_client_id"`
	KafkaProduceTimeout                 time.Duration `config:"kafka_produce_timeout"`
	KafkaConsumerGroupPrefix            string        `config:"kafka_consumer_group_prefix"`
	KafkaConsumerFetchMaxWait           time.Duration `config:"kafka_consumer_fetch_max_wait"`
	KafkaConsumerFetchMinBytes          int32         `config:"kafka_consumer_fetch_min_bytes"`
	KafkaConsumerFetchMaxBytes          int32         `config:"kafka_consumer_fetch_max_bytes"`
	KafkaConsumerFetchMaxPartitionBytes int32         `config:"kafka_consumer_fetch_max_partition_bytes"`
	KafkaConsumerMaxConcurrentFetches   int           `config:"kafka_consumer_max_concurrent_fetches"`
	KafkaConsumerLivenessTimeout        time.Duration `config:"kafka_consumer_liveness_timeout"`
	ConsumerHealthcheckPort             string        `config:"consumer_healthcheck_port"`

	TemporalHost                          string        `config:"temporal_host"  validate:"required"`
	TemporalStickyWorkflowCacheSize       int           `config:"temporal_sticky_workflow_cache_size"`
	TemporalDataConverterLargePayloadSize int           `config:"temporal_dataconverter_large_payload_size"`
	LargePayloadType                      string        `config:"large_payload_type"`
	TemporalBlobS3Prefix                  string        `config:"temporal_blob_s3_prefix"`
	TemporalBlobCacheDir                  string        `config:"temporal_blob_cache_dir"`
	TemporalBlobCacheMaxCount             int           `config:"temporal_blob_cache_max_count"`
	TemporalBlobCacheMaxSizeMB            int           `config:"temporal_blob_cache_max_size_mb"`
	TemporalBlobS3Timeout                 time.Duration `config:"temporal_blob_s3_timeout"`
	TemporalWorkflowFailurePanic          bool          `config:"temporal_workflow_failure_panic"`
	TemporalDisableRegistrationAliasing   bool          `config:"temporal_disable_registration_aliasing"`
	TemporalStickyScheduleToStartTimeout  time.Duration `config:"temporal_sticky_schedule_to_start_timeout"`
	TemporalDeadlockDetectionTimeout      time.Duration `config:"temporal_deadlock_detection_timeout"`

	GithubAppID            string `config:"github_app_id" validate:"required"`
	GithubAppKey           string `config:"github_app_key" validate:"required"`
	GithubAppKeySecretName string `config:"github_app_key_secret_name" validate:"required"`

	SandboxArtifactsBaseURL string `config:"sandbox_artifacts_base_url" validate:"required"`

	Middlewares               []string `config:"middlewares"`
	InternalMiddlewares       []string `config:"internal_middlewares"`
	RunnerMiddlewares         []string `config:"runner_middlewares"`
	AuthMiddlewares           []string `config:"auth_middlewares"`
	AdminDashboardMiddlewares []string `config:"admin_dashboard_middlewares"`
	SlackMiddlewares          []string `config:"slack_middlewares"`

	SlackClientID         string `config:"slack_client_id"`
	SlackClientSecret     string `config:"slack_client_secret"`
	SlackSigningSecret    string `config:"slack_signing_secret"`
	SlackStateJWTSecret   string `config:"slack_state_jwt_secret"`
	SlackOAuthRedirectURL string `config:"slack_oauth_redirect_url"`

	NuonAuthSessionKey     string   `config:"nuon_auth_session_key"`
	NuonAuthSessionTTL     int      `config:"nuon_auth_session_ttl"`
	NuonAuthTokenTTL       int      `config:"nuon_auth_token_ttl"`
	NuonAuthAllowedDomains []string `config:"nuon_auth_allowed_domains"`
	NuonAuthAllowAllUsers  bool     `config:"nuon_auth_allow_all_users"`

	OIDCFederationEnabled              bool `config:"oidc_federation_enabled"`
	OIDCFederationAllowInsecureIssuers bool `config:"oidc_federation_allow_insecure_issuers"`

	OAuthDCREnabled      bool `config:"oauth_dcr_enabled"`
	OAuthAccessTokenTTL  int  `config:"oauth_access_token_ttl"`
	OAuthRefreshTokenTTL int  `config:"oauth_refresh_token_ttl"`

	NuonAuthProviderType string `config:"nuon_auth_provider_type"`
	NuonAuthClientID     string `config:"nuon_auth_client_id"`
	NuonAuthClientSecret string `config:"nuon_auth_client_secret"`
	NuonAuthIssuerURL    string `config:"nuon_auth_issuer_url"`
	NuonAuthRedirectURL  string `config:"nuon_auth_redirect_url"`
	NuonAuthProviderName string `config:"nuon_auth_provider_name"`
	NuonBrandedLogin     bool   `config:"nuon_branded_login"`

	PostHogKey  string `config:"posthog_key"`
	PostHogHost string `config:"posthog_host"`

	AppURL        string `config:"app_url" validate:"required"`
	RunnerAPIURL  string `config:"runner_api_url" validate:"required"`
	PublicAPIURL  string `config:"public_api_url" validate:"required"`
	AdminAPIURL   string `config:"admin_api_url" validate:"required"`
	TemporalUIURL string `config:"temporal_ui_url" validate:"required"`

	ForceSandboxMode           bool          `config:"force_sandbox_mode"`
	ForceOnboardingSandboxMode bool          `config:"force_onboarding_sandbox_mode"`
	SandboxModeSleep           time.Duration `config:"sandbox_mode_sleep" validate:"required"`
	SandboxModeEnableRunners   bool          `config:"sandbox_mode_enable_runners"`

	ForcedEnabledFeatures string `config:"forced_enabled_features"`

	IntegrationGithubInstallID string `config:"integration_github_install_id" validate:"required"`

	LoopsAPIKey string `config:"loops_api_key" validate:"required"`
	// Deprecated: legacy Slack webhook send path is gone; field is read only to populate unused NotificationsConfig rows pending a follow-up cleanup.
	InternalSlackWebhookURL string `config:"internal_slack_webhook_url"`
	DisableNotifications    bool   `config:"disable_notifications"`

	WebhookURLs    []string      `config:"webhook_urls"`
	WebhookTimeout time.Duration `config:"webhook_timeout"`

	AuditOTLPEndpoint string `config:"audit_otlp_endpoint"`
	AuditOTLPToken    string `config:"audit_otlp_token"`

	TelemetryJWKS          string `config:"telemetry_jwks,secure"`
	TelemetryJWTIssuer     string `config:"telemetry_jwt_issuer"`
	TelemetryRelayEndpoint string `config:"telemetry_relay_endpoint"`

	RunnerContainerImageURL      string `config:"runner_container_image_url" validate:"required"`
	RunnerContainerImageURLGCP   string `config:"runner_container_image_url_gcp"`
	RunnerContainerImageURLAzure string `config:"runner_container_image_url_azure"`
	RunnerContainerImageTag      string `config:"runner_container_image_tag" validate:"required"`
	UseLocalRunners              bool   `config:"use_local_runners"`

	AWSIIDCertsDir string `config:"aws_iid_certs_dir"`

	AWSCloudFormationStackTemplateBucketRegion string `config:"aws_cloudformation_stack_template_bucket_region"`
	AWSCloudFormationStackTemplateBucket       string `config:"aws_cloudformation_stack_template_bucket"`
	AWSCloudFormationStackTemplateBaseURL      string `config:"aws_cloudformation_stack_template_base_url"`
	AWSCloudFormationStackTemplateRoleARN      string `config:"aws_cloudformation_stack_template_role_arn"`
	RunnerEnableSupport                        bool   `config:"runner_enable_support"`
	RunnerDefaultSupportIAMRole                string `config:"runner_default_support_iam_role_arn"`

	ManagementAccountID string `config:"management_account_id" validate:"required"`

	ManagementRegion string `config:"management_region"`

	ManagementIAMRoleARN     string `config:"management_iam_role_arn"`
	ManagementECRRegistryID  string `config:"management_ecr_registry_id"`
	ManagementECRRegistryARN string `config:"management_ecr_registry_arn"`

	AWSPhoneHomeCMKARN         string `config:"aws_phone_home_cmk_arn"`
	AWSPhoneHomeSecretsRoleARN string `config:"aws_phone_home_secrets_role_arn"`
	PhoneHomeScriptURL         string `config:"phone_home_script_url"`

	ManagementGARRepositoryURL string `config:"management_gar_repository_url"`

	ManagementACRRegistryURL      string `config:"management_acr_registry_url"`
	ManagementAzureTenantID       string `config:"management_azure_tenant_id"`
	ManagementAzureClientID       string `config:"management_azure_client_id"`
	ManagementAzureSubscriptionID string `config:"management_azure_subscription_id"`
	ManagementAzureResourceGroup  string `config:"management_azure_resource_group"`
	ManagementAzureOIDCIssuerURL  string `config:"management_azure_oidc_issuer_url"`

	AppRegion string `config:"app_region" validate:"required"`

	DNSManagementIAMRoleARN string `config:"dns_management_iam_role_arn"`
	DNSZoneID               string `config:"dns_zone_id" validate:"required"`
	DNSRootDomain           string `config:"dns_root_domain" validate:"required"`

	SegmentWriteKey  string `config:"segment_write_key" validate:"required"`
	DisableAnalytics bool   `config:"disable_analytics"`

	MaxRequestSize     int64         `config:"max_request_size" validate:"required"`
	MaxRequestDuration time.Duration `config:"max_request_duration" validate:"required"`

	ForceDebugMode              bool `config:"force_debug_mode"`
	LogRequestBody              bool `config:"log_request_body"`
	EnableHttpBinDebugEndpoints bool `config:"enable_httpbin_debug_endpoints"`
	EnableEndpointAuditing      bool `config:"enable_endpoint_auditing"`
	EvaluationJourneyEnabled    bool `config:"evaluation_journey_enabled"`

	ChaosRate   int           `config:"chaos_rate"`
	ChaosErrors []string      `config:"chaos_errors"`
	ChaosRoutes []string      `config:"chaos_routes"`
	ChaosSleep  time.Duration `config:"chaos_sleep"`

	ProcessInstallUptimeThreshold time.Duration `config:"process_install_uptime_threshold"`
	ProcessMngUptimeThreshold     time.Duration `config:"process_mng_uptime_threshold"`

	QueueHandlerGracePeriod time.Duration `config:"queue_handler_grace_period"`

	QueueIdleTimeout time.Duration `config:"queue_idle_timeout"`

	QueueContinueAsNewHintPeriod time.Duration `config:"queue_continue_as_new_hint_period"`

	QueueContinueAsNewHistoryMax int `config:"queue_continue_as_new_history_max"`

	QueueDrainTimeout time.Duration `config:"queue_drain_timeout"`

	ActionCronsEnabled bool `config:"action_crons_enabled"`

	MinCLIVersion string `config:"min_cli_version"`

	ServerSideSyncMinCLIVersion string `config:"server_side_sync_min_cli_version"`

	GeneralPurgeStaleDataCron        string        `config:"general_purge_stale_data_cron"`
	GeneralPurgeStaleDataDurationAgo time.Duration `config:"general_purge_stale_data_duration_ago" validate:"required"`

	QueueSignalCleanupEnabled bool `config:"queue_signal_cleanup_enabled"`

	BlobBackfillRatePerSecond int `config:"blob_backfill_rate_per_second"`

	BlobReadEnabled bool `config:"blob_read_enabled"`

	SlackAutoLinkTeamID        string `config:"slack_auto_link_team_id"`
	SlackAutoLinkChannelID     string `config:"slack_auto_link_channel_id"`
	SlackAutoLinkOrgLabelKey   string `config:"slack_auto_link_org_label_key"`
	SlackAutoLinkOrgLabelValue string `config:"slack_auto_link_org_label_value"`

	InternalEmailDomains []string `config:"internal_email_domains"`

	SFTrialEndpoint string `config:"sf_trial_access_endpoint"`

	BlobStorageBucket   string `config:"blob_storage_bucket" validate:"required"`
	BlobStorageRegion   string `config:"blob_storage_region" validate:"required"`
	BlobStorageProvider string `config:"blob_storage_provider" validate:"required,oneof=s3 gcs"`

	EnqueuerMaxWorkers int `config:"enqueuer_max_workers"`

	HeartbeaterFlushInterval time.Duration `config:"heartbeater_flush_interval"`
	HeartbeaterBatchSize     int           `config:"heartbeater_batch_size"`

	DisableEmitterSignals bool `config:"disable_emitter_signals"`

	StalePlanThreshold string `config:"stale_plan_threshold"`

	UseLegacyMaintenanceRoleDefault bool `config:"use_legacy_maintenance_role_default"`
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

func (c *Config) CFTemplateUploadCreds() *credentials.Config {
	if c.IsGCP() && c.AWSCloudFormationStackTemplateRoleARN != "" {
		return &credentials.Config{
			Region: c.AWSCloudFormationStackTemplateBucketRegion,
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:                c.AWSCloudFormationStackTemplateRoleARN,
				SessionName:            "ctl-api-install-templates",
				SessionDurationSeconds: 30 * 60,
				UseGCPOIDC:             true,
			},
		}
	}
	return &credentials.Config{
		Region:     c.AWSCloudFormationStackTemplateBucketRegion,
		UseDefault: true,
	}
}

// why: ManagementSecretsCreds returns credentials for the Nuon AWS account holding the
// phone-home secret, or nil when this control plane has no path to it.
//
// The secret is always in AWS; only the chain differs. An AWS-hosted control plane
// assumes the management role directly, the same one ECR management uses. A
// GCP-hosted one takes one extra step, exchanging the pod's GCP identity for the
// role via sts:AssumeRoleWithWebIdentity. Azure has no federation path to AWS —
// AssumeRoleConfig offers UseGithubOIDC and UseGCPOIDC only — so it returns nil and
// callers skip.
//
// Note this reports how *Nuon* authenticates, not what cloud the install is on.
// Gating phone-home auth on Config.IsAWS instead would silently disable it for every
// AWS install running under a GCP-hosted control plane.
func (c *Config) ManagementSecretsCreds() *credentials.Config {
	roleARN := c.AWSPhoneHomeSecretsRoleARN
	if roleARN == "" && c.IsAWS() {
		roleARN = c.ManagementIAMRoleARN
	}
	if roleARN == "" {
		return nil
	}

	return &credentials.Config{
		Region: c.ManagementRegion,
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:                roleARN,
			SessionName:            "ctl-api-phone-home-secrets",
			SessionDurationSeconds: 60 * 60,
			UseGCPOIDC:             c.IsGCP(),
		},
	}
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

	orgfeatures.SetForced(cfg.ForcedEnabledFeatures)

	switch {
	case cfg.IsGCP():
		if cfg.ManagementGARRepositoryURL == "" {
			return nil, fmt.Errorf("management_gar_repository_url is required when cloud_provider=gcp")
		}
	case cfg.IsAzure():
		if cfg.ManagementACRRegistryURL == "" {
			return nil, fmt.Errorf("management_acr_registry_url is required when cloud_provider=azure")
		}
	default:
		if cfg.ManagementIAMRoleARN == "" {
			return nil, fmt.Errorf("management_iam_role_arn is required when cloud_provider=aws")
		}
		if cfg.ManagementECRRegistryID == "" {
			return nil, fmt.Errorf("management_ecr_registry_id is required when cloud_provider=aws")
		}
	}

	return &cfg, nil
}
