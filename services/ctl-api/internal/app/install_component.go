package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallComponent struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	InstallID   string    `json:"install_id" gorm:"index:install_component_group,unique;notnull"`
	Install     Install   `faker:"-" json:"-"`
	ComponentID string    `json:"component_id" gorm:"index:install_component_group,unique;notnull"`
	Component   Component `faker:"-" json:"component"`

	InstallDeploys []InstallDeploy `faker:"-" gorm:"constraint:OnDelete:CASCADE;" json:"install_deploys"`

	// after query fields filled in after querying
	Status InstallDeployStatus `json:"status" gorm:"-" swaggertype:"string"`
}

func (c *InstallComponent) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *InstallComponent) AfterQuery(tx *gorm.DB) error {
	c.Status = InstallDeployStatusUnknown
	if len(c.InstallDeploys) > 0 {
		c.Status = c.InstallDeploys[0].Status
	}

	return nil
}
