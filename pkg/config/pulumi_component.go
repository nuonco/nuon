package config

import (
	"github.com/invopop/jsonschema"
)

// PulumiComponentConfig is the configuration for a Pulumi component.
type PulumiComponentConfig struct {
	Runtime       string `mapstructure:"runtime" toml:"runtime" jsonschema:"required"`
	PulumiVersion string `mapstructure:"pulumi_version,omitempty" toml:"pulumi_version,omitempty"`

	ConfigMap map[string]string `mapstructure:"config,omitempty" toml:"config,omitempty"`
	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty" jsonschema:"oneof_required=public_repo"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" toml:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`

	MaxAutoRetries               *int  `mapstructure:"max_auto_retries,omitempty" toml:"max_auto_retries,omitempty" nuonhash:"omitempty"`
	SkipNoops                    *bool `mapstructure:"skip_noops,omitempty" toml:"skip_noops,omitempty" nuonhash:"omitempty"`
	AutoApproveOnPoliciesPassing *bool `mapstructure:"auto_approve_on_policies_passing,omitempty" toml:"auto_approve_on_policies_passing,omitempty" nuonhash:"omitempty"`
}

func (p PulumiComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("runtime").Short("Pulumi runtime").Required().
		Long("The Pulumi runtime to use for the program (go, nodejs, python)").
		Example("go").
		Example("nodejs").
		Example("python").
		Field("pulumi_version").Short("Pulumi version").
		Long("Version of the Pulumi CLI to use for deployments. If not specified, uses the latest version").
		Example("3.100.0").
		Field("config").Short("Pulumi stack config").
		Long("Map of Pulumi stack configuration values as key-value pairs. Supports templating. Keys use the format 'namespace:key' (e.g., 'aws:region')").
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables passed to Pulumi as key-value pairs").
		Field("public_repo").Short("public repository configuration").
		Long("Configuration for a public repository accessible without authentication").
		Field("connected_repo").Short("connected repository configuration").
		Long("Configuration for a private repository connected to the Nuon platform").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled. Supports templating").
		Example("0 2 * * *").
		Field("build_timeout").Short("build operation timeout").
		Long("Duration string for build operations (e.g., \"30m\", \"1h\"). Default: 5m. Max: 1h").
		Default("5m").
		Field("deploy_timeout").Short("deploy operation timeout").
		Long("Duration string for deploy operations (e.g., \"30m\", \"1h\"). Default: 60m. Max: 1h").
		Default("60m").
		Field("max_auto_retries").Short("maximum automatic retry attempts on deploy failure").
		Long("Maximum number of automatic retry attempts for failed deployments. Set to 0 to disable auto-retry. Default: 0 (disabled)").
		Default("0").
		Example("3").
		Example("5")
}

func (p *PulumiComponentConfig) Parse() error {
	return nil
}

func (p *PulumiComponentConfig) Validate() error {
	return nil
}
