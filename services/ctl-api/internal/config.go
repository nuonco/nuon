package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"

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

	config.RegisterDefault("event_loop_general_purge_stale_data_cron", "0 6 * * *")
	config.RegisterDefault("event_loop_general_purge_stale_data_duration_ago", "168h")

	// Nuon Auth Service Configs
	config.RegisterDefault("nuon_auth_session_key", "insecure-session-key-for-dev-giqi8x82Ti2+qTQ5ofpazomHkQPSnMY")
	config.RegisterDefault("nuon_auth_allow_all_users", false)
	config.RegisterDefault("nuon_auth_session_ttl", 24*60)
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

	HTTPPort         string `config:"http_port" validate:"required"`
	InternalHTTPPort string `config:"internal_http_port" validate:"required"`
	RunnerHTTPPort   string `config:"runner_http_port" validate:"required"`
	AuthHTTPPort     string `config:"auth_http_port" validate:"required"`

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
	AuthMiddlewares     []string `config:"auth_middlewares"`

	// Nuon Auth Config
	NuonAuthSessionKey     string   `config:"nuon_auth_session_key"`
	NuonAuthSessionTTL     int      `config:"nuon_auth_session_ttl"`
	NuonAuthAllowedDomains []string `config:"nuon_auth_allowed_domains"` // domains from which emails can register
	NuonAuthAllowAllUsers  bool     `config:"nuon_auth_allow_all_users"` // if true, any user with an allowedDomain can sign in

	// Nuon Auth: Default Provider ConfigS
	NuonAuthProviderType string `config:"nuon_auth_provider_type"` // NOTE: becomes required after auth is in GA
	NuonAuthClientID     string `config:"nuon_auth_client_id"`
	NuonAuthClientSecret string `config:"nuon_auth_client_secret"`
	NuonAuthIssuerURL    string `config:"nuon_auth_issuer_url"`
	NuonAuthRedirectURL  string `config:"nuon_auth_redirect_url"`

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

	EventLoopGeneralPurgeStaleDataCron        string        `config:"event_loop_general_purge_stale_data_cron"`
	EventLoopGeneralPurgeStaleDataDurationAgo time.Duration `config:"event_loop_general_purge_stale_data_duration_ago" validate:"required"`
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

// serviceYAML represents the structure of the service.yml file
type serviceYAML struct {
	Env map[string]interface{} `yaml:"env"`
}

// NewTestConfig returns a configuration suitable for integration tests.
// It loads sensible defaults from the config system, then merges values from
// ~/nuonco/mono/services/ctl-api/service.yml if available.
// Environment variables take precedence over service.yml values.
func NewTestConfig() (*Config, error) {
	// Try to load and merge from service.yml in mono repo
	homeDir, err := os.UserHomeDir()
	if err == nil {
		serviceYMLPath := filepath.Join(homeDir, "nuonco", "mono", "services", "ctl-api", "service.yml")
		if err := loadServiceYAMLEnv(serviceYMLPath); err != nil {
			// Non-fatal: log but continue with defaults
			fmt.Printf("Note: Could not load %s: %v\n", serviceYMLPath, err)
		} else {
			fmt.Printf("Loaded test config from %s\n", serviceYMLPath)
		}
	}

	// Now load config normally - this will use:
	// 1. Registered defaults (from init())
	// 2. service.yml values (loaded above as env vars)
	// 3. Environment variables (highest priority)
	var cfg Config
	if err := config.LoadInto(nil, &cfg); err != nil {
		return nil, fmt.Errorf("unable to load test config: %w", err)
	}

	// Validate the resulting config
	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate test config: %w", err)
	}

	return &cfg, nil
}

// loadServiceYAMLEnv reads the service.yml file and sets environment variables
// from its env: section. Only sets vars that aren't already set.
// Also provides test defaults for values that come from env_secrets.
func loadServiceYAMLEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var svc serviceYAML
	if err := yaml.Unmarshal(data, &svc); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	// Set environment variables from service.yml (only if not already set)
	for key, value := range svc.Env {
		if os.Getenv(key) == "" {
			// Convert value to string
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case int, int32, int64, float32, float64:
				strValue = fmt.Sprintf("%v", v)
			case bool:
				strValue = fmt.Sprintf("%t", v)
			default:
				strValue = fmt.Sprintf("%v", v)
			}
			os.Setenv(key, strValue)
		}
	}

	// Provide test defaults for values that come from env_secrets (AWS Secrets Manager)
	// These won't be in the env: section and need fallback values for testing
	testSecretDefaults := map[string]string{
		"GITHUB_APP_KEY":                     "test-github-app-key",
		"LOOPS_API_KEY":                      "test-loops-api-key",
		"SEGMENT_WRITE_KEY":                  "test-segment-key",
		"AUTH0_CLIENT_ID":                    "test-auth0-client-id",
		"NUON_AUTH_CLIENT_ID":                "test-nuon-auth-client-id",
		"NUON_AUTH_CLIENT_SECRET":            "test-nuon-auth-client-secret",
		"ENV":                                "test",
		"SERVICE_NAME":                       "ctl-api-test",
		"HOST_IP":                            "127.0.0.1",
		"VERSION":                            "test",
		"TEMPORAL_TASK_QUEUE":                "ctl-api",
		"TEMPORAL_HOST":                      "localhost:7233",
		"TEMPORAL_MAX_CONCURRENT_ACTIVITIES": "100",
		"RUNNER_CONTAINER_IMAGE_TAG":         "latest",

		// Management account defaults (not in service.yml)
		"MANAGEMENT_IAM_ROLE_ARN":     "arn:aws:iam::766121324316:role/management-test",
		"MANAGEMENT_ACCOUNT_ID":       "766121324316",
		"MANAGEMENT_ECR_REGISTRY_ID":  "766121324316",
		"MANAGEMENT_ECR_REGISTRY_ARN": "arn:aws:ecr:us-west-2:766121324316:registry/test",

		// Org runner defaults (not in service.yml)
		"ORG_RUNNER_K8S_CLUSTER_ID":      "test-cluster",
		"ORG_RUNNER_K8S_PUBLIC_ENDPOINT": "https://test.eks.amazonaws.com",
		"ORG_RUNNER_K8S_CA_DATA":         "test-ca-data",
		"ORG_RUNNER_OIDC_PROVIDER_URL":   "https://oidc.eks.us-west-2.amazonaws.com/id/test",
		"ORG_RUNNER_OIDC_PROVIDER_ARN":   "arn:aws:iam::766121324316:oidc-provider/test",
		"ORG_RUNNER_REGION":              "us-west-2",
		"ORG_RUNNER_SUPPORT_ROLE_ARN":    "arn:aws:iam::766121324316:role/org-runner-support",
		"ORG_RUNNER_K8S_IAM_ROLE_ARN":    "arn:aws:iam::766121324316:role/org-runner-k8s",

		// DNS defaults (not in service.yml)
		"DNS_MANAGEMENT_IAM_ROLE_ARN": "arn:aws:iam::766121324316:role/dns-management",
		"DNS_ZONE_ID":                 "Z1234567890ABC",
		"DNS_ROOT_DOMAIN":             "test.nuon.co",
	}

	for key, defaultValue := range testSecretDefaults {
		if os.Getenv(key) == "" {
			os.Setenv(key, defaultValue)
		}
	}

	return nil
}
