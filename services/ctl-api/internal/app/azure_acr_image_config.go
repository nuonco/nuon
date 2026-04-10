package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AzureACRImageConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	// connection to parent model
	ComponentConfigID   string `json:"component_config_id,omitzero" gorm:"notnull" temporaljson:"component_config_id,omitzero,omitempty"`
	ComponentConfigType string `json:"component_config_type,omitzero" gorm:"notnull" temporaljson:"component_config_type,omitzero,omitempty"`

	// actual configuration
	RegistryURL string `json:"registry_url,omitzero" gorm:"notnull" temporaljson:"registry_url,omitzero,omitempty"`
	TenantID    string `json:"tenant_id,omitzero" temporaljson:"tenant_id,omitzero,omitempty"`
	ClientID    string `json:"client_id,omitzero" temporaljson:"client_id,omitzero,omitempty"`
}

func (c *AzureACRImageConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AzureACRImageConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (c *AzureACRImageConfig) CredentialsConfig() *azurecredentials.Config {
	if c.TenantID == "" && c.ClientID == "" {
		return nil
	}
	return &azurecredentials.Config{
		UseDefault: true,
		ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
			SubscriptionTenantID: c.TenantID,
		},
	}
}

func (c *AzureACRImageConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
