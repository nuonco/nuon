package config

import (
	"fmt"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
)

type RunbookStepType string

const (
	RunbookStepTypeComponentDeploy    RunbookStepType = "component_deploy"
	RunbookStepTypeComponentTearDown  RunbookStepType = "component_tear_down"
	RunbookStepTypeAction             RunbookStepType = "action"
	RunbookStepTypeSandboxReprovision RunbookStepType = "sandbox_reprovision"
	RunbookStepTypeSandboxDeprovision RunbookStepType = "sandbox_deprovision"

	// RunbookStepTypeDeployLegacy is the prior name for component_deploy. Accepted
	// as input and canonicalized to component_deploy at parse/ingress time.
	RunbookStepTypeDeployLegacy RunbookStepType = "deploy"
)

type RunbookConfig struct {
	Name        string               `mapstructure:"name" toml:"name" jsonschema:"required"`
	Description string               `mapstructure:"description,omitempty" toml:"description,omitempty"`
	Readme      string               `mapstructure:"readme,omitempty" toml:"readme,omitempty" features:"get,template"`
	Labels      map[string]string    `mapstructure:"labels,omitempty" toml:"labels,omitempty"`
	Steps       []*RunbookStepConfig `mapstructure:"steps" toml:"steps" jsonschema:"required"`

	References   []refs.Ref `mapstructure:"-" jsonschema:"-"`
	Dependencies []string   `mapstructure:"dependencies,omitempty" toml:"dependencies,omitempty"`

	// DeprecationWarnings collects messages about legacy field usage observed during parse().
	// Populated by parse(); consumed by callers (e.g. the CLI sync) to surface to the user.
	DeprecationWarnings []string `mapstructure:"-" toml:"-" jsonschema:"-"`
}

type RunbookStepConfig struct {
	Name string          `mapstructure:"name" toml:"name" jsonschema:"required"`
	Type RunbookStepType `mapstructure:"type" toml:"type" jsonschema:"required"`

	// For type = "component_deploy" / "component_tear_down"
	ComponentName      string `mapstructure:"component_name,omitempty" toml:"component_name,omitempty"`
	DeployDependents   bool   `mapstructure:"deploy_dependents,omitempty" toml:"deploy_dependents,omitempty"`
	TearDownDependents bool   `mapstructure:"tear_down_dependents,omitempty" toml:"tear_down_dependents,omitempty"`

	// Legacy alias for DeployDependents — kept for back-compat with TOML configs
	// written before the rename. Folded into DeployDependents in parse().
	DeployDependenciesLegacy bool `mapstructure:"deploy_dependencies,omitempty" toml:"deploy_dependencies,omitempty"`

	// For type = "sandbox_reprovision" — when true, only run the sandbox infra plan + apply
	// and do NOT redeploy components on top.
	SkipComponentDeploys bool `mapstructure:"skip_component_deploys,omitempty" toml:"skip_component_deploys,omitempty"`

	// For type = "action" — reference existing action
	ActionName string `mapstructure:"action_name,omitempty" toml:"action_name,omitempty"`

	// For type = "action" — inline action (same fields as ActionStepConfig)
	Command        string            `mapstructure:"command,omitempty" toml:"command,omitempty" features:"template"`
	InlineContents string            `mapstructure:"inline_contents,omitempty" toml:"inline_contents,omitempty" features:"get,template"`
	EnvVarMap      map[string]string `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`
	Timeout        string            `mapstructure:"timeout,omitempty" toml:"timeout,omitempty"`
	Role           string            `mapstructure:"role,omitempty" toml:"role,omitempty"`

	References []refs.Ref `mapstructure:"-" jsonschema:"-"`
}

func (r RunbookConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the runbook").Required().
		Long("The runbook name is displayed in the Runbooks tab of the Nuon dashboard and used to identify it during sync").
		Example("v2.3-update").
		Example("database-migration").
		Field("readme").Short("readme file for the runbook").
		Long("Markdown file with runbook documentation and instructions. Supports Go templating and external file sources: HTTP(S) URLs, git repositories, file paths, and relative paths").
		Example("./release-notes.md").
		Field("steps").Short("ordered steps to execute in the runbook").Required().
		Long("Sequential list of deploy and action steps. Each step executes in order. Deploy steps can include dependency deployment. Action steps can reference existing actions or define inline actions")
}

func (r RunbookStepConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the step").Required().
		Long("Displayed in the workflow UI and runbook detail page").
		Example("deploy-database").
		Example("run-migrations").
		Field("type").Short("type of step").Required().
		Long("One of: 'component_deploy' (deploy a component; 'deploy' is accepted as a legacy alias), 'component_tear_down' (tear down a component), 'action' (run an action), 'sandbox_reprovision', or 'sandbox_deprovision' (run the corresponding sandbox lifecycle plan + apply)").
		Example("component_deploy").
		Example("component_tear_down").
		Example("action").
		Example("sandbox_reprovision").
		Field("component_name").Short("component to deploy or tear down (for component steps)").
		Long("Name of the component to deploy or tear down. Required when type is 'component_deploy' or 'component_tear_down'").
		Example("database").
		Example("api-server").
		Field("deploy_dependents").Short("also deploy transitive dependents").
		Long("When true, deploys the component and all components that transitively depend on it (downstream), in dependency order. Only applies to component_deploy steps").
		Field("deploy_dependencies").Short("legacy alias for deploy_dependents").
		Deprecated("use 'deploy_dependents' instead").
		Field("tear_down_dependents").Short("also tear down transitive dependents").
		Long("When true, tears down the component and all components that transitively depend on it (downstream), with dependents torn down first. Only applies to component_tear_down steps").
		Field("action_name").Short("existing action to run (for action steps)").
		Long("Name of a previously defined action workflow to execute. Mutually exclusive with inline action fields (command, inline_contents)").
		Example("database-migration").
		Field("command").Short("command to execute (for inline action steps)").
		Long("Shell command for an inline action. Supports Go templating").
		Example("./validate.sh").
		Field("inline_contents").Short("inline script contents (for inline action steps)").
		Long("Embed script contents directly or reference an external file. Supports Go templating and external URLs").
		Example("./scripts/validate.sh").
		Field("env_vars").Short("environment variables for inline action steps").
		Long("Map of environment variables passed to the inline action command").
		Field("timeout").Short("timeout for inline action steps").
		Long("Maximum execution time for inline action steps. Must be a valid Go duration string").
		Example("30s").
		Example("5m").
		Field("role").Short("IAM role for inline action execution").
		Long("IAM role name to use when executing the inline action step").
		Field("skip_component_deploys").Short("skip component deployments after sandbox reprovision").
		Long("Only applies to 'sandbox_reprovision' steps. When true, only the sandbox infrastructure is reprovisioned and components are NOT redeployed on top. Matches the dashboard's 'Skip component deployments' option")
}

func (r *RunbookConfig) parse() error {
	if r == nil {
		return nil
	}

	for _, step := range r.Steps {
		if step.Timeout != "" {
			_, err := time.ParseDuration(step.Timeout)
			if err != nil {
				return ErrConfig{
					Description: fmt.Sprintf("unable to parse timeout %s for step %s", step.Timeout, step.Name),
					Err:         err,
				}
			}
		}
		// Fold the legacy alias into the canonical field. New code should only read DeployDependents.
		if step.DeployDependenciesLegacy {
			step.DeployDependents = true
			r.DeprecationWarnings = append(r.DeprecationWarnings, fmt.Sprintf("runbook %q step %q: 'deploy_dependencies' is deprecated, use 'deploy_dependents' instead", r.Name, step.Name))
		}
		// Canonicalize the legacy "deploy" type to "component_deploy".
		if step.Type == RunbookStepTypeDeployLegacy {
			step.Type = RunbookStepTypeComponentDeploy
			r.DeprecationWarnings = append(r.DeprecationWarnings, fmt.Sprintf("runbook %q step %q: type 'deploy' is deprecated, use 'component_deploy' instead", r.Name, step.Name))
		}
	}

	references, err := refs.Parse(r)
	if err != nil {
		return errors.Wrap(err, "unable to parse runbook")
	}
	r.References = references

	for _, ref := range r.References {
		if !generics.SliceContains(ref.Type, []refs.RefType{refs.RefTypeComponents}) {
			continue
		}
		r.Dependencies = append(r.Dependencies, ref.Name)
	}
	r.Dependencies = generics.UniqueSlice(r.Dependencies)
	return nil
}
