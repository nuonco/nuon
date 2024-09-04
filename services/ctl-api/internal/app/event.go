package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type Event struct {
	ID          string `json:"id"`
	CreatedByID string `json:"created_by_id"`

	CreatedAt time.Time             `json:"created_at" gorm:"precision:6"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"precision"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"type:Nullable(Int8)"`

	OrgID       string `json:"org_id" gorm:"type:LowCardinality(String)"`
	AppID       string `json:"app_id" gorm:"type:LowCardinality(String)"`
	InstallID   string `json:"install_id" gorm:"type:LowCardinality(String)"`
	ComponentID string `json:"component_id" gorm:"type:LowCardinality(String)"`
	RunnerID    string `json:"runner_id"`
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
