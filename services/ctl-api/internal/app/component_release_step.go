package app

import (
	"time"

	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type ComponentReleaseStep struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	// parent release ID
	ComponentReleaseID string           `json:"component_release_id"`
	ComponentRelease   ComponentRelease `json:"-"`

	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`

	// When a step is created, a set of installs are targeted. However, by the time the release step goes out, the
	// install might have been setup in any order of ways.
	RequestedInstallIDs pq.StringArray  `json:"requested_install_ids" swaggertype:"array,string" gorm:"type:text[]"`
	InstallDeploys      []InstallDeploy `json:"install_deploys,omitempty" gorm:"constraint:OnDelete:CASCADE;"`

	// fields to control the delay of the individual step, as this is set based on the parent strategy
	Delay *string `json:"delay"`
}

func (a *ComponentReleaseStep) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewReleaseID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
