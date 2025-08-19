package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
)

type AppConfig struct {
	// Config file version
	Version string `mapstructure:"version" jsonschema:"required"`

	// Description for your app, which is rendered in the installers
	Description string `mapstructure:"description,omitempty"`
	// Display name for the app, rendered in the installer
	DisplayName string `mapstructure:"display_name,omitempty"`
	// Slack webhook url to receive notifications
	SlackWebhookURL string `mapstructure:"slack_webhook_url"`
	// Readme for the app
	Readme string `mapstructure:"readme,omitempty" features:"get,template"`

	// Default App Branch config
	Branch *AppBranchConfig `mapstructure:"branch,omitempty"`
	// Input configuration
	Inputs *AppInputConfig `mapstructure:"inputs,omitempty"`
	// Sandbox configuration
	Sandbox *AppSandboxConfig `mapstructure:"sandbox" jsonschema:"required"`
	// Runner configuration
	Runner *AppRunnerConfig `mapstructure:"runner" jsonschema:"required"`
	// Installer configuration
	Installer *InstallerConfig `mapstructure:"installer,omitempty"`
	// Permissions config
	Permissions *PermissionsConfig `mapstructure:"permissions,omitempty"`
	// Policies config
	Policies *PoliciesConfig `mapstructure:"policies,omitempty"`
	// Secrets config
	Secrets *SecretsConfig `mapstructure:"secrets,omitempty"`
	// Break-glass config
	BreakGlass *BreakGlass `mapstructure:"break_glass,omitempty"`
	// Stack config
	Stack *StackConfig `mapstructure:"stack,omitempty"`

	// NOTE: in order to prevent users having to declare multiple arrays of _different_ component types:
	// eg: [[terraform_module_components]]
	// eg: [[helm_chart_components]]
	// we have one flat type, and convert the toml to a mapstructure.
	// This requires a bit more work/indirection by us, but a bit less by our customers!

	// Components are used to connect container images, automation and infrastructure as code to your Nuon App
	Components ComponentList `mapstructure:"components,omitempty"`

	Installs []*Install `mapstructure:"installs,omitempty"`

	Actions []*ActionConfig `mapstructure:"actions,omitempty"`
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
	addDescription(schema, "version", "Config file version.")
	addDescription(schema, "display_name", "Display name which is rendered in the installer.")
	addDescription(schema, "description", "App description which is rendered in the installer.")
	addDescription(schema, "slack_webhook_url", "Optional notifications channel to send app notifications to.")

	addDescription(schema, "branch", "Default app branch configuration object")
	addDescription(schema, "inputs", "Inputs configuration object")
	addDescription(schema, "sandbox", "Sandbox configuration object")
	addDescription(schema, "installer", "Installer configuration object")
	addDescription(schema, "components", "Component configurations")
	addDescription(schema, "installs", "Install configurations")
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
	if a.Installer != nil {
		parseFns = append(parseFns, parseFn{
			"installer",
			a.Installer.parse,
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

	for _, parseFn := range parseFns {
		if err := parseFn.fn(); err != nil {
			return fmt.Errorf("error parsing %s: %w", parseFn.name, err)
		}
	}

	return nil
}
