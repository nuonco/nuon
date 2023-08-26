package app

import (
	"time"

	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ComponentReleaseStep struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// parent release ID
	ComponentReleaseID string `json:"component_release_id"`
	ComponentRelease   ComponentRelease

	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`

	// When a step is created, a set of installs are targeted. However, by the time the release step goes out, the
	// install might have been setup in any order of ways.
	RequestedInstallIDs pq.StringArray  `gorm:"type:text[]" json:"requested_install_ids" swaggertype:"array,string"`
	InstallDeploys      []InstallDeploy `json:"install_deploys,omitempty"`

	// fields to control the delay of the individual step, as this is set based on the parent strategy
	Delay *string `json:"delay"`
}

func (a *ComponentReleaseStep) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewReleaseID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
