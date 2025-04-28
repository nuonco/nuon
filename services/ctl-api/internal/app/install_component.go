package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallComponentStatus string

const (
	InstallComponentStatusUnset        InstallComponentStatus = ""
	InstallComponentStatusDeleted      InstallComponentStatus = "deleted"
	InstallComponentStatusDeleteFailed InstallComponentStatus = "delete_failed"
	InstallComponentStatusQueued       InstallComponentStatus = "queued"

	// all legacy statuses that could be set from install deploy
	InstallComponentStatusActive    InstallComponentStatus = "active"
	InstallComponentStatusInactive  InstallComponentStatus = "inactive"
	InstallComponentStatusError     InstallComponentStatus = "error"
	InstallComponentStatusNoop      InstallComponentStatus = "noop"
	InstallComponentStatusPlanning  InstallComponentStatus = "planning"
	InstallComponentStatusSyncing   InstallComponentStatus = "syncing"
	InstallComponentStatusExecuting InstallComponentStatus = "executing"
	InstallComponentStatusUnknown   InstallComponentStatus = "unknown"
	InstallComponentStatusPending   InstallComponentStatus = "pending"
)

type InstallComponent struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id" gorm:"index:install_component_group,unique;notnull" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `faker:"-" json:"-" temporaljson:"install,omitzero,omitempty"`

	ComponentID string    `json:"component_id" gorm:"index:install_component_group,unique;notnull" temporaljson:"component_id,omitzero,omitempty"`
	Component   Component `faker:"-" json:"component" temporaljson:"component,omitzero,omitempty"`

	InstallDeploys     []InstallDeploy    `faker:"-" gorm:"constraint:OnDelete:CASCADE;" json:"install_deploys" temporaljson:"install_deploys,omitzero,omitempty"`
	TerraformWorkspace TerraformWorkspace `json:"terraform_workspace" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"terraform_workspace,omitzero,omitempty"`

	Status            InstallComponentStatus `json:"status" gorm:"default:''" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                 `json:"status_description" gorm:"default:''" temporaljson:"status_description,omitzero,omitempty"`
}

func (c *InstallComponent) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewInstallComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *InstallComponent) AfterQuery(tx *gorm.DB) error {
	if c.Status == InstallComponentStatusUnset && len(c.InstallDeploys) > 0 {
		// TODO: we shouldn't need this check, after we migrated all statuses from latest deploys
		status := DeployStatusToComponentStatus(c.InstallDeploys[0].Status)
		c.Status = status
		return nil
	}

	c.Status = InstallComponentStatusUnknown

	return nil
}

func DeployStatusToComponentStatus(status InstallDeployStatus) InstallComponentStatus {
	switch status {
	case InstallDeployStatusActive:
		return InstallComponentStatusActive
	case InstallDeployStatusInactive:
		return InstallComponentStatusDeleted
	case InstallDeployStatusError:
		return InstallComponentStatusError
	case InstallDeployStatusPlanning:
		return InstallComponentStatusPlanning
	case InstallDeployStatusSyncing:
		return InstallComponentStatusSyncing
	case InstallDeployStatusExecuting:
		return InstallComponentStatusExecuting
	case InstallDeployStatusUnknown:
		return InstallComponentStatusUnknown
	case InstallDeployStatusPending:
		return InstallComponentStatusPending
	default:
		return InstallComponentStatusUnknown
	}
}
