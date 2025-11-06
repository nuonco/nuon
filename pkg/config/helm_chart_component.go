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
	Source   string `toml:"source" mapstructure:"source,omitempty" features:"get,template"`
	Contents string `toml:"contents" mapstructure:"contents,omitempty" features:"get,template"`
	Path     string `toml:"path" mapstructure:"path,omitempty" features:"get"`
}

func (h HelmValuesFile) JSONSchemaExtend(schema *jsonschema.Schema) {
	markDeprecated(schema, "source", "use 'path' instead")
	addDescription(schema, "path", "Path to the values file. Supports loading files via https://github.com/hashicorp/go-getter.")
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type HelmChartComponentConfig struct {
	ChartName string `mapstructure:"chart_name,omitempty" jsonschema:"required"`

	ValuesMap   map[string]string `mapstructure:"values,omitempty"`
	ValuesFiles []HelmValuesFile  `mapstructure:"values_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" jsonschema:"oneof_required=connected_repo"`

	Namespace     string `mapstructure:"namespace" features:"template"`
	StorageDriver string `mapstructure:"storage_driver" features:"template"`

	TakeOwnership bool `mapstructure:"take_ownership" features:"template"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	// deprecated
	Values []HelmValue `mapstructure:"value,omitempty"`
}

func (a HelmChartComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "release name")
	addDescription(schema, "storage_driver", "which helm storage driver to use (defaults to secrets)")
	addDescription(schema, "namespace", "namespace to deploy into. Defaults to {{.nuon.install.id}} and supports templating.")
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
		// Prefer Path over Source (Source is deprecated)
		sourceToUse := valuesFile.Path
		if sourceToUse == "" {
			sourceToUse = valuesFile.Source
		}

		if sourceToUse == "" {
			continue
		}

		byts, err := source.ReadSource(sourceToUse)
		if err != nil {
			return ErrConfig{
				Description: "error loading values file " + sourceToUse,
				Err:         err,
			}
		}

		h.ValuesFiles[idx].Contents = string(byts)
	}

	return nil
}

func (t *HelmChartComponentConfig) Validate() error {
	return nil
}
