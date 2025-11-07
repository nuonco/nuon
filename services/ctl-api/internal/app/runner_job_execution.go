package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerJobExecutionStatus string

const (
	// the following statuses denote an in-progress execution
	// initializing means the runner is starting the job
	RunnerJobExecutionStatusPending RunnerJobExecutionStatus = "pending"
	// initializing means the runner is starting the job
	RunnerJobExecutionStatusInitializing RunnerJobExecutionStatus = "initializing"
	// means the runner is in progress
	RunnerJobExecutionStatusInProgress RunnerJobExecutionStatus = "in-progress"
	// means the runner is cleaning up
	RunnerJobExecutionStatusCleaningUp RunnerJobExecutionStatus = "cleaning-up"

	// the following statuses denote a finished execution
	// once a runner has finished the job successfully
	RunnerJobExecutionStatusFinished RunnerJobExecutionStatus = "finished"
	// once a runner has failed the job
	RunnerJobExecutionStatusFailed RunnerJobExecutionStatus = "failed"
	// once the job has timed out
	RunnerJobExecutionStatusTimedOut RunnerJobExecutionStatus = "timed-out"
	// not attempted is when the runner can not attempt
	RunnerJobExecutionStatusNotAttempted RunnerJobExecutionStatus = "not-attempted"
	// when a job is cancelled
	RunnerJobExecutionStatusCancelled RunnerJobExecutionStatus = "cancelled"
	// when a job status is unknown
	RunnerJobExecutionStatusUnknown RunnerJobExecutionStatus = "unknown"
)

func (r RunnerJobExecutionStatus) IsRunning() bool {
	switch r {
	case RunnerJobExecutionStatusPending,
		RunnerJobExecutionStatusInitializing,
		RunnerJobExecutionStatusInProgress,
		RunnerJobExecutionStatusCleaningUp:
		return true
	default:
		return false
	}
}

// each runner job can be retried one or more times
// each execution will be tracked and have logs, metrics, events and more
type RunnerJobExecution struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_job_execution_runner_job_id,type:btree" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerJobID string    `json:"runner_job_id,omitzero" gorm:"notnull;defaultnull;index:idx_runner_job_execution_runner_job_id,type:btree" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerJob   RunnerJob `json:"-" temporaljson:"runner_job,omitzero,omitempty"`

	Status RunnerJobExecutionStatus `json:"status,omitzero" gorm:"not null;default null;index:idx_runner_job_execution_status,type:hash" temporaljson:"status,omitzero,omitempty"`

	Result  *RunnerJobExecutionResult  `json:"result,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"result,omitzero,omitempty"`
	Outputs *RunnerJobExecutionOutputs `json:"outputs,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"outputs,omitzero,omitempty"`
}

func (r *RunnerJobExecution) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (i *RunnerJobExecution) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerJobExecution{}, "runner_job_execution_partitions"),
			Columns: []string{
				"runner_job_id",
				"created_at",
			},
		},
		{
			Name: indexes.Name(db, &RunnerJobExecution{}, "runner_jobs"),
			Columns: []string{
				"runner_job_id",
			},
		},
		{
			Name: indexes.Name(db, &RunnerJobExecution{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
