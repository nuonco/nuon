package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/types/state"
)

type HelmValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,optional"`
}

type HelmDeployPlan struct {
	State *state.State `json:"state"`

	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

	// NOTE(jm): these fields should probably just come from the app config, however we keep them around for
	// debuggability
	Name            string `json:"name,attr"`
	Namespace       string `json:"namespace"`
	CreateNamespace bool   `json:"create_namespace"`

	ValuesFiles []string    `json:"values_files"`
	Values      []HelmValue `json:"values"`
}
