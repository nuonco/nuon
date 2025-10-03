package config

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type ActionConfig struct {
	Name     string                 `mapstructure:"name" jsonschema:"required"`
	Timeout  string                 `mapstructure:"timeout,omitempty"`
	Triggers []*ActionTriggerConfig `mapstructure:"triggers" jsonschema:"required"`
	Steps    []*ActionStepConfig    `mapstructure:"steps" jsonschema:"required"`

	References     []refs.Ref `mapstructure:"-" jsonschema:"-"`
	Dependencies   []string   `mapstructure:"dependencies,omitempty"`
	BreakGlassRole string     `mapstructure:"break_glass_role,omitempty"`
}

type ActionTriggerConfig struct {
	Type string `mapstructure:"type" jsonschema:"required"`

	Index         int64  `mapstructure:"index,omitempty"`
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

	// created during parsing
	References []refs.Ref `mapstructure:"-" jsonschema:"-"`
}

func (a *ActionConfig) parse() error {
	if a == nil {
		return nil
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

	references, err := refs.Parse(a)
	if err != nil {
		return errors.Wrap(err, "unable to parse components")
	}
	a.References = references

	for _, ref := range a.References {
		if !generics.SliceContains(ref.Type, []refs.RefType{refs.RefTypeComponents}) {
			continue
		}

		a.Dependencies = append(a.Dependencies, ref.Name)
	}
	a.Dependencies = generics.UniqueSlice(a.Dependencies)
	return nil
}
