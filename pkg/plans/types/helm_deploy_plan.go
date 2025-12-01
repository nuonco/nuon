package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/kube"
)

type HelmValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,optional"` //nolint:staticcheck // SA5008: custom optional tag
}

type HelmDeployPlan struct {
	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"` //nolint:staticcheck // SA5008: custom block tag

	// NOTE(jm): these fields should probably just come from the app config, however we keep them around for
	// debuggability
	Name            string `json:"name,attr"` //nolint:staticcheck // SA5008: custom attr tag
	Namespace       string `json:"namespace"`
	CreateNamespace bool   `json:"create_namespace"`
	StorageDriver   string `json:"storage_driver"`
	HelmChartID     string `json:"helm_chart_id"`

	ValuesFiles   []string    `json:"values_files"`
	Values        []HelmValue `json:"values"`
	TakeOwnership bool        `json:"take_ownership"`
}
