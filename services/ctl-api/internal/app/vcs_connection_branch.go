package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type VCSConnectionBranchStatus string

const (
	VCSConnectionBranchStatusActive  VCSConnectionBranchStatus = "active"
	VCSConnectionBranchStatusDeleted VCSConnectionBranchStatus = "deleted"
)

type VCSConnectionBranch struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Name   string                    `json:"name,omitzero" gorm:"notnull" temporaljson:"name,omitzero,omitempty"`
	Status VCSConnectionBranchStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	VCSConnectionRepo   VCSConnectionRepo `json:"vcs_connection_repo" temporaljson:"vcs_connection_repo,omitzero,omitempty"`
	VCSConnectionRepoID string            `json:"vcs_connection_repo_id,omitzero" gorm:"notnull" temporaljson:"vcs_connection_repo_id,omitzero,omitempty"`

	VCSConnectionCommits []VCSConnectionCommit `json:"vcs_connection_commits,omitzero" temporaljson:"vcs_connection_commits,omitzero,omitempty"`
}

func (b *VCSConnectionBranch) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = domains.NewVCSBranchID()
	}

	if b.CreatedByID == "" {
		b.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (b *VCSConnectionBranch) LatestKnownCommit() *VCSConnectionCommit {
	if len(b.VCSConnectionCommits) == 0 {
		return nil
	}
	return &b.VCSConnectionCommits[len(b.VCSConnectionCommits)-1]
}
