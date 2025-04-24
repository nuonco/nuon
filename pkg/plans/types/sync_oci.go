package plantypes

import "github.com/powertoolsdev/mono/pkg/plugins/configs"

type SyncOCIPlan struct {
	Src    *configs.OCIRegistryRepository `json:"src_registry" validate:"required"`
	SrcTag string                         `json:"src_tag" validate:"required"`

	Dst    *configs.OCIRegistryRepository `json:"dst_registry" validate:"required"`
	DstTag string                         `json:"dst_tag" validate:"required"`
}
