package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/kube"
)

type KubernetesManifestDeployPlan struct {
	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

	Namespace string `json:"namespace"`
	Manifest  string `json:"manifest"`
}
