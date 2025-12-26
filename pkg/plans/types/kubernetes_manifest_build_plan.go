package plantypes

// KubernetesManifestBuildPlan contains build configuration for kubernetes manifest components.
// This is used by the build runner to package manifests into OCI artifacts.
type KubernetesManifestBuildPlan struct {
	// Labels for the OCI artifact
	Labels map[string]string `json:"labels,omitempty"`

	// SourceType indicates how manifests are sourced: "inline" or "kustomize"
	SourceType string `json:"source_type"`

	// InlineManifest contains the raw manifest YAML (for inline source type)
	InlineManifest string `json:"inline_manifest,omitempty"`

	// KustomizePath is the path to the kustomization directory (for kustomize source type)
	// Relative to the repository root
	KustomizePath string `json:"kustomize_path,omitempty"`

	// KustomizeConfig contains additional kustomize build options
	KustomizeConfig *KustomizeBuildConfig `json:"kustomize_config,omitempty"`
}

// KustomizeBuildConfig contains kustomize-specific build options
type KustomizeBuildConfig struct {
	// Patches are additional patch files to apply after kustomize build
	Patches []string `json:"patches,omitempty"`

	// EnableHelm enables Helm chart inflation during kustomize build
	EnableHelm bool `json:"enable_helm,omitempty"`

	// LoadRestrictor controls file loading: "none" or "rootOnly" (default)
	LoadRestrictor string `json:"load_restrictor,omitempty"`
}
