package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type RunnerJobExecutionResult struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_job_execution_result,unique"`

	OrgID string `json:"org_id"`
	Org   Org    `json:"-"`

	RunnerJobExecutionID string `json:"runner_job_execution_id" gorm:"defaultnull;notnull;index:idx_job_execution_result,unique"`

	Success bool `json:"success"`

	ErrorCode     int           `json:"error_code"`
	ErrorMetadata pgtype.Hstore `json:"error_metadata" gorm:"type:hstore" swaggertype:"object,string"`
}

func (r *RunnerJobExecutionResult) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
