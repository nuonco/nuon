package parse

import (
	"github.com/powertoolsdev/mono/pkg/config"
)

// NOTE(jm): this is only required as a temporary migration path, while the old config syncing exists.
//
// This will be removed and we will pass the `config.AppConfig` into the directory parser once we have time to remove
// the old version.
type ConfigDir struct {
	Components []*config.Component    `name:"components,nonempty"`
	Actions    []*config.ActionConfig `name:"actions,nonempty"`

	BreakGlass *config.BreakGlass      `name:"break_glass,nonempty"`
	Installer  *config.InstallerConfig `name:"installer,nonempty"`
	Policies   *config.PoliciesConfig  `name:"policies,nonempty"`
	Secrets    *config.SecretsConfig   `name:"secrets,nonempty"`
	Inputs     *config.AppInputConfig  `name:"inputs,nonempty"`

	CloudFormationStack *config.CloudformationStackConfig `name:"cloudformation_stack,nonempty"`
	Sandbox             *config.AppSandboxConfig          `name:"sandbox,required,nonempty"`
	Runner              *config.AppRunnerConfig           `name:"runner,required,nonempty"`
	Metadata            *config.MetadataConfig            `name:"metadata,required,nonempty"`
	Permissions         *config.PermissionsConfig         `name:"permissions,required,nonempty"`
}

func (c *ConfigDir) toAppConfig() (*config.AppConfig, error) {
	cfg := &config.AppConfig{
		Components:          c.Components,
		Actions:             c.Actions,
		BreakGlass:          c.BreakGlass,
		Policies:            c.Policies,
		Secrets:             c.Secrets,
		Inputs:              c.Inputs,
		Installer:           c.Installer,
		Sandbox:             c.Sandbox,
		Runner:              c.Runner,
		Permissions:         c.Permissions,
		CloudFormationStack: c.CloudFormationStack,

		// Metadata
		Version:         c.Metadata.Version,
		Description:     c.Metadata.Description,
		DisplayName:     c.Metadata.DisplayName,
		SlackWebhookURL: c.Metadata.SlackWebhookURL,
		Readme:          c.Metadata.Readme,
	}

	return cfg, nil
}
