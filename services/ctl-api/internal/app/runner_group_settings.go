package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type RunnerGroupSettings struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"created_by"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_group_settings,unique"`

	OrgID string `json:"org_id" gorm:"index:idx_app_name,unique"`

	RunnerGroupID string `json:"runner_group_id" gorm:"index:idx_runner_group_settings,unique"`

	// configuration for deploying the runner
	ContainerImageURL      string        `json:"container_image_url"  gorm:"default null;not null"`
	ContainerImageTag      string        `json:"container_image_tag"  gorm:"default null;not null"`
	RunnerAPIURL           string        `json:"runner_api_url" gorm:"default null;not null"`
	SettingsRefreshTimeout time.Duration `json:"settings_refresh_timeout" swaggertype:"primitive,integer"`

	// Various settings for the runner to handle internally
	HeartBeatTimeout           time.Duration `json:"heart_beat_timeout" gorm:"default null;" swaggertype:"primitive,integer"`
	OTELCollectorConfiguration string        `json:"otel_collector_config" gorm:"default null;not null"`
}

func (r *RunnerGroupSettings) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerGroupSettingsID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
