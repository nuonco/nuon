package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallGroupRunInstall struct {
	InstallID  string `json:"install_id"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Status     string `json:"status"`
}

type InstallGroupRun struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchRunID string       `json:"app_branch_run_id,omitzero" gorm:"not null" temporaljson:"app_branch_run_id,omitzero,omitempty"`
	AppBranchRun   AppBranchRun `faker:"-" json:"-" temporaljson:"app_branch_run,omitzero,omitempty"`

	InstallGroupID string                `json:"install_group_id,omitzero" gorm:"not null" temporaljson:"install_group_id,omitzero,omitempty"`
	InstallGroup   AppBranchInstallGroup `faker:"-" json:"install_group,omitempty" temporaljson:"install_group,omitzero,omitempty"`

	InstallGroupName string `json:"install_group_name,omitzero" gorm:"not null" temporaljson:"install_group_name,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	Installs []InstallGroupRunInstall `json:"installs,omitempty" gorm:"type:jsonb;serializer:json" temporaljson:"installs,omitzero,omitempty"`

	TotalInstalls     int `json:"total_installs,omitzero" temporaljson:"total_installs,omitzero,omitempty"`
	CompletedInstalls int `json:"completed_installs,omitzero" temporaljson:"completed_installs,omitzero,omitempty"`
	FailedInstalls    int `json:"failed_installs,omitzero" temporaljson:"failed_installs,omitzero,omitempty"`

	StartedAt   *time.Time `json:"started_at,omitempty" temporaljson:"started_at,omitzero,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty" temporaljson:"completed_at,omitzero,omitempty"`
}

func (i *InstallGroupRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallGroupRun{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallGroupRun{}, "app_branch_run_id"),
			Columns: []string{
				"app_branch_run_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallGroupRun{}, "install_group_id"),
			Columns: []string{
				"install_group_id",
			},
		},
	}
}

func (i *InstallGroupRun) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallGroupRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
