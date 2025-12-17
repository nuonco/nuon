package config

import (
	"fmt"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
)

type ActionConfig struct {
	Name     string                 `mapstructure:"name" jsonschema:"required"`
	Timeout  string                 `mapstructure:"timeout,omitempty"`
	Triggers []*ActionTriggerConfig `mapstructure:"triggers" jsonschema:"required"`
	Steps    []*ActionStepConfig    `mapstructure:"steps" jsonschema:"required"`

	References     []refs.Ref `mapstructure:"-" jsonschema:"-"`
	Dependencies   []string   `mapstructure:"dependencies,omitempty"`
	BreakGlassRole string     `mapstructure:"break_glass_role,omitempty"`
}

type ActionTriggerConfig struct {
	Type string `mapstructure:"type" jsonschema:"required"`

	Index         int64  `mapstructure:"index,omitempty"`
	CronSchedule  string `mapstructure:"cron_schedule,omitempty"`
	ComponentName string `mapstructure:"component_name,omitempty"`
}

type ActionStepConfig struct {
	Name          string               `mapstructure:"name" jsonschema:"required"`
	EnvVarMap     map[string]string    `mapstructure:"env_vars,omitempty"`
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"`

	Command        string `mapstructure:"command" features:"template"`
	InlineContents string `mapstructure:"inline_contents" features:"get,template"`

	// created during parsing
	References []refs.Ref `mapstructure:"-" jsonschema:"-"`
}

func (a ActionConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the action").Required().
		Long("The action name is displayed in the Actions tab of the Nuon dashboard").
		Example("http_healthcheck").
		Example("database_migration").
		Field("timeout").Short("timeout for action execution").Required().
		Long("Maximum time the action can run. Maximum allowed is 30 minutes. Must be a valid Go duration string (e.g., 30s, 5m, 30m)").
		Example("15s").
		Example("5m").
		Example("30m").
		Field("triggers").Short("triggers that execute this action").Required().
		Long("Actions can be triggered manually, on a cron schedule, or by install lifecycle events (provision, deploy, teardown, etc). Define multiple triggers if needed").
		Field("steps").Short("steps to execute in this action").Required().
		Long("Ordered list of steps to execute. Each step requires a command and can optionally load scripts from repositories").
		Field("dependencies").Short("component dependencies referenced in this action").
		Long("Automatically extracted from template references in steps (e.g., {{.component.component_name}})").
		Field("break_glass_role").Short("IAM role for break-glass access to this action").
		Long("When set, allows the action to use a break glass role for elevated permissions during critical operations. Break glass roles are defined in CloudFormation stacks deployed to the customer's AWS account and provide temporary elevated access for emergency situations, migrations, or customer-initiated opt-in operations. See https://docs.nuon.co/updates/020-break-glass-actions for configuration details").
		Example("bucket-operations-break-glass").
		Example("database-migration-break-glass")
}

func (a ActionTriggerConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("type of trigger").Required().
		Long("Supported trigger types: manual, cron, pre-provision, post-provision, pre-reprovision, post-reprovision, pre-deprovision, post-deprovision, pre-deploy-all-components, post-deploy-all-components, pre-teardown-all-components, post-teardown-all-components, pre-deploy-component, post-deploy-component, pre-teardown-component, post-teardown-component, pre-deprovision-sandbox, post-deprovision-sandbox, pre-reprovision-sandbox, post-reprovision-sandbox, pre-update-inputs, post-update-inputs, pre-secrets-sync, post-secrets-sync").
		Example("manual").
		Example("cron").
		Example("post-provision").
		Field("cron_schedule").Short("cron schedule expression for scheduled triggers").
		Long("Standard cron format (minute hour day month weekday). For example, '*/5 * * * *' runs every 5 minutes").
		Example("*/5 * * * *").
		Example("0 */4 * * *").
		Field("component_name").Short("component name for component-specific triggers").
		Long("Required for pre-deploy-component, post-deploy-component, pre-teardown-component, post-teardown-component triggers").
		Example("database").
		Example("api-server").
		Field("index").Short("index for manual trigger").
		Long("Used to differentiate multiple manual triggers in the same action")
}

func (a ActionStepConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the step").Required().
		Long("Displayed in action logs and the Nuon dashboard").
		Example("healthcheck").
		Example("database_migration").
		Field("command").Short("command to execute").
		Long("Required field. Supports Go templating (e.g., {{.nuon.install.id}}). The command is executed in the runner environment").
		Example("./healthcheck").
		Example("bash -c 'curl https://example.com'").
		Field("env_vars").Short("environment variables to pass to the step").
		Long("Map of environment variables that will be available to the command. Supports Go templating for values").
		Field("public_repo").Short("public repository containing the step script").OneOfRequired("script_source").
		Long("Clone a public GitHub repository to load scripts from. Requires 'repo', 'branch', and optionally 'directory' fields").
		Field("connected_repo").Short("connected repository containing the step script").OneOfRequired("script_source").
		Long("Use a Nuon-connected repository to load scripts from. Requires 'repo', 'branch', and optionally 'directory' fields").
		Field("inline_contents").Short("inline script contents").
		Long("Embed script contents directly in the config file. Supports Go templating and external URLs: HTTP(S) (https://example.com/script.sh), git repositories (git::https://github.com/org/repo//path/to/script), file paths (file:///path/to/script.sh), and relative paths (./local/path)")
}

func (a *ActionConfig) parse() error {
	if a == nil {
		return nil
	}

	if a.Timeout != "" {
		_, err := time.ParseDuration(a.Timeout)
		if err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("unable to parse timeout %s", a.Timeout),
				Err:         err,
			}
		}
	}

	references, err := refs.Parse(a)
	if err != nil {
		return errors.Wrap(err, "unable to parse components")
	}
	a.References = references

	for _, ref := range a.References {
		if !generics.SliceContains(ref.Type, []refs.RefType{refs.RefTypeComponents}) {
			continue
		}

		a.Dependencies = append(a.Dependencies, ref.Name)
	}
	a.Dependencies = generics.UniqueSlice(a.Dependencies)
	return nil
}
