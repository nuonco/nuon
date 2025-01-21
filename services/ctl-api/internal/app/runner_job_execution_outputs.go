package app

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type RunnerJobExecutionOutputs struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id"`
	Org   Org    `json:"-"`

	RunnerJobExecutionID string             `json:"runner_job_execution_id" gorm:"defaultnull;notnull;index:idx_runner_job_execution_outputs,unique"`
	RunnerJobExecution   RunnerJobExecution `json:"-"`

	Outputs []byte `json:"outputs_json" gorm:"type:jsonb" swaggertype:"string"`

	// after query

	ParsedOutputs map[string]interface{} `json:"outputs" temporaljson:"-" gorm:"-" swaggertype:"object,object"`
}

func (r *RunnerJobExecutionOutputs) BeforeCreate(tx *gorm.DB) error {
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

func (r *RunnerJobExecutionOutputs) AfterQuery(tx *gorm.DB) error {
	if len(r.Outputs) > 0 {
		var outputs map[string]interface{}
		if err := json.Unmarshal(r.Outputs, &outputs); err != nil {
			return errors.Wrap(err, "unable to parse outputs json")
		}
		r.ParsedOutputs = outputs
	}

	return nil
}
