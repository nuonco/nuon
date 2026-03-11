package plantypes

import (
	"github.com/nuonco/nuon/pkg/kube"
)

type KubernetesManifestDeployPlan struct {
	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

	Namespace string `json:"namespace"`

	// Manifest is populated at runtime from the OCI artifact.
	// This field is no longer set during plan creation - it's populated by the runner
	// after pulling the OCI artifact during Initialize().
	Manifest string `json:"manifest,omitempty"`

	// OCIArtifact reference (set during plan creation, used by runner to pull manifest)
	OCIArtifact *OCIArtifactReference `json:"oci_artifact,omitempty"`
}

// OCIArtifactReference points to the packaged manifest artifact
type OCIArtifactReference struct {
	// URL is the full artifact URL (e.g., registry.nuon.co/org_id/app_id)
	URL string `json:"url"`

	// Tag is the artifact tag (typically the build ID)
	Tag string `json:"tag,omitempty"`

	// Digest is the immutable artifact digest (e.g., sha256:abc123...)
	Digest string `json:"digest,omitempty"`
}
