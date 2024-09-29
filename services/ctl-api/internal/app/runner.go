package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type RunnerStatus string

const (
	RunnerStatusError          RunnerStatus = "error"
	RunnerStatusActive         RunnerStatus = "active"
	RunnerStatusPending        RunnerStatus = "pending"
	RunnerStatusProvisioning   RunnerStatus = "provisioning"
	RunnerStatusDeprovisioning RunnerStatus = "deprovisioning"
	RunnerStatusDeprovisioned  RunnerStatus = "deprovisioned"
	RunnerStatusReprovisioning RunnerStatus = "reprovisioning"
	RunnerStatusUnknown        RunnerStatus = "unknown"
)

func (r RunnerStatus) IsHealthy() bool {
	return r == RunnerStatusActive
}

type Runner struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"created_by"`

	OrgID string `json:"org_id" gorm:"index:idx_app_name,unique"`
	Org   Org    `json:"org"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_name,unique"`

	Status            RunnerStatus `json:"status" gorm:"not null;default null" swaggertype:"string"`
	StatusDescription string       `json:"status_description" gorm:"not null;default null"`

	RunnerGroupID string      `json:"runner_id" gorm:"index:idx_runner_name,unique"`
	RunnerGroup   RunnerGroup `json:"-"`

	Name        string `json:"name" gorm:"index:idx_runner_name,unique"`
	DisplayName string `json:"display_name" gorm:"not null;default null"`

	Jobs []RunnerJob `json:"jobs" gorm:"constraint:OnDelete:CASCADE;"`

	RunnerJob *RunnerJob `json:"runner_job" gorm:"polymorphic:Owner;"`
}

func (r *Runner) BeforeCreate(tx *gorm.DB) error {
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
