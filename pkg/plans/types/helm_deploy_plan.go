package plantypes

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
)

type HelmValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,optional"`
}

type HelmDeployPlan struct {
	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

	// Auth for cloud providers
	AWSAuth   *awscredentials.Config   `json:"aws_auth,omitempty"`
	AzureAuth *azurecredentials.Config `json:"azure_auth,omitempty"`
	GCPAuth   *gcpcredentials.Config   `json:"gcp_auth,omitempty"`

	// NOTE(jm): these fields should probably just come from the app config, however we keep them around for
	// debuggability
	Name            string `json:"name,attr"`
	Namespace       string `json:"namespace"`
	CreateNamespace bool   `json:"create_namespace"`
	StorageDriver   string `json:"storage_driver"`
	HelmChartID     string `json:"helm_chart_id"`

	ValuesFiles   []string    `json:"values_files"`
	Values        []HelmValue `json:"values"`
	TakeOwnership bool        `json:"take_ownership"`

	// ValuesOverride is the install-level Helm values override (raw YAML). It is
	// merged as the highest-precedence layer at deploy time, winning over both
	// ValuesFiles and Values. Empty means no override (exact no-op).
	ValuesOverride string `json:"values_override,omitempty"`
}
