package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
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
	InstallComponentStatusDisabled  InstallComponentStatus = "disabled"
)

// InstallComponentHealthStatus is the live-health axis of an install
// component, separate from and never overwriting the deploy/lifecycle Status.
type InstallComponentHealthStatus string

const (
	InstallComponentHealthStatusUnset         InstallComponentHealthStatus = ""
	InstallComponentHealthStatusHealthy       InstallComponentHealthStatus = "healthy"
	InstallComponentHealthStatusProgressing   InstallComponentHealthStatus = "progressing"
	InstallComponentHealthStatusDegraded      InstallComponentHealthStatus = "degraded"
	InstallComponentHealthStatusUnhealthy     InstallComponentHealthStatus = "unhealthy"
	InstallComponentHealthStatusUnknown       InstallComponentHealthStatus = "unknown"
	InstallComponentHealthStatusNotApplicable InstallComponentHealthStatus = "not-applicable"
)

// HasDeployed reports whether a deploy has actually run, so health is never
// computed pre-deploy (a passing probe would lie healthy). Failed counts too.
func (s InstallComponentStatus) HasDeployed() bool {
	switch s {
	case InstallComponentStatusActive,
		InstallComponentStatusNoop,
		InstallComponentStatusError:
		return true
	}
	return false
}

// EverDeployed reports whether a deploy has ever run, staying true mid-redeploy
// (previous workload still serving) using a prior health verdict as proof.
func (ic *InstallComponent) EverDeployed() bool {
	// Torn down = deliberately deleted; evaluating it would alert on a wanted
	// removal. DeleteFailed is the opposite: lingering resources you want to see.
	switch ic.Status {
	case InstallComponentStatusInactive, InstallComponentStatusDeleted:
		return false
	}
	if ic.Status.HasDeployed() {
		return true
	}
	return ic.HealthStatus != InstallComponentHealthStatusUnset &&
		ic.HealthStatus != InstallComponentHealthStatusNotApplicable
}

// ClusterHealthSeen reports whether the runner has ever observed this
// component's resources — distinguishes "went unobservable" from "probe-only".
func (ic *InstallComponent) ClusterHealthSeen() bool {
	seen, _ := ic.HealthStatusV2.Metadata["cluster_seen"].(bool)
	return seen
}

// IsBadHealth reports whether a verdict is actionable — the only verdicts
// that notify. Progressing is transient; unknown is lost visibility, not alerts.
func (s InstallComponentHealthStatus) IsBadHealth() bool {
	return s == InstallComponentHealthStatusDegraded || s == InstallComponentHealthStatusUnhealthy
}

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
	DriftedObject      DriftedObject      `faker:"-" gorm:"-;" json:"drifted_object,omitzero" temporaljson:"drifted_object,omitzero,omitempty"`

	Links     map[string]any `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`
	HelmChart HelmChart      `json:"helm_chart" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"helm_chart,omitzero,omitempty"`

	Status            InstallComponentStatus `json:"status,omitzero" gorm:"default:''" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                 `json:"status_description,omitzero" gorm:"default:''" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus        `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`

	HealthStatus            InstallComponentHealthStatus `json:"health_status,omitzero" gorm:"default:''" swaggertype:"string" temporaljson:"health_status,omitzero,omitempty"`
	HealthStatusDescription string                       `json:"health_status_description,omitzero" gorm:"-" temporaljson:"health_status_description,omitzero,omitempty"`
	HealthStatusV2          CompositeStatus              `json:"health_status_v2,omitzero" gorm:"type:jsonb" temporaljson:"health_status_v2,omitzero,omitempty"`

	// Enabled is the resolved enabled/disabled state for a toggleable component
	// (from the synthetic enabled input, falling back to default_enabled); nil otherwise.
	Enabled *bool `json:"enabled,omitempty" gorm:"-" temporaljson:"-" swaggertype:"boolean" extensions:"x-nullable"`
}

type InstallComponentSummary struct {
	ID                      string                       `json:"id"`
	ComponentID             string                       `json:"component_id"`
	ComponentName           string                       `json:"component_name"`
	DeployStatus            InstallDeployStatus          `json:"deploy_status"`
	DeployStatusDescription string                       `json:"deploy_status_description"`
	BuildStatus             ComponentBuildStatus         `json:"build_status"`
	BuildStatusDescription  string                       `json:"build_status_description"`
	ComponentConfig         *ComponentConfigConnection   `json:"component_config"`
	Dependencies            []Component                  `json:"dependencies"`
	DriftedStatus           bool                         `json:"drifted_status"`
	HealthStatus            InstallComponentHealthStatus `json:"health_status"`
}

func (c *InstallComponent) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallComponent{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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

	if c.HealthStatusV2.Status != "" {
		c.HealthStatus = InstallComponentHealthStatus(c.HealthStatusV2.Status)
		c.HealthStatusDescription = c.HealthStatusV2.StatusHumanDescription
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
