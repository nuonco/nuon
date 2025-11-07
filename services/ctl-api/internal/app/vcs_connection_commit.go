package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type VCSConnectionCommit struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	VCSConnection   VCSConnection `json:"-" temporaljson:"vcs_connection,omitzero,omitempty"`
	VCSConnectionID string        `json:"vcs_connection_id,omitzero" gorm:"notnull" temporaljson:"vcs_connection_id,omitzero,omitempty"`

	ComponentBuilds []ComponentBuild `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_builds,omitzero,omitempty"`

	SHA         string `json:"sha,omitzero" gorm:"notnull" temporaljson:"sha,omitzero,omitempty"`
	AuthorName  string `json:"author_name,omitzero" temporaljson:"author_name,omitzero,omitempty"`
	AuthorEmail string `json:"author_email,omitzero" temporaljson:"author_email,omitzero,omitempty"`
	Message     string `json:"message,omitzero" temporaljson:"message,omitzero,omitempty"`
}

func (v *VCSConnectionCommit) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &VCSConnectionCommit{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (v *VCSConnectionCommit) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSConnectionID()

	if v.CreatedByID == "" {
		v.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if v.OrgID == "" {
		v.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
