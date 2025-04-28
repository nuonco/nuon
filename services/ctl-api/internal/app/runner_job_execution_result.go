package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type RunnerJobExecutionResult struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_job_execution_result,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerJobExecutionID string `json:"runner_job_execution_id" gorm:"defaultnull;notnull;index:idx_job_execution_result,unique" temporaljson:"runner_job_execution_id,omitzero,omitempty"`

	Success bool `json:"success" temporaljson:"success,omitzero,omitempty"`

	ErrorCode     int           `json:"error_code" temporaljson:"error_code,omitzero,omitempty"`
	ErrorMetadata pgtype.Hstore `json:"error_metadata" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"error_metadata,omitzero,omitempty"`
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
