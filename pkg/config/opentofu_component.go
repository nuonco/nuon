package config

import (
	"github.com/invopop/jsonschema"
)

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type OpentofuModuleComponentConfig struct {
	OpenTofuVersion string `mapstructure:"open_tofu_version" toml:"open_tofu_version"`

	EnvVarMap      map[string]string        `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`
	VarsMap        map[string]string        `mapstructure:"vars,omitempty" toml:"vars,omitempty"`
	VariablesFiles []TerraformVariablesFile `mapstructure:"var_file,omitempty" toml:"var_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty"  jsonschema:"oneof_required=public_repo"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" toml:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`

	MaxAutoRetries               *int  `mapstructure:"max_auto_retries,omitempty" toml:"max_auto_retries,omitempty" nuonhash:"omitempty"`
	SkipNoops                    *bool `mapstructure:"skip_noops,omitempty" toml:"skip_noops,omitempty" nuonhash:"omitempty"`
	AutoApproveOnPoliciesPassing *bool `mapstructure:"auto_approve_on_policies_passing,omitempty" toml:"auto_approve_on_policies_passing,omitempty" nuonhash:"omitempty"`

	Teardown *ComponentTeardownConfig `mapstructure:"teardown,omitempty" toml:"teardown,omitempty" nuonhash:"omitempty"`
}

func (t OpentofuModuleComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("open_tofu_version").Short("OpenTofu version").
		Long("Version of OpenTofu to use for deployments. OpenTofu 1.13 or newer is required; if omitted, a pinned default (v1.13.9) is used").
		Example("1.13.9").
		Example("1.14.0").
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables passed to OpenTofu as key-value pairs").
		Field("vars").Short("OpenTofu variables").
		Long("Map of OpenTofu input variables as key-value pairs. Supports templating").
		Field("var_file").Short("OpenTofu variable files").
		Long("Array of external OpenTofu variable files to load. Each file contents support templating and external file sources: HTTP(S) URLs (https://example.com/vars.tfvars), git repositories (git::https://github.com/org/repo//path/to/vars.tfvars), file paths (file:///path/to/vars.tfvars), and relative paths (./vars.tfvars)").
		Field("public_repo").Short("public repository configuration").
		Long("Configuration for a public repository accessible without authentication").
		Field("connected_repo").Short("connected repository configuration").
		Long("Configuration for a private repository connected to the Nuon platform").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled. Supports templating").
		Example("0 2 * * *").
		Example("*/10 * * * *").
		Field("build_timeout").Short("build operation timeout").
		Long("Duration string for build operations (e.g., \"10m\", \"30m\"). Default: 10m. Max: 30m").
		Default("10m").
		Example("30m").
		Field("deploy_timeout").Short("deploy operation timeout").
		Long("Duration string for deploy operations (e.g., \"30m\", \"1h\"). Default: 60m. Max: 1h").
		Default("60m").
		Example("30m").
		Example("1h").
		Field("max_auto_retries").Short("maximum automatic retry attempts on deploy failure").
		Long("Maximum number of automatic retry attempts for failed deployments. Set to 0 to disable auto-retry. Default: 0 (disabled)").
		Default("0").
		Example("3").
		Example("5").
		Field("skip_noops").Short("Skip the deploy step when the plan has no changes (a no-op). Defaults to false").
		Default("false").
		Example("true").
		Field("auto_approve_on_policies_passing").Short("Auto-approve the deploy when all policy checks pass. Defaults to false").
		Default("false").
		Example("true").
		Field("teardown").Short("teardown behavior").
		Long("Controls what happens to the component's resources when it is torn down. Resources listed in retain are removed from OpenTofu state and left in place instead of being destroyed")
}

func (t *OpentofuModuleComponentConfig) Parse() error {
	return nil
}

func (t *OpentofuModuleComponentConfig) Validate() error {
	if err := t.Teardown.Validate(); err != nil {
		return ErrConfig{
			Description: err.Error(),
			Err:         err,
		}
	}

	return nil
}
