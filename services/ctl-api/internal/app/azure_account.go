package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AzureAccount struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `faker:"-" json:"-"`

	InstallID string `json:"-" gorm:"notnull"`
	Install   Install

	Location                 string `json:"location" gorm:"notnull"`
	SubscriptionID           string `json:"subscription_id" gorm:"not null;default null"`
	SubscriptionTenantID     string `json:"subscription_tenant_id" gorm:"not null;default null"`
	ServicePrincipalAppID    string `json:"service_principal_app_id" gorm:"not null;default null"`
	ServicePrincipalPassword string `json:"service_principal_password" gorm:"not null;default null"`
}

func (a *AzureAccount) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAzureAccountID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
