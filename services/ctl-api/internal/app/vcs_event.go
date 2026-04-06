package app

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type VCSEventPayload map[string]any

func (p VCSEventPayload) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

func (p *VCSEventPayload) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}

type VCSEvent struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `swaggerignore:"true" json:"-" temporaljson:"org,omitzero,omitempty"`

	VCSConnectionID string `json:"vcs_connection_id" gorm:"not null" temporaljson:"vcs_connection_id,omitzero,omitempty"`
	EventType       string `json:"event_type" gorm:"not null;default:''" temporaljson:"event_type,omitzero,omitempty"`

	Payload VCSEventPayload  `json:"payload" gorm:"type:jsonb;default:null" temporaljson:"payload,omitzero,omitempty"`
	Status  *CompositeStatus `json:"status,omitempty" gorm:"column:status;type:jsonb;default:null" temporaljson:"status,omitzero,omitempty"`
}

func (v *VCSEvent) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &VCSEvent{}, "vcs_connection_id"),
			Columns: []string{"vcs_connection_id"},
		},
	}
}

func (v *VCSEvent) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSEventID()
	if v.OrgID == "" {
		v.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if v.CreatedByID == "" {
		v.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
