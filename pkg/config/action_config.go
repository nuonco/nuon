package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/config/source"
)

type ActionConfig struct {
	Source   string                 `mapstructure:"source,omitempty"`
	Name     string                 `mapstructure:"name" jsonschema:"required"`
	Triggers []*ActionTriggerConfig `mapstructure:"triggers" jsonschema:"required"`
	Steps    []*ActionStepConfig    `mapstructure:"steps" jsonschema:"required"`
}

type ActionTriggerConfig struct {
	Type         string `mapstructure:"type" jsonschema:"required"`
	CronSchedule string `mapstructure:"cron_schedule,omitempty"`
}

type ActionStepConfig struct {
	Name          string               `mapstructure:"name" jsonschema:"required"`
	EnvVarMap     map[string]string    `mapstructure:"env_vars,omitempty"`
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"  jsonschema:"oneof_required=connected_repo"`
	Command       string               `mapstructure:"command"`
}

func (a *ActionConfig) parse() error {
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
