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
	RunbookStepTypeDeploy             RunbookStepType = "deploy"
	RunbookStepTypeAction             RunbookStepType = "action"
	RunbookStepTypeSandboxReprovision RunbookStepType = "sandbox_reprovision"
	RunbookStepTypeSandboxDeprovision RunbookStepType = "sandbox_deprovision"
)

type RunbookCellType string

const (
	RunbookCellTypeMarkdown RunbookCellType = "markdown"
	RunbookCellTypeDeploy   RunbookCellType = "deploy"
	RunbookCellTypeAction   RunbookCellType = "action"
)

type RunbookCellConfig struct {
	Type    RunbookCellType `mapstructure:"type" toml:"type" jsonschema:"required"`
	Content string          `mapstructure:"content,omitempty" toml:"content,omitempty" features:"get,template"`
	Name    string          `mapstructure:"name,omitempty" toml:"name,omitempty"`

	// Fields mirrored from RunbookStepConfig for deploy/action cells
	ComponentName      string            `mapstructure:"component_name,omitempty" toml:"component_name,omitempty"`
	DeployDependencies bool              `mapstructure:"deploy_dependencies,omitempty" toml:"deploy_dependencies,omitempty"`
	ActionName         string            `mapstructure:"action_name,omitempty" toml:"action_name,omitempty"`
	Command            string            `mapstructure:"command,omitempty" toml:"command,omitempty" features:"template"`
	InlineContents     string            `mapstructure:"inline_contents,omitempty" toml:"inline_contents,omitempty" features:"get,template"`
	EnvVarMap          map[string]string `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`
	Timeout            string            `mapstructure:"timeout,omitempty" toml:"timeout,omitempty"`
	Role               string            `mapstructure:"role,omitempty" toml:"role,omitempty"`
}

func (r RunbookCellConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("type of cell").Required().
		Long("Either 'markdown' for documentation, 'deploy' for deploying a component, or 'action' for running an action").
		Example("markdown").
		Example("deploy").
		Example("action").
		Field("content").Short("markdown content (for markdown cells)").
		Long("Markdown text displayed between executable steps. Supports Go templating with install state variables").
		Field("name").Short("name of the step (for deploy/action cells)").
		Long("Displayed in the workflow UI and runbook detail page. Required for deploy and action cells").
		Example("deploy-database").
		Example("run-migrations").
		Field("component_name").Short("component to deploy (for deploy cells)").
		Long("Name of the component to deploy. Required when type is 'deploy'").
		Field("deploy_dependencies").Short("also deploy transitive dependencies").
		Long("When true, deploys the component and all its transitive dependencies in dependency order. Only applies to deploy cells").
		Field("action_name").Short("existing action to run (for action cells)").
		Long("Name of a previously defined action workflow to execute. Mutually exclusive with inline action fields").
		Field("command").Short("command to execute (for inline action cells)").
		Long("Shell command for an inline action. Supports Go templating").
		Field("inline_contents").Short("inline script contents (for inline action cells)").
		Long("Embed script contents directly or reference an external file. Supports Go templating and external URLs").
		Field("env_vars").Short("environment variables for inline action cells").
		Long("Map of environment variables passed to the inline action command").
		Field("timeout").Short("timeout for inline action cells").
		Long("Maximum execution time. Must be a valid Go duration string").
		Example("30s").
		Example("5m").
		Field("role").Short("IAM role for inline action execution").
		Long("IAM role name to use when executing the inline action cell")
}

type RunbookConfig struct {
	Name        string               `mapstructure:"name" toml:"name" jsonschema:"required"`
	Description string               `mapstructure:"description,omitempty" toml:"description,omitempty"`
	Readme      string               `mapstructure:"readme,omitempty" toml:"readme,omitempty" features:"get,template"`
	Labels      map[string]string    `mapstructure:"labels,omitempty" toml:"labels,omitempty"`
	Steps       []*RunbookStepConfig `mapstructure:"steps,omitempty" toml:"steps,omitempty"`
	Cells       []*RunbookCellConfig `mapstructure:"cells,omitempty" toml:"cells,omitempty"`

	References   []refs.Ref `mapstructure:"-" jsonschema:"-"`
	Dependencies []string   `mapstructure:"dependencies,omitempty" toml:"dependencies,omitempty"`
}

type RunbookStepConfig struct {
	Name string          `mapstructure:"name" toml:"name" jsonschema:"required"`
	Type RunbookStepType `mapstructure:"type" toml:"type" jsonschema:"required"`

	// For type = "deploy"
	ComponentName      string `mapstructure:"component_name,omitempty" toml:"component_name,omitempty"`
	DeployDependencies bool   `mapstructure:"deploy_dependencies,omitempty" toml:"deploy_dependencies,omitempty"`

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
		Field("steps").Short("ordered steps to execute in the runbook").
		Long("Sequential list of deploy and action steps. Each step executes in order. Deploy steps can include dependency deployment. Action steps can reference existing actions or define inline actions. Mutually exclusive with cells").
		Field("cells").Short("notebook-style cells interleaving documentation with steps").
		Long("Ordered list of cells that can be markdown documentation or executable steps (deploy/action). Cells are normalized to steps at config creation time. Mutually exclusive with steps")
}

func (r RunbookStepConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("name of the step").Required().
		Long("Displayed in the workflow UI and runbook detail page").
		Example("deploy-database").
		Example("run-migrations").
		Field("type").Short("type of step").Required().
		Long("One of: 'deploy' (deploy a component), 'action' (run an action), 'sandbox_reprovision', or 'sandbox_deprovision' (run the corresponding sandbox lifecycle plan + apply)").
		Example("deploy").
		Example("action").
		Example("sandbox_reprovision").
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
		Long("IAM role name to use when executing the inline action step").
		Field("skip_component_deploys").Short("skip component deployments after sandbox reprovision").
		Long("Only applies to 'sandbox_reprovision' steps. When true, only the sandbox infrastructure is reprovisioned and components are NOT redeployed on top. Matches the dashboard's 'Skip component deployments' option")
}

func (r *RunbookConfig) normalizeCellsToSteps() error {
	if len(r.Cells) == 0 {
		return nil
	}
	if len(r.Steps) > 0 {
		return ErrConfig{
			Description: "runbook config cannot have both cells and steps; use one or the other",
		}
	}

	for _, cell := range r.Cells {
		switch cell.Type {
		case RunbookCellTypeMarkdown:
			continue
		case RunbookCellTypeDeploy, RunbookCellTypeAction:
			r.Steps = append(r.Steps, &RunbookStepConfig{
				Name:               cell.Name,
				Type:               RunbookStepType(cell.Type),
				ComponentName:      cell.ComponentName,
				DeployDependencies: cell.DeployDependencies,
				ActionName:         cell.ActionName,
				Command:            cell.Command,
				InlineContents:     cell.InlineContents,
				EnvVarMap:          cell.EnvVarMap,
				Timeout:            cell.Timeout,
				Role:               cell.Role,
			})
		default:
			return ErrConfig{
				Description: fmt.Sprintf("unknown cell type %q", cell.Type),
			}
		}
	}

	return nil
}

func (r *RunbookConfig) parse() error {
	if r == nil {
		return nil
	}

	if err := r.normalizeCellsToSteps(); err != nil {
		return err
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
