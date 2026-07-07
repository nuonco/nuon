package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// AppKubernetesContextConfig is a single named kubernetes context binding
// on an AppConfig version. Each context names a stable identifier that
// components reference via ComponentConfigConnection.KubernetesContextName,
// and points at a peer terraform_module or pulumi component that emits the
// cluster outputs.
//
// SourceComponentName is the stable string written from config; the FK
// SourceComponentID is resolved at create time once component IDs are known
// (same pattern ComponentConfigDependency uses for name -> ID resolution).
type AppKubernetesContextConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `faker:"-" json:"-" temporaljson:"app,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	AppKubernetesContextsConfig   AppKubernetesContextsConfig `json:"-" faker:"-" temporaljson:"app_kubernetes_contexts_config,omitzero,omitempty"`
	AppKubernetesContextsConfigID string                      `json:"app_kubernetes_contexts_config_id,omitzero" temporaljson:"app_kubernetes_contexts_config_id,omitzero,omitempty"`

	Name string `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`

	SourceComponentName string `json:"source_component_name,omitzero" temporaljson:"source_component_name,omitzero,omitempty"`
	SourceComponentID   string `json:"source_component_id,omitzero" temporaljson:"source_component_id,omitzero,omitempty"`
}

func (a *AppKubernetesContextConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppKubernetesContextConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppKubernetesContextConfig{}, "app_config_id"),
			Columns: []string{
				"app_config_id",
			},
		},
		{
			// Each AppConfig version can have at most one context per name.
			// Scoped to live rows (deleted_at = 0) so soft-deleted history
			// doesn't conflict.
			Name: indexes.Name(db, &AppKubernetesContextConfig{}, "app_config_id_name"),
			Columns: []string{
				"app_config_id",
				"name",
			},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
			Option:      "WHERE deleted_at = 0",
		},
	}
}

func (a *AppKubernetesContextConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppCfgID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
