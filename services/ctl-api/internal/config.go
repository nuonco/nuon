package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/aws/credentials"
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

	// defaults for kafka
	config.RegisterDefault("kafka_enabled", false)
	config.RegisterDefault("kafka_brokers", "localhost:9092")
	config.RegisterDefault("kafka_security_protocol", "PLAINTEXT")
	config.RegisterDefault("kafka_produce_timeout", "5s")
	// kafka_client_id is deliberately not defaulted: it is derived per-process
	// from service_name/service_type/service_deployment unless set explicitly.
	// Group names must keep the ctl-api prefix — the KafkaUser ACL grants group
	// access by prefix, so a name outside it fails authorization at join, which
	// presents as a hang rather than an error because the client retries.
	config.RegisterDefault("kafka_consumer_group_prefix", "ctl-api-consumer")
	// consumer flush cadence: each fetch (and so each ClickHouse insert) waits
	// until min_bytes accumulate or max_wait elapses, whichever comes first.
	// min_bytes is the target batch/part size; max_wait caps latency and, at low
	// volume, bounds the insert rate to ~1 per partition per interval.
	config.RegisterDefault("kafka_consumer_fetch_max_wait", "5s")
	config.RegisterDefault("kafka_consumer_fetch_min_bytes", 256*1024)
	// Ceilings on a single fetch, capping how much a consumer can buffer before
	// decode. Left unset, franz-go allows 50MiB per broker with unbounded fetch
	// concurrency. See pkg/kafka/consumer.go for the arithmetic; overridden
	// per-deployment for the higher-volume consumers.
	config.RegisterDefault("kafka_consumer_fetch_max_bytes", 8*1024*1024)
	config.RegisterDefault("kafka_consumer_fetch_max_partition_bytes", 2*1024*1024)
	config.RegisterDefault("kafka_consumer_max_concurrent_fetches", 2)
	// A handler call is bounded by clickhouse_db_write_timeout (10s) for the
	// main insert, plus — only when decode() hit a bad record — up to
	// kafka_produce_timeout (5s) trying the dead-letter topic before its own
	// clickhouse_db_write_timeout-bounded fallback. Worst realistic case is
	// ~25s (5s + 10s + 10s); 60s leaves a healthy margin above that without
	// being so loose it stops meaning anything. Real p95 CREATE latency
	// against these tables is well under 1s (see
	// plans/08-kafka-phase5-consumer-hardening.md), so tripping this at all
	// means a genuine stuck handler, not ordinary backend slowness.
	config.RegisterDefault("kafka_consumer_liveness_timeout", "60s")
	// Not 8086 (worker_healthcheck_port) — that server is always on locally
	// too, and consumer + worker run side by side under `nuonctl dev`.
	config.RegisterDefault("consumer_healthcheck_port", "8090")

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
	config.RegisterDefault("blob_storage_provider", "s3")
	config.RegisterDefault("org_creation_email_allow_list", "nuon.co")
	config.RegisterDefault("recommended_cli_version", "0.19.1102")
	config.RegisterDefault("temporal_dataconverter_large_payload_size", 1024*128)
	config.RegisterDefault("large_payload_type", "blob")

	config.RegisterDefault("temporal_blob_s3_prefix", "temporal-blobs/")
	config.RegisterDefault("temporal_blob_cache_dir", "/tmp/temporal-blobs")
	config.RegisterDefault("temporal_blob_cache_max_count", 10000)
	config.RegisterDefault("temporal_blob_cache_max_size_mb", 1024)
	config.RegisterDefault("temporal_blob_s3_timeout", "30s")

	config.RegisterDefault("auto_enabled_features", "")
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

	// queue handler grace period: how long a finished handler stays alive before auto-terminating
	// short for local dev; prod overrides via config
	config.RegisterDefault("queue_handler_grace_period", "1m")

	// queue idle timeout: how long a queue workflow stays alive with no activity before terminating
	// short for local dev; prod overrides via config
	config.RegisterDefault("queue_idle_timeout", "10m")

	// queue continue-as-new hint period: how often the CAN listener checks for restart hints
	config.RegisterDefault("queue_continue_as_new_hint_period", "1m")

	// queue drain timeout: how long to wait for in-flight signals to complete before restarting/stopping
	config.RegisterDefault("queue_drain_timeout", "5m")

	// runner process uptime thresholds: how long before auto-shutdown
	// defaults are short for local dev; prod overrides via config
	config.RegisterDefault("process_install_uptime_threshold", "8h")
	config.RegisterDefault("process_mng_uptime_threshold", "168h")
	config.RegisterDefault("process_build_uptime_threshold", "8h")

	config.RegisterDefault("general_purge_stale_data_cron", "0 6 * * *")
	config.RegisterDefault("general_purge_stale_data_duration_ago", "168h")

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

	// When true, the lowest-precedence operation role is defaulted from the parent
	// install workflow type (provision/reprovision -> provision role, deprovision ->
	// deprovision role; everything else -> maintenance role). When false, deploys and
	// action runs default to the maintenance role. Global rollout switch for all orgs;
	// override with WORKFLOW_DEFAULT_ROLE_ENABLED env var.
	config.RegisterDefault("workflow_default_role_enabled", false)
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

	// kafka configuration. Two security protocols: PLAINTEXT locally, and SSL
	// with the client certificate Strimzi issues our KafkaUser in deployed
	// environments. KafkaEnabled gates whether producers use Kafka vs writing
	// straight to ClickHouse.
	KafkaEnabled          bool   `config:"kafka_enabled"`
	KafkaBrokers          string `config:"kafka_brokers"`
	KafkaSecurityProtocol string `config:"kafka_security_protocol"`
	KafkaTLSCAPath        string `config:"kafka_tls_ca_path"`
	KafkaTLSCertPath      string `config:"kafka_tls_cert_path"`
	KafkaTLSKeyPath       string `config:"kafka_tls_key_path"`
	KafkaClientID         string `config:"kafka_client_id"`
	// Bounds a synchronous produce, used by the log write paths. franz-go retries
	// a buffered record effectively forever, so this is what keeps a broker outage
	// from blocking a runner's log write rather than falling back to ClickHouse.
	KafkaProduceTimeout time.Duration `config:"kafka_produce_timeout"`
	// Each consumer derives its own group as <prefix>-<name>, so one consumer
	// restarting doesn't rebalance the others.
	KafkaConsumerGroupPrefix            string        `config:"kafka_consumer_group_prefix"`
	KafkaConsumerFetchMaxWait           time.Duration `config:"kafka_consumer_fetch_max_wait"`
	KafkaConsumerFetchMinBytes          int32         `config:"kafka_consumer_fetch_min_bytes"`
	KafkaConsumerFetchMaxBytes          int32         `config:"kafka_consumer_fetch_max_bytes"`
	KafkaConsumerFetchMaxPartitionBytes int32         `config:"kafka_consumer_fetch_max_partition_bytes"`
	KafkaConsumerMaxConcurrentFetches   int           `config:"kafka_consumer_max_concurrent_fetches"`
	// Liveness threshold for a stuck handler call — see the healthcheck server
	// in internal/health/consumer.go. Deliberately separate from any consumer's
	// own fetch tuning: it's a backstop against a hang, not a knob to tune.
	KafkaConsumerLivenessTimeout time.Duration `config:"kafka_consumer_liveness_timeout"`
	ConsumerHealthcheckPort      string        `config:"consumer_healthcheck_port"`

	// temporal configuration
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

	// OIDC workload identity federation
	OIDCFederationEnabled              bool `config:"oidc_federation_enabled"`                // enables the /v1/oidc token exchange and trust policy endpoints (default off)
	OIDCFederationAllowInsecureIssuers bool `config:"oidc_federation_allow_insecure_issuers"` // allow http:// issuer URLs in trust policies (local dev only)

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

	AutoEnabledFeatures string `config:"auto_enabled_features"`

	// flags for controlling creation of integration users
	IntegrationGithubInstallID string `config:"integration_github_install_id" validate:"required"`

	// notifications configuration
	LoopsAPIKey string `config:"loops_api_key" validate:"required"`
	// Deprecated: legacy Slack webhook send path is gone; field is read only to populate unused NotificationsConfig rows pending a follow-up cleanup.
	InternalSlackWebhookURL string `config:"internal_slack_webhook_url"`
	DisableNotifications    bool   `config:"disable_notifications"`

	// webhook configuration
	WebhookURLs    []string      `config:"webhook_urls"`
	WebhookTimeout time.Duration `config:"webhook_timeout"`

	// Audit log export. Audit records are the only telemetry ctl-api ships over
	// OTLP; everything else keeps going to stderr untouched. Leave the endpoint
	// empty to disable, which is the default until the gateway collector exists.
	AuditOTLPEndpoint string `config:"audit_otlp_endpoint"`
	AuditOTLPToken    string `config:"audit_otlp_token"`

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
	AWSCloudFormationStackTemplateRoleARN      string `config:"aws_cloudformation_stack_template_role_arn"`
	RunnerEnableSupport                        bool   `config:"runner_enable_support"`
	RunnerDefaultSupportIAMRole                string `config:"runner_default_support_iam_role_arn"`

	// configuration for managing cloud infra for orgs, apps and installs
	ManagementAccountID string `config:"management_account_id" validate:"required"`

	// ManagementRegion is where management-account resources live. Distinct from
	// AppRegion: it is baked into customer-deployed artifacts (the phone-home
	// Lambda's NUON_PHONE_HOME_SECRET_REGION), so it must be stated explicitly
	// rather than inherited from the pod's AWS_REGION.
	ManagementRegion string `config:"management_region"`

	// AWS management (not required for GCP)
	ManagementIAMRoleARN     string `config:"management_iam_role_arn"`
	ManagementECRRegistryID  string `config:"management_ecr_registry_id"`
	ManagementECRRegistryARN string `config:"management_ecr_registry_arn"`

	// phone home auth (see plans/phone-home-auth-shared-secret.md)
	//
	// The phone-home secret always lives in AWS Secrets Manager in a Nuon-owned AWS
	// account, whichever cloud this control plane runs on, because the reader is the
	// customer's phone-home Lambda and Secrets Manager is the only store it can
	// reach. CloudProvider only selects the credential chain — see
	// ManagementSecretsCreds.
	//
	// AWSPhoneHomeCMKARN is the shared CMK encrypting the secret. It is required for
	// cross-account reads: the AWS-managed aws/secretsmanager key policy cannot be
	// edited and is scoped to kms:CallerAccount, so a customer principal fails
	// kms:Decrypt no matter how permissive the secret's resource policy is.
	AWSPhoneHomeCMKARN string `config:"aws_phone_home_cmk_arn"`
	// AWSPhoneHomeSecretsRoleARN is the role in the Nuon AWS account that owns the
	// secret. Set it on GCP-hosted control planes, where ManagementIAMRoleARN is
	// legitimately empty; on AWS it defaults to ManagementIAMRoleARN.
	AWSPhoneHomeSecretsRoleARN string `config:"aws_phone_home_secrets_role_arn"`
	// PhoneHomeScriptURL overrides the phone-home Lambda source for every app in this
	// environment, below a per-app AppRunnerConfig.PhoneHomeScriptURL and above the
	// pinned default. Exists so a dev or staging control plane can exercise an
	// unreleased script without moving the default that production shares.
	PhoneHomeScriptURL string `config:"phone_home_script_url"`

	// GCP management (not required for AWS)
	ManagementGARRepositoryURL string `config:"management_gar_repository_url"`
	// When set, org runners share this stack-created SA (WI bindings appended
	// per org) instead of ctl-api creating one SA per org at runtime.
	ManagementGCPOrgRunnerSAEmail string `config:"management_gcp_org_runner_sa_email"`

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

	// Queue drain timeout: how long to wait for in-flight signals before restarting/stopping
	QueueDrainTimeout time.Duration `config:"queue_drain_timeout"`

	// Action crons
	ActionCronsEnabled bool `config:"action_crons_enabled"`

	MinCLIVersion string `config:"min_cli_version"`

	// below this the CLI syncs app configs client-side, which never sends action or
	// runbook ids, so those are silently dropped from installs
	RecommendedCLIVersion string `config:"recommended_cli_version"`

	GeneralPurgeStaleDataCron        string        `config:"general_purge_stale_data_cron"`
	GeneralPurgeStaleDataDurationAgo time.Duration `config:"general_purge_stale_data_duration_ago" validate:"required"`

	// When enabled, the daily cron hard-deletes process_healthcheck queue signals older than 7 days.
	QueueSignalCleanupEnabled bool `config:"queue_signal_cleanup_enabled"`

	// BlobBackfillRatePerSecond caps how many S3 PUTs/sec the blob backfill activity issues. Defaults to 500 when unset.
	BlobBackfillRatePerSecond int `config:"blob_backfill_rate_per_second"`

	// BlobReadEnabled gates reading large fields from the S3 blob. When false
	// (default) reads fall back to the legacy column. Currently gates composite
	// plan reads.
	BlobReadEnabled bool `config:"blob_read_enabled"`

	// ResourceGrantsEnabled gates object-level resource grants. When false
	// (default) the org gate behaves exactly as before: an account without an
	// org-wide permission is rejected. When true, org membership and org
	// authorization are split so a grant-scoped account defers to a downstream
	// resource check instead of being rejected at the org gate.
	ResourceGrantsEnabled bool `config:"resource_grants_enabled"`

	// Slack auto-link reconciler. TeamID + OrgLabelKey must both be set;
	// ChannelID is optional and seeds a default org-wide subscription per link.
	SlackAutoLinkTeamID        string `config:"slack_auto_link_team_id"`
	SlackAutoLinkChannelID     string `config:"slack_auto_link_channel_id"`
	SlackAutoLinkOrgLabelKey   string `config:"slack_auto_link_org_label_key"`
	SlackAutoLinkOrgLabelValue string `config:"slack_auto_link_org_label_value"`

	// InternalEmailDomains: creator emails matching these skip the default
	// slack-auto-link label seeding in CreateOrg.
	InternalEmailDomains []string `config:"internal_email_domains"`

	// SFTrialEndpoint posts a Salesforce trial-signup record when a user creates
	// their first org. Empty disables the integration (e.g. BYOC/self-hosted).
	SFTrialEndpoint string `config:"sf_trial_access_endpoint"`

	// Blob storage configuration. Provider selects the backend: "s3" (default,
	// AWS-hosted installs) or "gcs" (self-hosted control-plane installs on GCP,
	// where BlobStorageBucket is a native GCS bucket rather than S3).
	BlobStorageBucket   string `config:"blob_storage_bucket" validate:"required"`
	BlobStorageRegion   string `config:"blob_storage_region" validate:"required"`
	BlobStorageProvider string `config:"blob_storage_provider" validate:"required,oneof=s3 gcs"`

	// Enqueuer worker pool size — how many signals can be enqueued in parallel.
	EnqueuerMaxWorkers int `config:"enqueuer_max_workers"`

	// Heartbeater configuration — batched ClickHouse writes for runner heartbeats.
	HeartbeaterFlushInterval time.Duration `config:"heartbeater_flush_interval"`
	HeartbeaterBatchSize     int           `config:"heartbeater_batch_size"`

	// DisableEmitterSignals when true causes all emitter-originated signals to be
	// skipped (not emitted and not processed).
	DisableEmitterSignals bool `config:"disable_emitter_signals"`

	// Flow check thresholds
	StalePlanThreshold string `config:"stale_plan_threshold"`

	// WorkflowDefaultRoleEnabled is a global switch (all orgs) that defaults the
	// lowest-precedence operation role from the parent install workflow type.
	WorkflowDefaultRoleEnabled bool `config:"workflow_default_role_enabled"`
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

// ManagementSecretsCreds returns credentials for the Nuon AWS account holding the
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
