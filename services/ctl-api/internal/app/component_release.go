package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ComponentReleaseStrategy string

const (
	// Parallel means that all steps start at the same time
	ComponentReleaseStrategyParallel ComponentReleaseStrategy = "parallel"

	// Sync with delay splits the installs into steps (based on count/step), and then just waits the period of time
	ComponentReleaseStrategySyncWithDelay ComponentReleaseStrategy = "sync_with_delay"
)

type ComponentRelease struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	ComponentBuildID string         `json:"build_id"`
	ComponentBuild   ComponentBuild `json:"build"`

	TotalComponentReleaseSteps int                    `json:"total_release_steps" gorm:"-"`
	ComponentReleaseSteps      []ComponentReleaseStep `json:"release_steps,omitempty" gorm:"constraint:OnDelete:CASCADE;"`

	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`
}

func (a *ComponentRelease) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewReleaseID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
