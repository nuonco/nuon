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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	ComponentBuildID string         `json:"build_id" temporaljson:"component_build_id,omitzero,omitempty"`
	ComponentBuild   ComponentBuild `json:"build" swaggerignore:"true" temporaljson:"component_build,omitzero,omitempty"`

	TotalComponentReleaseSteps int                    `json:"total_release_steps" gorm:"-" temporaljson:"total_component_release_steps,omitzero,omitempty"`
	ComponentReleaseSteps      []ComponentReleaseStep `json:"release_steps,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_release_steps,omitzero,omitempty"`

	Status            ReleaseStatus `json:"status" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string        `json:"status_description" temporaljson:"status_description,omitzero,omitempty"`
}

func (a *ComponentRelease) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewReleaseID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
