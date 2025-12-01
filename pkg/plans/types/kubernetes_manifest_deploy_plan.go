package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/kube"
)

type KubernetesManifestDeployPlan struct {
	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"` //nolint:staticcheck // SA5008: custom block tag

	Namespace string `json:"namespace"`
	Manifest  string `json:"manifest"`
}
