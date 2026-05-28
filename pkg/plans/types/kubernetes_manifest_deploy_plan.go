package plantypes

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/types/state"
)

type KubernetesManifestDeployPlan struct {
	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

	// Auth for cloud providers
	AWSAuth   *awscredentials.Config   `json:"aws_auth,omitempty"`
	AzureAuth *azurecredentials.Config `json:"azure_auth,omitempty"`
	GCPAuth   *gcpcredentials.Config   `json:"gcp_auth,omitempty"`

	Namespace string `json:"namespace"`

	// Manifest is populated at runtime from the OCI artifact.
	// This field is no longer set during plan creation - it's populated by the runner
	// after pulling the OCI artifact during Initialize().
	Manifest string `json:"manifest,omitempty"`

	// OCIArtifact reference (set during plan creation, used by runner to pull manifest)
	OCIArtifact *OCIArtifactReference `json:"oci_artifact,omitempty"`

	// State carries the install state the runner needs to interpolate
	// nuon template placeholders into the kustomize-produced manifest.yaml
	// after pulling the OCI artifact. Nil for inline-manifest deploys,
	// which are pre-rendered planner-side and short-circuit the OCI pull
	// on the runner. Same shape as SandboxRunPlan.State.
	State *state.State `json:"state,omitempty"`
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
