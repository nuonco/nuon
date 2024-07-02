package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type ComponentReleaseStrategy string

const (
	// Parallel means that all steps start at the same time
	ComponentReleaseStrategyParallel ComponentReleaseStrategy = "parallel"

	// Sync with delay splits the installs into steps (based on count/step), and then just waits the period of time
	ComponentReleaseStrategySyncWithDelay ComponentReleaseStrategy = "sync_with_delay"
)

type ReleaseStatus string

const (
	ReleaseStatusPlanning       ReleaseStatus = "planning"
	ReleaseStatusError          ReleaseStatus = "error"
	ReleaseStatusActive         ReleaseStatus = "active"
	ReleaseStatusProvisioning   ReleaseStatus = "provisioning"
	ReleaseStatusDeprovisioning ReleaseStatus = "deprovisioning"

	ReleaseStatusSyncing   ReleaseStatus = "syncing"
	ReleaseStatusExecuting ReleaseStatus = "executing"
)

type ComponentRelease struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	ComponentBuildID string         `json:"build_id"`
	ComponentBuild   ComponentBuild `json:"build" swaggerignore:"true"`

	TotalComponentReleaseSteps int                    `json:"total_release_steps" gorm:"-"`
	ComponentReleaseSteps      []ComponentReleaseStep `json:"release_steps,omitempty" gorm:"constraint:OnDelete:CASCADE;"`

	Status            ReleaseStatus `json:"status" swaggertype:"string"`
	StatusDescription string        `json:"status_description"`
}

func (a *ComponentRelease) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewReleaseID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
