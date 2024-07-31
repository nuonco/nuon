package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppSandboxConfig struct {
	Source string `mapstructure:"source,omitempty"`

	TerraformVersion string               `mapstructure:"terraform_version" jsonschema:"required"`
	ConnectedRepo    *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	PublicRepo       *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`

	VarMap map[string]string `mapstructure:"vars,omitempty"`

	AWSDelegationIAMRoleARN string `mapstructure:"aws_delegation_iam_role_arn,omitempty"`

	// Deprecated
	Vars []TerraformVariable `mapstructure:"var,omitempty" toml:"var"`
}

func (a *AppSandboxConfig) parse() error {
	if len(a.Vars) > 0 {
		return ErrConfig{
			Description: "the var array is deprecated, please use vars instead",
		}
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
