package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppInput struct {
	Name        string `mapstructure:"name,omitempty" toml:"name"`
	Description string `mapstructure:"description,omitempty" toml:"description"`
	Group       string `mapstructure:"group,omitempty" toml:"group"`
	Default     string `mapstructure:"default" toml:"default"`
	Required    bool   `mapstructure:"required" toml:"required"`
	DisplayName string `mapstructure:"display_name,omitempty" toml:"display_name"`
	Sensitive   bool   `mapstructure:"sensitive" toml:"sensitive"`
}

type AppInputGroup struct {
	Name        string `mapstructure:"name,omitempty" toml:"name"`
	Description string `mapstructure:"description,omitempty" toml:"description"`
	DisplayName string `mapstructure:"display_name,omitempty" toml:"display_name"`
}

type AppInputConfig struct {
	Inputs []AppInput      `mapstructure:"input,omitempty" toml:"input"`
	Groups []AppInputGroup `mapstructure:"group,omitempty" toml:"group"`

	Source  string   `mapstructure:"source,omitempty"`
	Sources []string `mapstructure:"sources,omitempty"`
}

func (a *AppInputConfig) ToResourceType() string {
	return "nuon_app_input"
}

func (a *AppInputConfig) ToResource() (map[string]interface{}, error) {
	if err := a.parse(ConfigContextSource); err != nil {
		return nil, fmt.Errorf("error parsing app input config: %w", err)
	}

	resource, err := toMapStructure(a)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"

	delete(resource, "source")
	delete(resource, "sources")

	return nestWithName("input", resource), nil
}

func (a *AppInputConfig) parse(ctx ConfigContext) error {
	if ctx != ConfigContextSource {
		return nil
	}

	sources := make([]string, 0)
	if a.Source != "" {
		sources = append(sources, a.Source)
	}
	sources = append(sources, a.Sources...)

	for _, src := range sources {
		obj, err := source.LoadSource(src)
		if err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("unable to load source %s", src),
				Err:         err,
			}
		}

		var inpCfg AppInputConfig
		if err := mapstructure.Decode(obj, &inpCfg); err != nil {
			return fmt.Errorf("unable to parse input source %s: %w", src, err)
		}

		a.Inputs = append(a.Inputs, inpCfg.Inputs...)
		a.Groups = append(a.Groups, inpCfg.Groups...)
	}

	return nil
}
