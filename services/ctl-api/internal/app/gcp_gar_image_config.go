package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type GCPGARImageConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	ComponentConfigID   string `json:"component_config_id,omitzero" gorm:"notnull" temporaljson:"component_config_id,omitzero,omitempty"`
	ComponentConfigType string `json:"component_config_type,omitzero" gorm:"notnull" temporaljson:"component_config_type,omitzero,omitempty"`

	GCPProjectID             string `json:"gcp_project_id,omitzero" gorm:"notnull" temporaljson:"gcp_project_id,omitzero,omitempty"`
	GCPRegion                string `json:"gcp_region,omitzero" gorm:"notnull" temporaljson:"gcp_region,omitzero,omitempty"`
	ServiceAccountEmail      string `json:"service_account_email,omitzero" temporaljson:"service_account_email,omitzero,omitempty"`
	WorkloadIdentityProvider string `json:"workload_identity_provider,omitzero" temporaljson:"workload_identity_provider,omitzero,omitempty"`
}

func (c *GCPGARImageConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &GCPGARImageConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (c *GCPGARImageConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
