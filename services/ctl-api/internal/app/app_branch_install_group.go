package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppBranchInstallGroup struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchConfigID string          `json:"app_branch_config_id,omitzero" gorm:"not null;uniqueIndex:idx_app_branch_install_group_order" temporaljson:"app_branch_config_id,omitzero,omitempty"`
	AppBranchConfig   AppBranchConfig `faker:"-" json:"-" temporaljson:"app_branch_config,omitzero,omitempty"`

	Name       string         `json:"name,omitzero" gorm:"not null" temporaljson:"name,omitzero,omitempty"`
	Order      int            `json:"order,omitzero" gorm:"not null;uniqueIndex:idx_app_branch_install_group_order" temporaljson:"order,omitzero,omitempty"`
	InstallIDs pq.StringArray `gorm:"type:text[]" json:"install_ids,omitzero" temporaljson:"install_ids,omitzero,omitempty" swaggertype:"array,string"`

	RequiresApproval  bool `json:"requires_approval,omitzero" gorm:"default:false" temporaljson:"requires_approval,omitzero,omitempty"`
	RollbackOnFailure bool `json:"rollback_on_failure,omitzero" gorm:"default:true" temporaljson:"rollback_on_failure,omitzero,omitempty"`
	MaxParallel       int  `json:"max_parallel,omitzero" gorm:"default:5" temporaljson:"max_parallel,omitzero,omitempty"`
}

func (a *AppBranchInstallGroup) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppBranchInstallGroup{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppBranchInstallGroup{}, "app_branch_config_id"),
			Columns: []string{
				"app_branch_config_id",
			},
		},
	}
}

func (a *AppBranchInstallGroup) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppBranchInstallGroupID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
