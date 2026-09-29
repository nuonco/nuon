package plan

import "github.com/nuonco/nuon/pkg/diff"

type KubernetesManifestPlanOperation string

const (
	KubernetesManifestPlanOperationApply  KubernetesManifestPlanOperation = "apply"
	KubernetesManifestPlanOperationDelete KubernetesManifestPlanOperation = "delete"
)

type KubernetesManifestPlanContents struct {
	Plan         string                          `json:"plan"`
	Op           KubernetesManifestPlanOperation `json:"op"`
	ContentDiff  []diff.ResourceDiff             `json:"k8s_content_diff,omitempty"`
	DryRunOutput string                          `json:"dry_run_output,omitempty"`
}
