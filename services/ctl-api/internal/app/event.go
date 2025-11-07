package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type Event struct {
	ID          string `json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"precision:6" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"precision" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"type:Nullable(Int8)" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID       string `json:"org_id,omitzero" gorm:"type:LowCardinality(String)" temporaljson:"org_id,omitzero,omitempty"`
	AppID       string `json:"app_id,omitzero" gorm:"type:LowCardinality(String)" temporaljson:"app_id,omitzero,omitempty"`
	InstallID   string `json:"install_id,omitzero" gorm:"type:LowCardinality(String)" temporaljson:"install_id,omitzero,omitempty"`
	ComponentID string `json:"component_id,omitzero" gorm:"type:LowCardinality(String)" temporaljson:"component_id,omitzero,omitempty"`
	RunnerID    string `json:"runner_id,omitzero" temporaljson:"runner_id,omitzero,omitempty"`
}

func (r *Event) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Event{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *Event) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}
