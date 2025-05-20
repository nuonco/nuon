package plantypes

import (
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

type ContainerImagePullPlan struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`

	RepoCfg *configs.OCIRegistryRepository `json:"repo_config"`
}
