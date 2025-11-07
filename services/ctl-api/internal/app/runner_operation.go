package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerOperationType string

const (
	RunnerOperationTypeProvision               RunnerOperationType = "provision"
	RunnerOperationTypeProvisionServiceAccount RunnerOperationType = "provision_service_account"
	RunnerOperationTypeReprovision             RunnerOperationType = "reprovision"
	RunnerOperationTypeDeprovision             RunnerOperationType = "deprovision"
)

type RunnerOperationStatus string

const (
	RunnerOperationStatusFinished   RunnerOperationStatus = "finished"
	RunnerOperationStatusInProgress RunnerOperationStatus = "in-progress"
	RunnerOperationStatusPending    RunnerOperationStatus = "pending"
	RunnerOperationStatusError      RunnerOperationStatus = "error"
)

type RunnerOperation struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// job details
	LogStream LogStream `json:"log_stream,omitzero" gorm:"polymorphic:Owner;" temporaljson:"log_stream,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerID string `json:"runner_id,omitzero" temporaljson:"runner_id,omitzero,omitempty"`
	Runner   Runner `json:"-" faker:"-" temporaljson:"runner,omitzero,omitempty"`

	OpType            RunnerOperationType   `json:"operation_type,omitzero" temporaljson:"op_type,omitzero,omitempty"`
	Status            RunnerOperationStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
}

func (i *RunnerOperation) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerOperation{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (i *RunnerOperation) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewRunnerOperationID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
