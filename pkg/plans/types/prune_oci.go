package plantypes

import (
	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type PruneOCIPlan struct {
	Registry *configs.OCIRegistryRepository `json:"registry" validate:"required"`

	DryRun             bool     `json:"dry_run"`
	OwnedTags          []string `json:"owned_tags"`
	KeepTags           []string `json:"keep_tags"`
	GracePeriodSeconds int64    `json:"grace_period_seconds" validate:"gte=0"`
	MaxDeletes         int      `json:"max_deletes" validate:"gt=0"`

	ClusterInfo *kube.ClusterInfo `json:"cluster_info,block" validate:"required"`

	MinSandboxMode
}
