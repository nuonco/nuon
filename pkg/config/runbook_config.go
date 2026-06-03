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
	RunbookStepTypeDeploy RunbookStepType = "deploy"
	RunbookStepTypeAction RunbookStepType = "action"
)

type RunbookConfig struct {
	Name        string               `mapstructure:"name" toml:"name" jsonschema:"required"`
	Description string               `mapstructure:"description,omitempty" toml:"description,omitempty"`
	Readme      string               `mapstructure:"readme,omitempty" toml:"readme,omitempty" features:"get,template"`
	Labels      map[string]string    `mapstructure:"labels,omitempty" toml:"labels,omitempty"`
	Steps       []*RunbookStepConfig `mapstructure:"steps" toml:"steps" jsonschema:"required"`

	References   []refs.Ref `mapstructure:"-" jsonschema:"-"`
	Dependencies []string   `mapstructure:"dependencies,omitempty" toml:"dependencies,omitempty"`
}

type RunbookStepConfig struct {
	Name string          `mapstructure:"name" toml:"name" jsonschema:"required"`
	Type RunbookStepType `mapstructure:"type" toml:"type" jsonschema:"required"`

	// For type = "deploy"
	ComponentName      string `mapstructure:"component_name,omitempty" toml:"component_name,omitempty"`
	DeployDependencies bool   `mapstructure:"deploy_dependencies,omitempty" toml:"deploy_dependencies,omitempty"`

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
		Long("Either 'deploy' for deploying a component, or 'action' for running an action").
		Example("deploy").
		Example("action").
		Field("component_name").Short("component to deploy (for deploy steps)").
		Long("Name of the component to deploy. Required when type is 'deploy'").
		Example("database").
		Example("api-server").
		Field("deploy_dependencies").Short("also deploy transitive dependencies").
		Long("When true, deploys the component and all its transitive dependencies in dependency order. Only applies to deploy steps").
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
		Long("IAM role name to use when executing the inline action step")
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
