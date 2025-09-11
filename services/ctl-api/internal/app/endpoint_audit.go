package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type EndpointAudit struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Method     string            `json:"method,omitzero" gorm:"not null;default null" temporaljson:"method,omitzero,omitempty"`
	Name       string            `json:"name,omitzero" gorm:"not null;default null" temporaljson:"name,omitzero,omitempty"`
	Route      string            `json:"route,omitzero" gorm:"not null;default null" temporaljson:"route,omitzero,omitempty"`
	LastUsedAt generics.NullTime `json:"last_used_at,omitzero" gorm:"type:timestamp;default:null" temporaljson:"last_used_at,omitzero,omitempty"`

	Deprecated bool `json:"deprecated,omitzero" gorm:"not null;default:false" temporaljson:"deprecated,omitzero,omitempty"`
}

func (a *EndpointAudit) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewEndpointAuditID()
	}

	return nil
}

func (c *EndpointAudit) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &EndpointAudit{}, "uq"),
			Columns: []string{
				"deleted_at",
				"method",
				"name",
				"route",
			},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
	}
}
