package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallSandboxStatus string

const (
	InstallSandboxStatusDeleted      InstallSandboxStatus = "deleted"
	InstallSandboxStatusDeleteFailed InstallSandboxStatus = "delete_failed"

	// Synced from sandbow runs
	InstallSandboxStatusActive         InstallSandboxStatus = "active"
	InstallSandboxStatusError          InstallSandboxStatus = "error"
	InstallSandboxStatusQueued         InstallSandboxStatus = "queued"
	InstallSandboxStatusDeprovisioned  InstallSandboxStatus = "deprovisioned"
	InstallSandboxStatusDeprovisioning InstallSandboxStatus = "deprovisioning"
	InstallSandboxStatusProvisioning   InstallSandboxStatus = "provisioning"
	InstallSandboxStatusReprovisioning InstallSandboxStatus = "reprovisioning"
	InstallSandboxStatusAccessError    InstallSandboxStatus = "access_error"
	InstallSandboxStatusUnknown        InstallSandboxStatus = "unknown"
	InstallSandboxStatusEmpty          InstallSandboxStatus = "empty"
)

type InstallSandbox struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	InstallID string  `json:"install_id" gorm:"notnull"`
	Install   Install `faker:"-" json:"-"`

	Status            InstallSandboxStatus `json:"status" gorm:"not null;default null" swaggertype:"string"`
	StatusDescription string               `json:"status_description" gorm:"not null;default null"`

	InstallSandboxRuns []InstallSandboxRun `json:"install_sandbox_runs,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
}

func (c *InstallSandbox) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewInstallSandboxID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (c *InstallSandbox) AfterQuery(tx *gorm.DB) error {
	return nil
}

func (c *InstallSandbox) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallSandbox{}, "uq"),
			Columns: []string{
				"install_id",
				"deleted_at",
			},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
	}
}

func SandboxRunStatusToInstallSandboxStatus(status SandboxRunStatus) InstallSandboxStatus {
	switch status {
	case SandboxRunStatusActive:
		return InstallSandboxStatusActive
	case SandboxRunStatusError:
		return InstallSandboxStatusError
	case SandboxRunStatusQueued:
		return InstallSandboxStatusQueued
	case SandboxRunStatusDeprovisioned:
		return InstallSandboxStatusDeprovisioned
	case SandboxRunStatusDeprovisioning:
		return InstallSandboxStatusDeprovisioning
	case SandboxRunStatusProvisioning:
		return InstallSandboxStatusProvisioning
	case SandboxRunStatusReprovisioning:
		return InstallSandboxStatusReprovisioning
	case SandboxRunStatusAccessError:
		return InstallSandboxStatusAccessError
	case SandboxRunStatusUnknown:
		return InstallSandboxStatusUnknown
	case SandboxRunStatusEmpty:
		return InstallSandboxStatusEmpty
	default:
		return InstallSandboxStatusUnknown
	}
}
