package config

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type ActionConfig struct {
	// Deprecated
	Source string `mapstructure:"source,omitempty"`

	Name     string                 `mapstructure:"name" jsonschema:"required"`
	Timeout  string                 `mapstructure:"timeout,omitempty"`
	Triggers []*ActionTriggerConfig `mapstructure:"triggers" jsonschema:"required"`
	Steps    []*ActionStepConfig    `mapstructure:"steps" jsonschema:"required"`
}

type ActionTriggerConfig struct {
	Type string `mapstructure:"type" jsonschema:"required"`

	CronSchedule  string `mapstructure:"cron_schedule,omitempty"`
	ComponentName string `mapstructure:"component_name,omitempty"`
}

type ActionStepConfig struct {
	Name          string               `mapstructure:"name" jsonschema:"required"`
	EnvVarMap     map[string]string    `mapstructure:"env_vars,omitempty"`
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"`

	Command        string `mapstructure:"command" features:"template"`
	InlineContents string `mapstructure:"inline_contents" features:"get,template"`
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

	if a.Timeout != "" {
		_, err := time.ParseDuration(a.Timeout)
		if err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("unable to parse timeout %s", a.Timeout),
				Err:         err,
			}
		}
	}

	if err := mapstructure.Decode(obj, &a); err != nil {
		return err
	}
	return nil
}
