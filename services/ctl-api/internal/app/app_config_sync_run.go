package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AppConfigSyncRun struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`

	AppBranchID string    `json:"app_branch_id,omitzero" gorm:"not null;default:null" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch   AppBranch `json:"-" temporaljson:"app_branch,omitzero,omitempty"`

	WorkflowID string          `json:"workflow_id,omitzero" gorm:"not null" temporaljson:"flow_id,omitzero,omitempty"`
	Workflow   InstallWorkflow `json:"-" temporaljson:"flow,omitzero,omitempty"`

	VCSConnectionCommitID *string              `json:"-" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitzero" temporaljson:"vcs_connection_commit,omitzero,omitempty"`
	Directory             string               `json:"directory,omitzero" gorm:"not null" temporaljson:"directory,omitzero,omitempty"`

	ConfigPayload []byte `json:"config_payload,omitzero" temporaljson:"config_payload,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
}

func (a *AppConfigSyncRun) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAppConfigSyncRunID()
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
