package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AzureAccount struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"-" gorm:"notnull" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `json:"-" temporaljson:"install,omitzero,omitempty"`

	Location                 string `json:"location,omitzero" gorm:"notnull" temporaljson:"location,omitzero,omitempty"`
	SubscriptionID           string `json:"subscription_id,omitzero" gorm:"not null;default null" temporaljson:"subscription_id,omitzero,omitempty"`
	SubscriptionTenantID     string `json:"subscription_tenant_id,omitzero" gorm:"not null;default null" temporaljson:"subscription_tenant_id,omitzero,omitempty"`
	ServicePrincipalAppID    string `json:"service_principal_app_id,omitzero" gorm:"not null;default null" temporaljson:"service_principal_app_id,omitzero,omitempty"`
	ServicePrincipalPassword string `json:"service_principal_password,omitzero" gorm:"not null;default null" temporaljson:"service_principal_password,omitzero,omitempty"`
}

func (a *AzureAccount) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AzureAccount{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *AzureAccount) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAzureAccountID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
