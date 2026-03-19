package config

import (
	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/source"
)

type HelmValue struct {
	Name  string `toml:"name" mapstructure:"name,omitempty"`
	Value string `toml:"value" mapstructure:"value,omitempty"`
}

func (h HelmValue) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("helm value name").Example("replicaCount").Example("image.tag").
		Field("value").Short("helm value").Example("3").Example("{{.nuon.install.id}}")
}

type HelmValuesFile struct {
	Source   string `toml:"source" mapstructure:"source,omitempty" features:"get,template"`
	Contents string `toml:"contents" mapstructure:"contents,omitempty" features:"get,template"`
	Path     string `toml:"path" mapstructure:"path,omitempty" features:"get"`
}

func (h HelmValuesFile) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Deprecated("use 'path' instead").OneOfRequired("source").
		Field("path").Short("Path to the values file.").Example("./values/clickhouse-operator.yaml").Example("./values/coder.yaml").OneOfRequired("path").
		Field("contents").Short("Contents of the values file.").Example("./values/whoami.yaml").OneOfRequired("contents")
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type HelmChartComponentConfig struct {
	ChartName string `mapstructure:"chart_name,omitempty" toml:"chart_name,omitempty" jsonschema:"required"`

	ValuesMap   map[string]string `mapstructure:"values,omitempty" toml:"values,omitempty"`
	ValuesFiles []HelmValuesFile  `mapstructure:"values_file,omitempty" toml:"values_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	HelmRepo      *HelmRepoConfig      `mapstructure:"helm_repo,omitempty" toml:"helm_repo,omitempty" jsonschema:"oneof_required=helm_repo"`

	Namespace     string `mapstructure:"namespace" toml:"namespace" features:"template"`
	StorageDriver string `mapstructure:"storage_driver,omitempty" toml:"storage_driver,omitempty" features:"template"`

	TakeOwnership bool `mapstructure:"take_ownership" toml:"take_ownership" features:"template"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" toml:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`

	// deprecated
	Values []HelmValue `mapstructure:"value,omitempty" toml:"value,omitempty"`
}

func (a HelmChartComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("chart_name").Short("Helm chart name").Required().
		Long("Name of the Helm chart to deploy. Must match the chart name in the repository or Helm repo").
		Example("karpenter-nodepools").
		Example("prometheus").
		Example("cert-manager").
		Field("values").Short("inline Helm values").
		Long("Map of Helm values as key-value pairs. These are passed to helm install/upgrade as --set arguments. Supports Nuon templating").
		Field("values_file").Short("Helm values files").
		Long("Array of external Helm values files to load. Each entry can specify a path to a local file or inline contents. Supports Nuon templating and external file sources").
		Field("public_repo").Short("public repository with the Helm chart").
		Long("Configuration for a public Git repository containing the Helm chart source. Mutually exclusive with connected_repo and helm_repo").
		OneOfRequired("public_repo").
		Field("connected_repo").Short("connected repository with the Helm chart").
		Long("Configuration for a Nuon-connected private repository containing the Helm chart source. Mutually exclusive with public_repo and helm_repo").
		OneOfRequired("connected_repo").
		Field("helm_repo").Short("Helm chart repository").
		Long("Configuration for pulling a chart from a Helm repository (e.g., a public chart registry). Mutually exclusive with public_repo and connected_repo").
		OneOfRequired("helm_repo").
		Field("namespace").Short("Kubernetes namespace to deploy into").
		Long("Kubernetes namespace where the Helm release will be installed. Defaults to {{.nuon.install.id}}. Supports Nuon templating").
		Example("clickhouse").
		Example("monitoring").
		Example("{{.nuon.install.id}}").
		Field("storage_driver").Short("Helm storage driver").
		Long("Backend storage driver for Helm release metadata. Defaults to configmap").
		Enum("configmap", "secret").
		Example("configmap").
		Field("take_ownership").Short("adopt existing Helm releases").
		Long("If true, Nuon will adopt an existing Helm release with the same name and namespace that was not originally managed by Nuon. Useful when migrating existing deployments to Nuon").
		Field("value").Short("deprecated: use values map instead").
		Long("Deprecated: Array of name/value pairs for Helm values. Use the values map instead").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled").
		Example("0 2 * * *").
		Field("build_timeout").Short("build operation timeout").
		Long("Duration string for build operations (e.g., \"30m\", \"1h\"). Default: 5m. Max: 1h").
		Default("5m").
		Example("30m").
		Example("1h").
		Field("deploy_timeout").Short("deploy operation timeout").
		Long("Duration string for deploy operations (e.g., \"30m\", \"1h\"). Default: 30m. Max: 1h").
		Default("30m").
		Example("30m").
		Example("1h")
}

type HelmRepoConfig struct {
	RepoURL string `mapstructure:"repo_url" toml:"repo_url" jsonschema:"required"`
	Chart   string `mapstructure:"chart" toml:"chart" jsonschema:"required"`
	Version string `mapstructure:"version,omitempty" toml:"version,omitempty"`
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
