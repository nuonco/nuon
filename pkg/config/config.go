package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

const (
	currentVersion string = "v1"
)

type AppConfig struct {
	Version string `mapstructure:"version" toml:"version"`

	// Top level fields on the app itself, which are _not_ synced by this package
	Description     string `mapstructure:"description,omitempty" toml:"description"`
	DisplayName     string `mapstructure:"display_name,omitempty" toml:"display_name"`
	SlackWebhookURL string `mapstructure:"slack_webhook_url,omitempty" toml:"slack_webhook_url"`

	// top level fields
	Inputs    *AppInputConfig   `mapstructure:"inputs,omitempty"`
	Sandbox   *AppSandboxConfig `mapstructure:"sandbox"`
	Runner    *AppRunnerConfig  `mapstructure:"runner"`
	Installer *InstallerConfig  `mapstructure:"installer,omitempty"`

	// NOTE: in order to prevent users having to declare multiple arrays of _different_ component types:
	// eg: [[terraform_module_components]]
	// eg: [[helm_chart_components]]
	// we have one flat type, and convert the toml to a mapstructure.
	// This requires a bit more work/indirection by us, but a bit less by our customers!
	Components []*Component `mapstructure:"components"`
}

func (a *AppConfig) Validate(v *validator.Validate) error {
	if a.Version != currentVersion {
		return fmt.Errorf("version must be v1")
	}

	if err := v.Struct(a); err != nil {
		return err
	}

	return nil
}

type parseFn struct {
	name string
	fn   func(ConfigContext) error
}

func (a *AppConfig) Parse(ctx ConfigContext) error {
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

	for idx, comp := range a.Components {
		parseFns = append(parseFns, parseFn{
			name: fmt.Sprintf("component.%v", idx),
			fn:   comp.parse,
		})
	}

	for _, parseFn := range parseFns {
		if err := parseFn.fn(ctx); err != nil {
			return fmt.Errorf("error parsing %s: %w", parseFn.name, err)
		}
	}

	return nil
}
