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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"index:install_component_group,unique;notnull" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `faker:"-" json:"-" temporaljson:"install,omitzero,omitempty"`

	ComponentID string    `json:"component_id,omitzero" gorm:"index:install_component_group,unique;notnull" temporaljson:"component_id,omitzero,omitempty"`
	Component   Component `faker:"-" json:"component,omitzero" temporaljson:"component,omitzero,omitempty"`

	InstallDeploys     []InstallDeploy    `faker:"-" gorm:"constraint:OnDelete:CASCADE;" json:"install_deploys,omitzero" temporaljson:"install_deploys,omitzero,omitempty"`
	TerraformWorkspace TerraformWorkspace `json:"terraform_workspace,omitzero" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"terraform_workspace,omitzero,omitempty"`

	Links     map[string]any `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`
	HelmChart HelmChart      `json:"helm_chart" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"helm_chart,omitzero,omitempty"`

	Status            InstallComponentStatus `json:"status,omitzero" gorm:"default:''" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                 `json:"status_description,omitzero" gorm:"default:''" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus        `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`
}

type InstallComponentSummary struct {
	ID                      string                     `json:"id"`
	ComponentID             string                     `json:"component_id"`
	ComponentName           string                     `json:"component_name"`
	DeployStatus            InstallDeployStatus        `json:"deploy_status"`
	DeployStatusDescription string                     `json:"deploy_status_description"`
	BuildStatus             ComponentBuildStatus       `json:"build_status"`
	BuildStatusDescription  string                     `json:"build_status_description"`
	ComponentConfig         *ComponentConfigConnection `json:"component_config"`
	Dependencies            []Component                `json:"dependencies"`
}

func (c *InstallComponent) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewInstallComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *InstallComponent) AfterQuery(tx *gorm.DB) error {
	if c.StatusV2.Status != "" {
		c.Status = InstallComponentStatus(c.StatusV2.Status)
		c.StatusDescription = c.StatusV2.StatusHumanDescription
	}

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
