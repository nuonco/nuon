package app

import (
	"time"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OperationType string

const (
	OperationProvision   OperationType = "provision"
	OperationDeprovision OperationType = "deprovision"
	OperationDeploy      OperationType = "deploy"
	OperationTeardown    OperationType = "teardown"
	OperationReprovision OperationType = "reprovision"
	OperationTrigger     OperationType = "trigger"
)

var ValidOperations = []OperationType{
	OperationProvision,
	OperationDeprovision,
	OperationReprovision,
	OperationDeploy,
	OperationTeardown,
	OperationTrigger,
}

// AppOperationRoleRule represents a single rule mapping principal + operation -> role
type AppOperationRoleRule struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"org" gorm:"-" temporaljson:"org,omitzero,omitempty"`

	AppOperationRoleConfigID string                 `json:"app_operation_role_config_id" gorm:"app_operation_role_config_id"`
	AppOperationRoleConfig   AppOperationRoleConfig `json:"-" gorm:"-"`

	PrincipalType principal.Type `json:"principal_type" gorm:"column:principal_type;not null;index" swaggertype:"string"`
	PrincipalName string         `json:"principal_name" gorm:"column:principal_name;index"`
	Operation     OperationType  `json:"operation" gorm:"column:operation;not null;index" swaggertype:"string"`
	Role          string         `json:"role" gorm:"column:role;not null"`
}

func (o *AppOperationRoleRule) BeforeCreate(tx *gorm.DB) error {
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
