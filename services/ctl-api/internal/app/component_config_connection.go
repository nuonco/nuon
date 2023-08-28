package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ComponentConfigConnection struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull"`

	ComponentID string    `json:"component_id" gorm:"notnull"`
	Component   Component `json:"-"`

	ComponentBuilds []ComponentBuild `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	TerraformModuleComponentConfig *TerraformModuleComponentConfig `json:"terraform_module,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	HelmComponentConfig            *HelmComponentConfig            `json:"helm,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	ExternalImageComponentConfig   *ExternalImageComponentConfig   `json:"external_image,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	DockerBuildComponentConfig     *DockerBuildComponentConfig     `json:"docker_build,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
}

func (c *ComponentConfigConnection) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
