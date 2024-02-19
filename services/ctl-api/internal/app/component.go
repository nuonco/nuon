package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Component struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26;" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_component_name,unique" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Name string `json:"name" gorm:"notnull;index:idx_app_component_name,unique"`

	AppID string `json:"app_id" gorm:"notnull;index:idx_app_component_name,unique"`
	App   App    `faker:"-" json:"-"`

	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`

	ConfigVersions    int                         `gorm:"-" json:"config_versions"`
	ComponentConfigs  []ComponentConfigConnection `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	InstallComponents []InstallComponent          `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Dependencies  []*Component `gorm:"many2many:component_dependencies;constraint:OnDelete:CASCADE;" json:"-"`
	DependencyIDs []string     `gorm:"-" json:"dependencies"`

	// after query loaded items
	LatestConfig *ComponentConfigConnection `gorm:"-" json:"-"`
}

func (c *Component) AfterQuery(tx *gorm.DB) error {
	if len(c.ComponentConfigs) < 1 {
		return nil
	}

	c.LatestConfig = &c.ComponentConfigs[0]
	return nil
}

func (c *Component) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
