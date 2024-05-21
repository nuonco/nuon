package config

import (
	"github.com/powertoolsdev/mono/pkg/config/source"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type HelmValue struct {
	Name  string `toml:"name" mapstructure:"name,omitempty"`
	Value string `toml:"value" mapstructure:"value,omitempty"`
}

type HelmValuesFile struct {
	Source   string `toml:"source" mapstructure:"source,omitempty"`
	Contents string `toml:"contents" mapstructure:"contents,omitempty"`
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type HelmChartComponentConfig struct {
	Name         string            `mapstructure:"name,omitempty"`
	Dependencies []string          `mapstructure:"dependencies,omitempty"`
	ChartName    string            `mapstructure:"chart_name,omitempty"`
	Values       []HelmValue       `mapstructure:"value,omitempty"`
	ValuesMap    map[string]string `mapstructure:"values,omitempty"`

	ValuesFiles []HelmValuesFile `mapstructure:"values_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"`
}

func (h *HelmChartComponentConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(h)
	if err != nil {
		return nil, err
	}
	resource["app_id"] = "${var.app_id}"
	delete(resource, "values")

	generics.DeleteSliceKey(resource["values_file"], "source")

	return resource, nil
}

func (h *HelmChartComponentConfig) parse(ctx ConfigContext) error {
	for k, v := range h.ValuesMap {
		h.Values = append(h.Values, HelmValue{
			Name:  k,
			Value: v,
		})
	}
	if ctx == ConfigContextConfigOnly {
		return nil
	}

	for idx, valuesFile := range h.ValuesFiles {
		if valuesFile.Source == "" {
			continue
		}

		byts, err := source.ReadSource(valuesFile.Source)
		if err != nil {
			return ErrConfig{
				Description: "error loading values file " + valuesFile.Source,
				Err:         err,
			}
		}

		h.ValuesFiles[idx].Contents = string(byts)
	}

	return nil
}
