package parse

import (
	"errors"

	"github.com/powertoolsdev/mono/pkg/config"
)

// NOTE(jm): this is only required as a temporary migration path, while the old config syncing exists.
//
// This will be removed and we will pass the `config.AppConfig` into the directory parser once we have time to remove
// the old version.
type ConfigDir struct {
	Branch     *config.AppBranchConfig `name:"branch"`
	Components []*config.Component     `name:"components"`
	Actions    []*config.ActionConfig  `name:"actions"`
	Installs   []*config.Install       `name:"installs"`

	Policies    *config.PoliciesConfig `name:"policies"`
	PoliciesDir []config.AppPolicy     `name:"policies"`

	Secrets    *config.SecretsConfig `name:"secrets"`
	SecretsDir []*config.AppSecret   `name:"secrets"`

	Inputs         *config.AppInputConfig `name:"inputs"`
	InputsDir      []config.AppInput      `name:"inputs"`
	InputGroupsDir []config.AppInputGroup `name:"input_groups"`

	Permissions    *config.PermissionsConfig `name:"permissions"`
	PermissionsDir []*config.AppAWSIAMRole   `name:"permissions"`

	BreakGlass    *config.BreakGlass      `name:"break_glass"`
	BreakGlassDir []*config.AppAWSIAMRole `name:"break_glass"`

	Stack     *config.StackConfig      `name:"stack"`
	Sandbox   *config.AppSandboxConfig `name:"sandbox"`
	Runner    *config.AppRunnerConfig  `name:"runner"`
	Metadata  *config.MetadataConfig   `name:"metadata,required"`
	Installer *config.InstallerConfig  `name:"installer"`
}

func (c *ConfigDir) getPolicies() (*config.PoliciesConfig, error) {
	if c.Policies == nil && len(c.PoliciesDir) < 1 {
		return nil, nil
	}
	if c.Policies != nil && len(c.PoliciesDir) > 0 {
		return nil, ParseErr{
			Description: "Can not provide policies both with a policies.toml and policies/ directory",
			Err:         errors.New("Can not provide policies both with a policies.toml and policies/ directory"),
		}
	}

	if c.Policies != nil {
		return c.Policies, nil
	}

	return &config.PoliciesConfig{
		Policies: c.PoliciesDir,
	}, nil
}

func (c *ConfigDir) getSecrets() (*config.SecretsConfig, error) {
	if c.Secrets == nil && c.SecretsDir == nil {
		return nil, nil
	}
	if c.Secrets != nil && c.SecretsDir != nil {
		return nil, ParseErr{
			Description: "Can not provide secrets both with a secrets.toml and secrets/ directory",
			Err:         errors.New("Can not provide secrets both with a secrets.toml and secrets/ directory"),
		}
	}

	if c.Secrets != nil {
		return c.Secrets, nil
	}

	return &config.SecretsConfig{
		Secrets: c.SecretsDir,
	}, nil
}

func (c *ConfigDir) getInputs() (*config.AppInputConfig, error) {
	if c.Inputs == nil && len(c.InputsDir) < 1 && (len(c.InputGroupsDir) < 1 && len(c.InputsDir) < 1) {
		return nil, nil
	}
	if c.Inputs != nil && (len(c.InputsDir) > 0 || len(c.InputGroupsDir) > 0) {
		return nil, ParseErr{
			Description: "Can not provide inputs both with a inputs.toml and inputs/ directory",
			Err:         errors.New("Can not provide inputs both with a inputs.toml and inputs/ directory"),
		}
	}

	if c.Inputs != nil {
		return c.Inputs, nil
	}

	return &config.AppInputConfig{
		Inputs: c.InputsDir,
		Groups: c.InputGroupsDir,
	}, nil
}

func (c *ConfigDir) getPermissions() (*config.PermissionsConfig, error) {
	if c.Permissions == nil && len(c.PermissionsDir) < 1 {
		return nil, nil
	}

	if c.Permissions != nil && len(c.PermissionsDir) > 0 {
		return nil, ParseErr{
			Description: "Can not provide permissions both with a permissions.toml and permissions/ directory",
			Err:         errors.New("Can not provide permissions both with a permissions.toml and permissions/ directory"),
		}
	}

	if c.Permissions != nil {
		return c.Permissions, nil
	}

	return &config.PermissionsConfig{
		Roles: c.PermissionsDir,
	}, nil
}

func (c *ConfigDir) toAppConfig() (*config.AppConfig, error) {
	permissions, err := c.getPermissions()
	if err != nil {
		return nil, err
	}

	secrets, err := c.getSecrets()
	if err != nil {
		return nil, err
	}
	inputs, err := c.getInputs()
	if err != nil {
		return nil, err
	}
	policies, err := c.getPolicies()
	if err != nil {
		return nil, err
	}

	cfg := &config.AppConfig{
		Components:  c.Components,
		Actions:     c.Actions,
		Installs:    c.Installs,
		BreakGlass:  c.BreakGlass,
		Secrets:     secrets,
		Branch:      c.Branch,
		Inputs:      inputs,
		Installer:   c.Installer,
		Sandbox:     c.Sandbox,
		Runner:      c.Runner,
		Permissions: permissions,
		Stack:       c.Stack,
		Policies:    policies,
	}

	if c.Metadata != nil {
		cfg.Version = c.Metadata.Version
		cfg.Description = c.Metadata.Description
		cfg.DisplayName = c.Metadata.DisplayName
		cfg.SlackWebhookURL = c.Metadata.SlackWebhookURL
		cfg.Readme = c.Metadata.Readme
	}

	return cfg, nil
}
