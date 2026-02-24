package app

import (
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// AppOperationRoleConfig stores operation role configuration for an app
type AppOperationRoleConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID       string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	Rules []*AppOperationRoleRule `json:"rules,omitempty" gorm:"foreignKey:AppOperationRoleConfigID"`
}

func (a *AppOperationRoleConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{}
}

func (a *AppOperationRoleConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if a.AppID != "" && a.OrgID != "" {
		var count int64
		if err := tx.Model(&App{}).Where("id = ? AND org_id = ?", a.AppID, a.OrgID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("app %s does not belong to org %s", a.AppID, a.OrgID)
		}
	}

	return nil
}
