package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/labels"
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

	// LabelSelector dynamically resolves installs at deploy time by matching labels.
	// Mutually exclusive with InstallIDs — set one or the other, not both.

	LabelSelector *labels.Selector `json:"label_selector,omitempty" gorm:"type:jsonb;serializer:json;default:null" temporaljson:"label_selector,omitzero,omitempty"`

	// UseForPreviews marks this group for plan-only preview runs (e.g., PR previews).
	UseForPreviews bool `json:"use_for_previews,omitzero" gorm:"default:false" temporaljson:"use_for_previews,omitzero,omitempty"`
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
