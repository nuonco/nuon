package config

import (
	"errors"

	"github.com/invopop/jsonschema"
)

type KubernetesManifestComponentConfig struct {
	// Inline manifest (mutually exclusive with Kustomize)
	Manifest string `mapstructure:"manifest,omitempty" toml:"manifest,omitempty"  features:"get,template"`

	// Kustomize configuration (mutually exclusive with Manifest)
	Kustomize *KustomizeConfig `mapstructure:"kustomize,omitempty" toml:"kustomize"`

	// VCS configuration for kustomize sources (similar to Helm chart)
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty"`

	// Namespace supports template variables (e.g., {{.nuon.install.id}})
	Namespace     string  `mapstructure:"namespace,omitempty" toml:"namespace,omitempty" jsonschema:"required"`
	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" toml:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`

	MaxAutoRetries               *int  `mapstructure:"max_auto_retries,omitempty" toml:"max_auto_retries,omitempty" nuonhash:"omitempty"`
	SkipNoops                    *bool `mapstructure:"skip_noops,omitempty" toml:"skip_noops,omitempty" nuonhash:"omitempty"`
	AutoApproveOnPoliciesPassing *bool `mapstructure:"auto_approve_on_policies_passing,omitempty" toml:"auto_approve_on_policies_passing,omitempty" nuonhash:"omitempty"`
}

// KustomizeConfig configures kustomize build options
type KustomizeConfig struct {
	// Path to kustomization directory (relative to source root)
	Path string `mapstructure:"path" jsonschema:"required" toml:"path"`

	// Additional patch files to apply after kustomize build
	Patches []string `mapstructure:"patches,omitempty" toml:"patches,omitempty"`

	// Enable Helm chart inflation during kustomize build
	EnableHelm bool `mapstructure:"enable_helm,omitempty" toml:"enable_helm,omitempty"`

	// Load restrictor: none, rootOnly (default: rootOnly)
	LoadRestrictor string `mapstructure:"load_restrictor,omitempty" toml:"load_restrictor,omitempty"`
}

func (k KustomizeConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("path").Short("kustomization directory path").
		Long("Path to the kustomization directory, relative to the source root.").
		Example("overlays/production").Example(".").
		Field("patches").Short("additional patch files").
		Long("Additional patch files to apply after kustomize build.").
		Example("patches/namespace.yaml").
		Field("enable_helm").Short("enable Helm chart inflation").
		Long("Enable Helm chart inflation during kustomize build.").
		Field("load_restrictor").Short("file load restrictor").
		Long("Controls how kustomize loads files. Options: none, rootOnly. Default: rootOnly.").
		Example("rootOnly").Example("none")
}

func (k KubernetesManifestComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("manifest").Short("Kubernetes manifest").
		Long("Path to a YAML manifest file or inline manifest content for Kubernetes resources. Supports templating with variables like {{.nuon.install.id}}. Mutually exclusive with kustomize").
		Example("./manifests/deployment.yaml").
		OneOfRequired("manifest_source").
		Field("kustomize").Short("Kustomize configuration").
		Long("Configuration for building manifests from a kustomize overlay. Mutually exclusive with manifest. Requires either public_repo or connected_repo").
		OneOfRequired("manifest_source").
		Field("public_repo").Short("public repository with kustomize source").
		Long("Configuration for a public Git repository containing kustomize overlays. Only valid when using kustomize, not inline manifests").
		Field("connected_repo").Short("connected repository with kustomize source").
		Long("Configuration for a Nuon-connected private repository containing kustomize overlays. Only valid when using kustomize, not inline manifests").
		Field("namespace").Short("Kubernetes namespace").
		Long("Kubernetes namespace where the manifest will be deployed. Supports template variables").
		Example("default").
		Example("clickhouse").
		Example("{{.nuon.install.id}}").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled").
		Example("0 2 * * *").
		Field("build_timeout").Short("build operation timeout").
		Long("Duration string for build operations (e.g., \"30m\", \"1h\"). Default: 5m. Max: 1h").
		Default("5m").
		Example("30m").
		Example("1h").
		Field("deploy_timeout").Short("deploy operation timeout").
		Long("Duration string for deploy operations (e.g., \"30m\", \"1h\"). Default: 15m. Max: 1h").
		Default("15m").
		Example("30m").
		Example("1h").
		Field("max_auto_retries").Short("maximum automatic retry attempts on deploy failure").
		Long("Maximum number of automatic retry attempts for failed deployments. Set to 0 to disable auto-retry. Default: 0 (disabled)").
		Default("0").
		Example("3").
		Example("5")
}

func (t *KubernetesManifestComponentConfig) Validate() error {
	// Exactly one of manifest or kustomize must be set
	hasManifest := t.Manifest != ""
	hasKustomize := t.Kustomize != nil

	if !hasManifest && !hasKustomize {
		return errors.New("one of 'manifest' or 'kustomize' must be specified")
	}
	if hasManifest && hasKustomize {
		return errors.New("only one of 'manifest' or 'kustomize' can be specified")
	}

	// Validate kustomize config
	if t.Kustomize != nil {
		if t.Kustomize.Path == "" {
			return errors.New("kustomize.path is required")
		}
		// Kustomize requires a VCS source
		if t.PublicRepo == nil && t.ConnectedRepo == nil {
			return errors.New("kustomize requires either 'public_repo' or 'connected_repo' to be specified")
		}
	}

	// VCS config should only be set with kustomize
	if !hasKustomize && (t.PublicRepo != nil || t.ConnectedRepo != nil) {
		return errors.New("'public_repo' and 'connected_repo' are only valid with kustomize, not inline manifests")
	}

	// Only one VCS source can be specified
	if t.PublicRepo != nil && t.ConnectedRepo != nil {
		return errors.New("only one of 'public_repo' or 'connected_repo' can be specified")
	}

	return nil
}

func (k *KubernetesManifestComponentConfig) Parse() error {
	return nil
}
