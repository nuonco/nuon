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
	SkipCRDs      bool        `json:"skip_crds,omitempty"`

	// ValuesOverride is the install-level Helm values override (raw YAML). It is
	// merged as the highest-precedence layer at deploy time, winning over both
	// ValuesFiles and Values. Empty means no override (exact no-op).
	ValuesOverride string `json:"values_override,omitempty"`

	// Recover, when set, makes this job unstick a release that helm left in a
	// pending state instead of applying the chart. Nil is the normal deploy.
	//
	// A recovery reads the chart and values from the stored revision, so the
	// runner skips fetching the OCI artifact entirely — requiring it would make
	// recovery fail whenever the artifact is unreachable, which is exactly when
	// an install is most likely to be wedged.
	Recover *HelmRecover `json:"recover,omitempty"`
}

// HelmRecover carries the recovery directive. It is a struct rather than a bool
// so a caller-chosen target revision can be added later without another change
// to the plan shape.
type HelmRecover struct{}
