package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// job details
	LogStream LogStream `json:"log_stream" gorm:"polymorphic:Owner;"`

	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	RunnerID string `json:"runner_id"`
	Runner   Runner `json:"-" faker:"-"`

	OpType            RunnerOperationType   `json:"operation_type"`
	Status            RunnerOperationStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string                `json:"status_description" gorm:"notnull"`
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
