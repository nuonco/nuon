package config

import (
	"github.com/invopop/jsonschema"

	"github.com/powertoolsdev/mono/pkg/config/source"
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
	MinComponent

	Name         string   `mapstructure:"name" jsonschema:"required"`
	Dependencies []string `mapstructure:"dependencies"`
	ChartName    string   `mapstructure:"chart_name,omitempty" jsonschema:"required"`

	ValuesMap   map[string]string `mapstructure:"values"`
	ValuesFiles []HelmValuesFile  `mapstructure:"values_file"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"  jsonschema:"oneof_required=connected_repo"`

	// deprecated
	Values []HelmValue `mapstructure:"value"`
}

func (a HelmChartComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "release name")
	addDescription(schema, "dependencies", "additional component dependencies")
	addDescription(schema, "chart_name", "chart name")
	addDescription(schema, "value", "array of helm values (not recommended)")
	addDescription(schema, "values", "map of helm values")
	addDescription(schema, "public_repo", "public repo with the helm chart")
	addDescription(schema, "connected_repo", "connected repo with the helm chart")
}

func (h *HelmChartComponentConfig) Parse() error {
	if len(h.Values) > 0 {
		return ErrConfig{
			Description: "the value array is deprecated, please use values instead.",
		}
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
