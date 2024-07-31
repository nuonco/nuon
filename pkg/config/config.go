package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/invopop/jsonschema"
)

const (
	currentVersion string = "v1"
)

type AppConfig struct {
	// Config file version
	Version string `mapstructure:"version" jsonschema:"required"`

	// Description for your app, which is rendered in the installers
	Description string `mapstructure:"description,omitempty" jsonschema:"required"`
	// Display name for the app, rendered in the installer
	DisplayName string `mapstructure:"display_name,omitempty" jsonschema:"required"`
	// Slack webhook url to receive notifications
	SlackWebhookURL string `mapstructure:"slack_webhook_url"`

	// Input configuration
	Inputs *AppInputConfig `mapstructure:"inputs"`
	// Sandbox configuration
	Sandbox *AppSandboxConfig `mapstructure:"sandbox" jsonschema:"required"`
	// Runner configuration
	Runner *AppRunnerConfig `mapstructure:"runner" jsonschema:"required"`
	// Installer configuration
	Installer *InstallerConfig `mapstructure:"installer"`

	// NOTE: in order to prevent users having to declare multiple arrays of _different_ component types:
	// eg: [[terraform_module_components]]
	// eg: [[helm_chart_components]]
	// we have one flat type, and convert the toml to a mapstructure.
	// This requires a bit more work/indirection by us, but a bit less by our customers!

	// Components are used to connect container images, automation and infrastructure as code to your Nuon App
	Components []*Component `mapstructure:"components"`
}

func (a AppConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "version", "Config file version.")
	addDescription(schema, "display_name", "Display name which is rendered in the installer.")
	addDescription(schema, "description", "App description which is rendered in the installer.")
	addDescription(schema, "slack_webhook_url", "Optional notifications channel to send app notifications to.")

	addDescription(schema, "inputs", "Inputs configuration object")
	addDescription(schema, "sandbox", "Sandbox configuration object")
	addDescription(schema, "installer", "Installer configuration object")
	addDescription(schema, "components", "Component configurations")
}

func (a *AppConfig) validateVersion() error {
	if a.Version != currentVersion {
		return ErrConfig{
			Description: "version must be v1",
		}
	}

	return nil
}

func (a *AppConfig) Validate(v *validator.Validate) error {
	fns := []func() error{
		func() error {
			return v.Struct(a)
		},
		a.validateVersion,
		a.Installer.Validate,
	}
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

type parseFn struct {
	name string
	fn   func() error
}

// NOTE(jm): this should go away completely, with decoder hooks
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

	for _, parseFn := range parseFns {
		if err := parseFn.fn(); err != nil {
			return fmt.Errorf("error parsing %s: %w", parseFn.name, err)
		}
	}

	return nil
}
