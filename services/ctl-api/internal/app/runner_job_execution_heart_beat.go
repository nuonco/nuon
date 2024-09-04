package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

// clickhouse table
type RunnerJobExecutionHeartBeat struct {
	ID          string `gorm:"" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"not null;default:null"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:""`

	RunnerID             string `json:"runner_id" gorm:""`
	RunnerJobID          string `json:"runner_job_id" gorm:""`
	RunnerJobExecutionID string `json:"runner_job_execution_id" gorm:""`

	AliveTime time.Duration `json:"alive_time" swaggertype:"primitive,integer"`
}

func (r *RunnerJobExecutionHeartBeat) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}
