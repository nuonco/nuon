package plan

import "github.com/powertoolsdev/mono/pkg/diff"

type KubernetesManifestPlanOperation string

const (
	KubernetesManifestPlanOperationApply  KubernetesManifestPlanOperation = "apply"
	KubernetesManifestPlanOperationDelete KubernetesManifestPlanOperation = "delete"
)

// KubernetesManifestPlanContents for kubernetes plan, summarized before after state of all resources
type KubernetesManifestPlanContents struct {
	Plan        string                          `json:"plan"`
	Op          KubernetesManifestPlanOperation `json:"op"`
	ContentDiff []diff.ResourceDiff             `json:"k8s_content_diff,omitempty"`
}
