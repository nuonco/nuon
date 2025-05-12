package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/kube"
)

type HelmValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,optional"`
}

type HelmDeployPlan struct {
	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

	// NOTE(jm): these fields should probably just come from the app config, however we keep them around for
	// debuggability
	Name            string `json:"name,attr"`
	Namespace       string `json:"namespace"`
	CreateNamespace bool   `json:"create_namespace"`
	StorageDriver   string `json:"storage_driver"`

	ValuesFiles []string    `json:"values_files"`
	Values      []HelmValue `json:"values"`
}
