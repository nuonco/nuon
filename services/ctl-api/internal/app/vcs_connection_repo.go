package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type VCSConnectionRepoStatus string

const (
	VCSConnectionRepoStatusActive   VCSConnectionRepoStatus = "active"
	VCSConnectionRepoStatusDeleted  VCSConnectionRepoStatus = "deleted"
	VCSConnectionRepoStatusArchived VCSConnectionRepoStatus = "archived"
)

type VCSConnectionRepo struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Name   string                  `json:"name,omitzero" gorm:"notnull" temporaljson:"name,omitzero,omitempty"`
	Status VCSConnectionRepoStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	VCSConnection   VCSConnection `json:"-" temporaljson:"vcs_connection,omitzero,omitempty"`
	VCSConnectionID string        `json:"vcs_connection_id,omitzero" gorm:"notnull" temporaljson:"vcs_connection_id,omitzero,omitempty"`
}

func (a *VCSConnectionRepo) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewVCSID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}
