package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon/pkg/generics"
)

type AppConfig struct {
	// Config file version
	Version string `mapstructure:"version" toml:"version" jsonschema:"required"`

	// Description for your app, which is rendered in the installers
	Description string `mapstructure:"description,omitempty" toml:"description,omitempty"`
	// Display name for the app, rendered in the installer
	DisplayName string `mapstructure:"display_name,omitempty" toml:"display_name,omitempty"`
	// Slack webhook url to receive notifications
	SlackWebhookURL string `mapstructure:"slack_webhook_url" toml:"slack_webhook_url"`
	// Readme for the app
	Readme string `mapstructure:"readme,omitempty" toml:"readme,omitempty" features:"get,template"`
	// Color codes for label keys
	LabelColors map[string]string `mapstructure:"label_colors,omitempty" toml:"label_colors,omitempty"`

	// Default App Branch config
	Branch *AppBranchConfig `mapstructure:"branch,omitempty" toml:"branch,omitempty"`
	// App branch configs (from branches/ directory)
	Branches []*AppBranchConfig `mapstructure:"branches,omitempty" toml:"branches,omitempty"`
	// Input configuration
	Inputs *AppInputConfig `mapstructure:"inputs,omitempty" toml:"inputs,omitempty"`
	// Sandbox configuration
	Sandbox *AppSandboxConfig `mapstructure:"sandbox" toml:"sandbox" jsonschema:"required"`
	// Runner configuration
	Runner *AppRunnerConfig `mapstructure:"runner" toml:"runner" jsonschema:"required"`
	// Permissions config
	Permissions *PermissionsConfig `mapstructure:"permissions,omitempty" toml:"permissions,omitempty"`
	// Policies config
	Policies *PoliciesConfig `mapstructure:"policies,omitempty" toml:"policies,omitempty"`
	// Secrets config
	Secrets *SecretsConfig `mapstructure:"secrets,omitempty" toml:"secrets,omitempty"`
	// Break-glass config
	BreakGlass *BreakGlass `mapstructure:"break_glass,omitempty" toml:"break_glass,omitempty"`
	// Stack config
	Stack *StackConfig `mapstructure:"stack,omitempty" toml:"stack,omitempty"`
	// Operation rules
	OperationRoles *OperationRolesConfig `mapstructure:"operation_roles,omitempty" toml:"operation_roles,omitempty"`
	// Kubernetes contexts
	KubernetesContexts *KubernetesContextsConfig `mapstructure:"kubernetes_contexts,omitempty" toml:"kubernetes_contexts,omitempty"`

	// NOTE: in order to prevent users having to declare multiple arrays of _different_ component types:
	// eg: [[terraform_module_components]]
	// eg: [[helm_chart_components]]
	// we have one flat type, and convert the toml to a mapstructure.
	// This requires a bit more work/indirection by us, but a bit less by our customers!

	// Components are used to connect container images, automation and infrastructure as code to your Nuon App
	Components ComponentList `mapstructure:"components,omitempty" toml:"components,omitempty"`

	Installs       []*Install      `mapstructure:"installs,omitempty" toml:"installs,omitempty"`
	InstallsConfig *InstallsConfig `mapstructure:"installs_config,omitempty" toml:"installs_config,omitempty"`

	Actions []*ActionConfig `mapstructure:"actions,omitempty" toml:"actions,omitempty"`

	Runbooks []*RunbookConfig `mapstructure:"runbooks,omitempty" toml:"runbooks,omitempty"`

	Triggers *TriggersConfig `mapstructure:"triggers,omitempty" toml:"triggers,omitempty"`
}
type ComponentList []*Component

func (a *ComponentList) Validate() error {
	for _, comp := range *a {
		if err := comp.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a AppConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("version").Short("config file version").Required().
		Long("Version string for the configuration format").
		Example("1").
		Example("2").
		Field("display_name").Short("display name rendered in the installer").
		Long("Display name for the app shown to customers during installation").
		Example("My SaaS Application").
		Example("Analytics Platform").
		Field("description").Short("app description rendered in the installer").
		Long("Detailed description of the app shown to customers during installation").
		Example("A cloud-native SaaS application").
		Example("Enterprise analytics and reporting solution").
		Field("slack_webhook_url").Short("slack webhook url for notifications").
		Long("Optional Slack webhook URL to send app notifications and alerts to a channel").
		Field("readme").Short("readme file for the app").
		Long("Markdown file with app documentation. Supports templating and external file sources: HTTP(S) URLs (https://example.com/readme.md), git repositories (git::https://github.com/org/repo//readme.md), file paths (file:///path/to/readme.md), and relative paths (./readme.md)").
		Field("label_colors").Short("label key color codes").
		Long("Map of label key names to hex color codes for customizing label display in the dashboard").
		Field("branch").Short("default app branch configuration").
		Long("Default branch configuration for all installs. Can be overridden per install").
		Field("inputs").Short("input configuration").
		Long("Define inputs that customers provide during installation").
		Field("sandbox").Short("sandbox configuration").Required().
		Long("Sandbox/cluster configuration for the application infrastructure").
		Field("runner").Short("runner configuration").Required().
		Long("Runner configuration for executing deployments and workflows").
		Field("installer").Short("installer configuration").
		Long("Configuration for the customer-facing installer experience").
		Field("permissions").Short("permissions configuration").
		Long("RBAC and permission policies for install access control").
		Field("policies").Short("policies configuration").
		Long("Define policies for install management and automation").
		Field("secrets").Short("secrets configuration").
		Long("Manage secret variables shared across components").
		Field("break_glass").Short("break-glass configuration").
		Long("Configure break-glass roles for emergency access to installs").
		Field("stack").Short("stack configuration").
		Long("Stack configuration for infrastructure orchestration").
		Field("kubernetes_contexts").Short("kubernetes context bindings").
		Long("Named bindings to peer components that produce cluster connection details. Components opt in via their top-level kubernetes_context field").
		Field("components").Short("component configurations").
		Long("List of components (terraform, helm, containers, etc) to deploy").
		Field("installs").Short("install configurations").
		Long("Install-specific overrides and configurations").
		Field("actions").Short("action configurations").
		Long("Custom workflows and actions that can be executed on installs").
		Field("runbooks").Short("runbook configurations").
		Long("Ordered release procedures with deploy and action steps that can be executed on installs").
		Field("triggers").Short("trigger configuration").
		Long("Configure event sources and rules that trigger app branch runs")
}

type parseFn struct {
	name string
	fn   func() error
}

func (a *AppConfig) Parse() error {
	parseFns := []parseFn{
		{
			"sandbox",
			a.Sandbox.parse,
		},
		{
			"runner",
			a.Runner.parse,
		},
	}

	if a.Branch != nil {
		parseFns = append(parseFns, parseFn{
			"branch",
			a.Branch.parse,
		})
	}
	if a.Inputs != nil {
		parseFns = append(parseFns, parseFn{
			"inputs",
			a.Inputs.parse,
		})
	}
	if a.Permissions != nil {
		parseFns = append(parseFns, parseFn{
			"permissions",
			a.Permissions.parse,
		})
	}
	if a.Secrets != nil {
		parseFns = append(parseFns, parseFn{
			"secrets",
			a.Secrets.parse,
		})
	}
	if a.Policies != nil {
		parseFns = append(parseFns, parseFn{
			"policies",
			a.Policies.parse,
		})
	}
	if a.Stack != nil {
		parseFns = append(parseFns, parseFn{
			"stack",
			a.Stack.parse,
		})
	}
	if a.KubernetesContexts != nil {
		parseFns = append(parseFns, parseFn{
			"kubernetes_contexts",
			a.KubernetesContexts.parse,
		})
	}

	for idx, action := range a.Actions {
		parseFns = append(parseFns, parseFn{
			fmt.Sprintf("actions.%d", idx),
			action.parse,
		})
	}

	for idx, comp := range a.Components {
		parseFns = append(parseFns, parseFn{
			fmt.Sprintf("components.%d", idx),
			comp.parse,
		})
	}

	for idx, install := range a.Installs {
		parseFns = append(parseFns, parseFn{
			fmt.Sprintf("installs.%d", idx),
			install.Parse,
		})
	}

	for idx, runbook := range a.Runbooks {
		parseFns = append(parseFns, parseFn{
			fmt.Sprintf("runbooks.%d", idx),
			runbook.parse,
		})
	}

	for _, parseFn := range parseFns {
		if err := parseFn.fn(); err != nil {
			return fmt.Errorf("error parsing %s: %w", parseFn.name, err)
		}
	}

	// Default cloud_platform on all roles from runner type if not explicitly set
	if err := a.defaultCloudPlatformFromRunner(); err != nil {
		return err
	}

	// Resolve kubernetes_contexts <-> component wiring after both are parsed.
	if err := a.resolveKubernetesContexts(); err != nil {
		return err
	}

	return nil
}

// resolveKubernetesContexts cross-validates the kubernetes_contexts block
// against the components list and injects an implicit dependency from each
// component using a context to that context's source component, so build /
// deploy ordering works without users also listing the cluster component in
// `dependencies`.
//
// TODO(apps-sync-validate): warn (but don't fail) when a referenced source
// component's TF/Pulumi source doesn't declare cluster.{name,endpoint,...}
// outputs. Static validation belongs at apps-sync time; runtime template
// failure during deploy planning is acceptable for now.
func (a *AppConfig) resolveKubernetesContexts() error {
	componentsByName := make(map[string]*Component, len(a.Components))
	for _, c := range a.Components {
		if c == nil {
			continue
		}
		componentsByName[c.Name] = c
	}

	contextsByName := map[string]*KubernetesContext{}
	if a.KubernetesContexts != nil {
		for _, ctx := range a.KubernetesContexts.Contexts {
			if ctx == nil {
				continue
			}
			source, ok := componentsByName[ctx.Component]
			if !ok {
				return fmt.Errorf("kubernetes_context %q references unknown component %q", ctx.Name, ctx.Component)
			}
			if source.Type != TerraformModuleComponentType && source.Type != PulumiComponentType {
				return fmt.Errorf("kubernetes_context %q source component %q must be terraform_module or pulumi (got %q)", ctx.Name, ctx.Component, source.Type)
			}
			contextsByName[ctx.Name] = ctx
		}
	}

	for _, c := range a.Components {
		if c == nil || c.KubernetesContext == "" {
			continue
		}
		ctx, ok := contextsByName[c.KubernetesContext]
		if !ok {
			return fmt.Errorf("component %q references unknown kubernetes_context %q", c.Name, c.KubernetesContext)
		}
		c.AddDependency(ctx.Component)
	}

	for _, action := range a.Actions {
		if action == nil || action.KubernetesContext == "" {
			continue
		}
		ctx, ok := contextsByName[action.KubernetesContext]
		if !ok {
			return fmt.Errorf("action %q references unknown kubernetes_context %q", action.Name, action.KubernetesContext)
		}
		action.Dependencies = generics.UniqueSlice(append(action.Dependencies, ctx.Component))
	}

	return nil
}

func (a *AppConfig) defaultCloudPlatformFromRunner() error {
	if a.Runner == nil || a.Runner.RunnerType == "" {
		return nil
	}

	platform := a.Runner.RunnerType

	setDefault := func(role *AppAWSIAMRole) error {
		if role == nil {
			return nil
		}
		if role.CloudPlatform == "" {
			role.CloudPlatform = platform
			return nil
		}
		if role.CloudPlatform != platform {
			display := role.DisplayName
			if display == "" {
				display = role.Description
			}
			return fmt.Errorf("role %q has cloud_platform %q but runner_type is %q", display, role.CloudPlatform, platform)
		}
		return nil
	}

	if a.Permissions != nil {
		for _, role := range a.Permissions.Roles {
			if err := setDefault(role); err != nil {
				return err
			}
		}
		for _, role := range a.Permissions.CustomRoles {
			if err := setDefault(role); err != nil {
				return err
			}
		}
	}

	if a.BreakGlass != nil {
		for _, role := range a.BreakGlass.Roles {
			if err := setDefault(role); err != nil {
				return err
			}
		}
	}

	return nil
}
