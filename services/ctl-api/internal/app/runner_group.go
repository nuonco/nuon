package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
)

type RunnerGroupType string

const (
	RunnerGroupTypeInstall RunnerGroupType = "install"
	RunnerGroupTypeOrg     RunnerGroupType = "org"
)

type RunnerGroup struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `gorm:"index:idx_runner_group_owner" json:"-"`

	OrgID string `json:"org_id" gorm:"default null;not null"`

	// parent can org, install or in the future, builtin runner group
	OwnerID   string `json:"owner_id" gorm:"index:idx_runner_group_owner;notnull;default null"`
	OwnerType string `json:"owner_type" gorm:"notnull;default null"`

	Runners  []Runner            `json:"runners" gorm:"constraint:OnDelete:CASCADE;"`
	Settings RunnerGroupSettings `json:"settings" gorm:"constraint:OnDelete:CASCADE;"`
	Type     RunnerGroupType     `json:"type" gorm:"notnull;defaultnull"`
	Platform AppRunnerType       `json:"platform" gorm:"notnull;defaultnull"`
}

func (r *RunnerGroup) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerGroupID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r *RunnerGroup) EventLoops() []bulk.EventLoop {
	evs := make([]bulk.EventLoop, 0)
	for _, runner := range r.Runners {
		evs = append(evs, bulk.EventLoop{
			Namespace: "runners",
			ID:        runner.ID,
		})
	}

	return evs
}
