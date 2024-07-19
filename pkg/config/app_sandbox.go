package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppSandboxConfig struct {
	Source string `mapstructure:"source,omitempty"`

	TerraformVersion string               `mapstructure:"terraform_version,omitempty" toml:"terraform_version"`
	ConnectedRepo    *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo"`
	PublicRepo       *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo"`

	Vars   []TerraformVariable `mapstructure:"var,omitempty" toml:"var"`
	VarMap map[string]string   `mapstructure:"-" toml:"vars"`

	AWSDelegationIAMRoleARN string `mapstructure:"aws_delegation_iam_role_arn" toml:"aws_delegation_iam_role_arn"`
}

func (t *AppSandboxConfig) ToResourceType() string {
	return "nuon_app_sandbox"
}

func (t *AppSandboxConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"
	delete(resource, "source")

	return nestWithName("sandbox", resource), nil
}

func (a *AppSandboxConfig) parse(ctx ConfigContext) error {
	for k, v := range a.VarMap {
		a.Vars = append(a.Vars, TerraformVariable{
			Name:  k,
			Value: v,
		})
	}

	if ctx == ConfigContextConfigOnly {
		return nil
	}
	if a.Source == "" {
		return nil
	}

	obj, err := source.LoadSource(a.Source)
	if err != nil {
		return ErrConfig{
			Description: fmt.Sprintf("unable to load source %s", a.Source),
			Err:         err,
		}
	}

	if err := mapstructure.Decode(obj, &a); err != nil {
		return err
	}
	return nil
}
