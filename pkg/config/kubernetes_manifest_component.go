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
		Field("patches").Short("additional patch files").
		Long("Additional patch files to apply after kustomize build.").
		Field("enable_helm").Short("enable Helm chart inflation").
		Long("Enable Helm chart inflation during kustomize build.").
		Field("load_restrictor").Short("file load restrictor").
		Long("Controls how kustomize loads files. Options: none, rootOnly. Default: rootOnly.")
}

func (k KubernetesManifestComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("manifest").Short("Kubernetes manifest").
		Long("YAML manifest content for Kubernetes resources. Supports templating with variables like {{.nuon.install.id}}").
		Example("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\ndata:\n  env: production").
		OneOfRequired("manifest_source").
		Field("kustomize").Short("Kustomize configuration (mutually exclusive with manifest)").
		Long("Configuration for building manifests from a kustomize overlay. Mutually exclusive with manifest.").
		OneOfRequired("manifest_source").
		Field("namespace").Short("Kubernetes namespace").
		Long("Kubernetes namespace where the manifest will be deployed. Supports template variables.").
		Example("default").
		Example("production").
		Example("{{.nuon.install.id}}").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled.").Example("0 2 * * *")
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
