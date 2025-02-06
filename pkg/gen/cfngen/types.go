package cfngen

import (
	"fmt"

	"github.com/iancoleman/strcase"
)

// AppConfigValues are values fed to the cloudformation generator that are expected to originate from
// a vendor's app configuration.
type AppConfigValues struct {
	// BYOVPC indicates whether the installer should be generated for a pre-existing VPC (true), or to create a new one (false)
	BYOVPC bool `toml:"byovpc"`
	// InstanceType is the EC2 instance type to use for the runner. Defaults to t3a.medium
	InstanceType string `toml:"instance_type"`
	// AdditionalTags are additional tags to apply to all resources in the stack.
	AdditionalTags map[string]string `toml:"additional_tags"`
	// Roles is the set of user-specified roles. Up to four are expected - provision, deprovision, maintenance, and breakglass.
	Roles []*RoleConfig `toml:"roles"`
	// Secrets is an arbitrary set of secrets that the end customer will provide as part of stack creation.
	Secrets []*SecretConfig `toml:"secrets"`
}

type RoleConfig struct {
	Name        string         `toml:"name"`
	Description string         `toml:"description"`
	DisplayName string         `toml:"display_name"`
	Policies    []PolicyConfig `toml:"policies"`
}

func (s RoleConfig) KnownRoleName() string {
	switch s.Name {
	case "provision":
		return "Provision"
	case "deprovision":
		return "Deprovision"
	case "maintenance":
		return "Maintenance"
	case "breakglass":
		return "BreakGlass"
	default:
		panic(fmt.Sprintf("unknown role name %s", s.Name))
	}
}

func (s RoleConfig) DefaultParam() string {
	if s.Name == "provision" {
		return "true"
	}
	return "false"
}

func (s RoleConfig) ParamName() string {
	return fmt.Sprintf("Enable%s", s.KnownRoleName())
}

func (s RoleConfig) StrRole() string {
	return fmt.Sprintf("%sRole", s.KnownRoleName())
}

func (s RoleConfig) StrPolicy() string {
	return fmt.Sprintf("%sPolicy", s.KnownRoleName())
}

func (s RoleConfig) StrEnabled() string {
	return fmt.Sprintf("%sEnabled", s.KnownRoleName())
}

type PolicyConfig struct {
	Builtin             bool   `toml:"builtin,omitempty"`
	Name                string `toml:"name"`
	PolicyJSON          string `toml:"policy_json,omitempty"`
	PermissionsBoundary string `toml:"permissions_boundary,omitempty"`
}

type SecretConfig struct {
	Name             string `toml:"name"`
	Required         bool   `toml:"required"`
	DisplayName      string `toml:"display_name"`
	Description      string `toml:"description"`
	KubernetesSecret string `toml:"kubernetes_secret"`
}

func (s SecretConfig) CamelSecret() string {
	return strcase.ToCamel(s.Name)
}
