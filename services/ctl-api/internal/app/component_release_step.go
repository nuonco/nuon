package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ComponentReleaseStep struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// parent release ID
	ComponentReleaseID string           `json:"component_release_id,omitzero" temporaljson:"component_release_id,omitzero,omitempty"`
	ComponentRelease   ComponentRelease `json:"-" temporaljson:"component_release,omitzero,omitempty"`

	Status            string `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string `json:"status_description,omitzero" temporaljson:"status_description,omitzero,omitempty"`

	// When a step is created, a set of installs are targeted. However, by the time the release step goes out, the
	// install might have been setup in any order of ways.
	RequestedInstallIDs pq.StringArray  `json:"requested_install_ids,omitzero" swaggertype:"array,string" gorm:"type:text[]" temporaljson:"requested_install_i_ds,omitzero,omitempty"`
	InstallDeploys      []InstallDeploy `json:"install_deploys,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_deploys,omitzero,omitempty"`

	// fields to control the delay of the individual step, as this is set based on the parent strategy
	Delay *string `json:"delay,omitzero" temporaljson:"delay,omitzero,omitempty"`
}

func (a *ComponentReleaseStep) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &ComponentReleaseStep{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *ComponentReleaseStep) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewReleaseID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
