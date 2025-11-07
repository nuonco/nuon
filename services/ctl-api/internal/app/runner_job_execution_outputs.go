package app

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerJobExecutionOutputs struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerJobExecutionID string             `json:"runner_job_execution_id,omitzero" gorm:"defaultnull;notnull;index:idx_runner_job_execution_outputs,unique" temporaljson:"runner_job_execution_id,omitzero,omitempty"`
	RunnerJobExecution   RunnerJobExecution `json:"-" temporaljson:"runner_job_execution,omitzero,omitempty"`

	Outputs []byte `json:"outputs_json,omitzero" gorm:"type:jsonb" swaggertype:"string" temporaljson:"outputs,omitzero,omitempty"`

	// after query

	ParsedOutputs map[string]interface{} `json:"outputs,omitzero" gorm:"-" swaggertype:"object,object" temporaljson:"parsed_outputs,omitzero,omitempty"`
}

func (r *RunnerJobExecutionOutputs) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerJobExecutionOutputs{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
