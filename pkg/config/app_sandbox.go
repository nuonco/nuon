package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppSandboxConfig struct {
	Source string `mapstructure:"source,omitempty"`

	TerraformVersion string               `mapstructure:"terraform_version" jsonschema:"required"`
	ConnectedRepo    *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	PublicRepo       *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	DriftSchedule    *string              `mapstructure:"drift_schedule,omitempty"`

	EnvVarMap      map[string]string        `mapstructure:"env_vars,omitempty"`
	VarsMap        map[string]string        `mapstructure:"vars,omitempty"`
	VariablesFiles []TerraformVariablesFile `mapstructure:"var_file,omitempty"`
	References     []refs.Ref               `mapstructure:"-" jsonschema:"-" nuonhash:"-"`
}

func (a *AppSandboxConfig) parse() error {
	if a == nil {
		return ErrConfig{
			Description: "an app sandbox config is required",
		}
	}

	references, err := refs.Parse(a)
	if err != nil {
		return errors.Wrap(err, "unable to parse components")
	}
	a.References = references

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

	if a.ConnectedRepo == nil && a.PublicRepo == nil {
		return ErrConfig{
			Description: "either a public repo or connected repo block must be set",
		}
	}
	return nil
}
