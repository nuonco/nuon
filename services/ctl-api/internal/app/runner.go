package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerStatus string

const (
	RunnerStatusError                   RunnerStatus = "error"
	RunnerStatusActive                  RunnerStatus = "active"
	RunnerStatusPending                 RunnerStatus = "pending"
	RunnerStatusProvisioning            RunnerStatus = "provisioning"
	RunnerStatusDeprovisioning          RunnerStatus = "deprovisioning"
	RunnerStatusDeprovisioned           RunnerStatus = "deprovisioned"
	RunnerStatusReprovisioning          RunnerStatus = "reprovisioning"
	RunnerStatusOffline                 RunnerStatus = "offline"
	RunnerStatusAwaitingInstallStackRun RunnerStatus = "awaiting-install-stack-run"

	RunnerStatusUnknown RunnerStatus = "unknown"
)

func (r RunnerStatus) String() string {
	return string(r)
}

func (r RunnerStatus) Code() int {
	switch r {

	// 2xx are for unknown
	case RunnerStatusPending:
		return 200
	case RunnerStatusProvisioning:
		return 201

		// 3xx statuses are for tear downs
	case RunnerStatusDeprovisioning:
		return 301
	case RunnerStatusDeprovisioned:
		return 300

		// 4xx
	case RunnerStatusError:
		return 400

		// 0 is active
	case RunnerStatusActive:
		return 0
	case RunnerStatusUnknown:
		return 500
	default:
		return 500
	}
}

func (r RunnerStatus) IsHealthy() bool {
	return r == RunnerStatusActive
}

type Runner struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"index:idx_app_name,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	Status            RunnerStatus `json:"status,omitzero" gorm:"not null;default null" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string       `json:"status_description,omitzero" gorm:"not null;default null" temporaljson:"status_description,omitzero,omitempty"`

	RunnerGroupID string      `json:"runner_group_id,omitzero" gorm:"index:idx_runner_name,unique" temporaljson:"runner_group_id,omitzero,omitempty"`
	RunnerGroup   RunnerGroup `json:"runner_group,omitzero" temporaljson:"runner_group,omitzero,omitempty"`

	Name        string `json:"name,omitzero" gorm:"index:idx_runner_name,unique" temporaljson:"name,omitzero,omitempty"`
	DisplayName string `json:"display_name,omitzero" gorm:"not null;default null" temporaljson:"display_name,omitzero,omitempty"`

	Jobs       []RunnerJob       `json:"jobs,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"jobs,omitzero,omitempty"`
	Operations []RunnerOperation `json:"operations,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"operations,omitzero,omitempty"`

	RunnerJob *RunnerJob `json:"runner_job,omitzero" gorm:"polymorphic:Owner;" temporaljson:"runner_job,omitzero,omitempty"`
}

func (r *Runner) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Runner{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
