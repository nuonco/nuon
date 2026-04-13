package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type StackType string

const (
	StackTypeAWS   StackType = "aws-cloudformation"
	StackTypeAzure StackType = "azure-bicep"
	StackTypeGCP   StackType = "gcp-terraform"
)

type AppStackConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID       string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	Type                    StackType `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
	Name                    string    `json:"name,omitzero" features:"template" temporaljson:"name,omitzero,omitempty"`
	Description             string    `json:"description,omitzero" features:"template" temporaljson:"description,omitzero,omitempty"`
	RunnerNestedTemplateURL string    `json:"runner_nested_template_url,omitzero" temporaljson:"runner_nested_template_url,omitzero,omitempty" features:"template"`
	VPCNestedTemplateURL    string    `json:"vpc_nested_template_url,omitzero" temporaljson:"vpc_nested_template_url,omitzero,omitempty" features:"template"`

	CustomNestedStacks []config.CustomNestedStack `json:"custom_nested_stacks,omitzero" gorm:"type:jsonb;serializer:json;default:'[]'" temporaljson:"custom_nested_stacks,omitzero,omitempty"`
}

func (a *AppStackConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppStackConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

// HasAzureCustomization returns true when the vendor has configured a
// stack.toml with type "azure-bicep", opting in to the programmatic ARM
// template builder. Without a stack config we fall back to the embedded
// monolithic Bicep template for backwards compatibility.
func (sc *AppStackConfig) HasAzureCustomization() bool {
	return sc.Type == StackTypeAzure ||
		sc.VPCNestedTemplateURL != "" ||
		sc.RunnerNestedTemplateURL != "" ||
		len(sc.CustomNestedStacks) > 0
}

func (a *AppStackConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
