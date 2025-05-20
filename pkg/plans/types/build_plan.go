package plantypes

import "github.com/powertoolsdev/mono/pkg/plugins/configs"

type BuildPlan struct {
	ComponentID      string `json:"component_id"`
	ComponentBuildID string `json:"component_build_id"`

	Src *GitSource `json:"git_source"`

	Dst    *configs.OCIRegistryRepository `json:"dst_registry" validate:"required"`
	DstTag string                         `json:"dst_tag" validate:"required"`

	HelmBuildPlan          *HelmBuildPlan          `json:"helm_build_plan,omitempty"`
	TerraformBuildPlan     *TerraformBuildPlan     `json:"terraform_build_plan,omitempty"`
	DockerBuildPlan        *DockerBuildPlan        `json:"docker_build_plan,omitempty"`
	ContainerImagePullPlan *ContainerImagePullPlan `json:"container_image_pull_plan,omitempty"`
}
