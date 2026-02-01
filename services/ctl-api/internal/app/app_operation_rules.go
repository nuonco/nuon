package app

import (
	"time"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// OperationRoleRule represents a single rule mapping principal + operation -> role
type OperationRoleRule struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"org" gorm:"-" temporaljson:"org,omitzero,omitempty"`

	AppOperationRoleConfigID string                 `json:"app_operation_role_config" gorm:"app_operation_role_config"`
	AppOperationRoleConfig   AppOperationRoleConfig `json:"app_operation" gorm:"-"`

	// "nuon::component:name", "nuon::sandbox", "nuon::action:name"
	Principal string `json:"principal" gorm:"column:principal;not null;index"`
	// "provision", "deprovision", "update", "reprovision", "trigger"
	Operation OperationType `json:"operation" gorm:"column:operation;not null;index"`
	// Role name (not ARN)
	Role string `json:"role" gorm:"column:role;not null"`

	Config *AppOperationRoleConfig `json:"-" gorm:"foreignKey:ConfigID"`
}

func (o *OperationRoleRule) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = domains.NewAppID()
	}
	if o.CreatedByID == "" {
		o.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if o.OrgID == "" {
		o.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
