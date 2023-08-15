package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ComponentConfigConnection struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	ComponentID string    `json:"component_id"`
	Component   Component `json:"-"`

	ComponentBuilds []ComponentBuild `json:"-"`

	TerraformModuleComponentConfig *TerraformModuleComponentConfig `json:"terraform_module,omitempty"`
	HelmComponentConfig            *HelmComponentConfig            `json:"helm,omitempty"`
	ExternalImageComponentConfig   *ExternalImageComponentConfig   `json:"external_image,omitempty"`
	DockerBuildComponentConfig     *DockerBuildComponentConfig     `json:"docker_build,omitempty"`
}

func (c *ComponentConfigConnection) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	return nil
}
