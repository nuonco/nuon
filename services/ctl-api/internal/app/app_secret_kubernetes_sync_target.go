package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// AppSecretKubernetesSyncTarget is a single Kubernetes destination for a Nuon secret. A secret may define multiple
// targets, each fanning out to one or more namespaces under a given secret name and key. This enables a single Nuon
// secret to sync into multiple namespaces, and multiple Nuon secrets to populate different keys of the same
// Kubernetes secret.
type AppSecretKubernetesSyncTarget struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppSecretConfigID string `json:"app_secret_config_id,omitzero" temporaljson:"app_secret_config_id,omitzero,omitempty"`

	Namespaces pq.StringArray `json:"namespaces,omitzero" gorm:"type:text[]" swaggertype:"array,string" features:"template" temporaljson:"namespaces,omitzero,omitempty"`
	Name       string         `json:"name,omitzero" features:"template" temporaljson:"name,omitzero,omitempty"`
	Key        string         `json:"key,omitzero" features:"template" temporaljson:"key,omitzero,omitempty"`
}

func (a *AppSecretKubernetesSyncTarget) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppSecretKubernetesSyncTarget{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppSecretKubernetesSyncTarget{}, "app_secret_config_id"),
			Columns: []string{
				"app_secret_config_id",
			},
		},
	}
}

func (a *AppSecretKubernetesSyncTarget) BeforeCreate(tx *gorm.DB) error {
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
