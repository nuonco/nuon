package config

import (
	"github.com/invopop/jsonschema"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type HelmValue struct {
	Name  string `toml:"name" mapstructure:"name,omitempty"`
	Value string `toml:"value" mapstructure:"value,omitempty"`
}

func (h HelmValue) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("helm value name").
		Field("value").Short("helm value")
}

type HelmValuesFile struct {
	Source   string `toml:"source" mapstructure:"source,omitempty" features:"get,template"`
	Contents string `toml:"contents" mapstructure:"contents,omitempty" features:"get,template"`
	Path     string `toml:"path" mapstructure:"path,omitempty" features:"get"`
}

func (h HelmValuesFile) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Deprecated("use 'path' instead").OneOfRequired("source").
		Field("path").Short("Path to the values file.").OneOfRequired("path").
		Field("contents").Short("Contents of the values file.").OneOfRequired("contents")
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type HelmChartComponentConfig struct {
	ChartName string `mapstructure:"chart_name,omitempty" jsonschema:"required"`

	ValuesMap   map[string]string `mapstructure:"values,omitempty"`
	ValuesFiles []HelmValuesFile  `mapstructure:"values_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	HelmRepo      *HelmRepoConfig      `mapstructure:"helm_repo,omitempty" jsonschema:"oneof_required=helm_repo"`

	Namespace     string `mapstructure:"namespace" features:"template"`
	StorageDriver string `mapstructure:"storage_driver,omitempty" features:"template"`

	TakeOwnership bool `mapstructure:"take_ownership" features:"template"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	// deprecated
	Values []HelmValue `mapstructure:"value,omitempty"`
}

func (a HelmChartComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("source path or URL").
		Long("Optional source path or URL for the component configuration. Supports HTTP(S) URLs, git repositories, file paths, and relative paths (./). Examples: https://example.com/config.yaml, git::https://github.com/org/repo//config.yaml, file:///path/to/config.yaml, ./local/config.yaml").
		Field("type").Short("component type").
		Field("name").Short("release name").
		Field("var_name").Short("variable name for component output").
		Long("Optional name to use when storing component outputs as variables. If not specified, uses the component name").
		Field("dependencies").Short("additional component dependencies").
		Field("storage_driver").Short("which helm storage driver to use (defaults to configmap)").Enum("configmap", "secret").
		Field("namespace").Short("namespace to deploy into. Defaults to {{.nuon.install.id}} and supports templating.").
		Field("chart_name").Short("chart name").Required().
		Field("value").Short("array of helm values (not recommended)").
		Field("values").Short("map of helm values").
		Field("public_repo").Short("public repo with the helm chart").OneOfRequired("public_repo").
		Field("connected_repo").Short("connected repo with the helm chart").OneOfRequired("connected_repo").
		Field("helm_repo").Short("helm repo config").OneOfRequired("helm_repo")
}

type HelmRepoConfig struct {
	RepoURL string `mapstructure:"repo_url" jsonschema:"required"`
	Chart   string `mapstructure:"chart" jsonschema:"required"`
	Version string `mapstructure:"version,omitempty"`
}

func (h HelmRepoConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("repo_url").Short("URL of the helm chart repository").Example("https://prometheus-community.github.io/helm-charts").Required().
		Field("chart").Short("name of the chart in the repository").Example("kube-prometheus-stack").Required().
		Field("version").Short("version of the chart to use").Example("79.4.1")
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
