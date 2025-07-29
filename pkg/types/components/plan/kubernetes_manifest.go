package plan

import "k8s.io/apimachinery/pkg/runtime/schema"

type KubernetesManifestPlanOperation string

const (
	KubernetesManifestPlanOperationApply  KubernetesManifestPlanOperation = "apply"
	KubernetesManifestPlanOperationDelete KubernetesManifestPlanOperation = "delete"
)

type KubernetesManifestDiff struct {
	GroupVersionKind     schema.GroupVersionKind         `json:"group_version_kind,omitempty"`
	GroupVersionResource schema.GroupVersionResource     `json:"group_version_resource,omitempty"`
	Namespace            string                          `json:"namespace,omitempty"`
	Name                 string                          `json:"name,omitempty"`
	Diff                 string                          `json:"diff,omitempty"`
	Before               string                          `json:"before,omitempty"`
	After                string                          `json:"after,omitempty"`
	Op                   KubernetesManifestPlanOperation `json:"op,omitempty"`
}

// KubernetesManifestPlanContents for kubernetes plan, summarized before after state of all resources
type KubernetesManifestPlanContents struct {
	Plan []*KubernetesManifestDiff `json:"plan"`
}
