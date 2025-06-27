package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

// ComponentPush represents a GitHub push event for a component.
type ComponentPush struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// INFO: we can use repo/branch/directory as the unique identifiers for the component configs to rebuild
	Directory string `json:"directory,omitzero" gorm:"notnull" temporaljson:"directory,omitzero,omitempty"`

	VCSConnectionCommitID string              `json:"-" gorm:"not null;default:null" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   VCSConnectionCommit `json:"vcs_connection_commit,omitzero" temporaljson:"vcs_connection_commit,omitzero,omitempty"`
}

func (c *ComponentPush) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentGithubPushID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
