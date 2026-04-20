package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallRoleUsage struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID       string                `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org         Org                   `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallRoleID string       `json:"install_role_id,omitzero" gorm:"notnull;default null" temporaljson:"install_role_id,omitzero,omitempty"`
	InstallRole   InstallRoles `json:"-" temporaljson:"install_role,omitzero,omitempty"`

	RunnerJobID string    `json:"runner_job_id,omitzero" gorm:"notnull;default null" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerJob   RunnerJob `json:"-" temporaljson:"runner_job,omitzero,omitempty"`

	RoleName   string `json:"role_name,omitzero" temporaljson:"role_name,omitzero,omitempty"`
	RoleSource string `json:"role_source,omitzero" temporaljson:"role_source,omitzero,omitempty"`

	RoleSelectionTrace RunnerJobPermissionTrace `json:"role_selection_trace,omitzero" gorm:"type:jsonb" temporaljson:"role_selection_trace,omitzero,omitempty"`
}

func (i *InstallRoleUsage) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallRoleUsage{}, "idx_org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallRoleUsage{}, "idx_install_role_id"),
			Columns: []string{
				"install_role_id",
			},
		},
	}
}

func (a *InstallRoleUsage) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
