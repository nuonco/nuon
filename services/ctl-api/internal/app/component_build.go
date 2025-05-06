package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type ComponentBuildStatus string

const (
	ComponentBuildStatusPlanning ComponentBuildStatus = "planning"
	ComponentBuildStatusError    ComponentBuildStatus = "error"
	ComponentBuildStatusBuilding ComponentBuildStatus = "building"
	ComponentBuildStatusActive   ComponentBuildStatus = "active"
	ComponentBuildStatusDeleting ComponentBuildStatus = "deleting"
)

type ComponentBuild struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// runner details
	RunnerJob RunnerJob `json:"runner_job,omitzero" gorm:"polymorphic:Owner;" temporaljson:"runner_job,omitzero,omitempty"`

	LogStream LogStream `json:"log_stream,omitzero" gorm:"polymorphic:Owner;" temporaljson:"log_stream,omitzero,omitempty"`

	ComponentConfigConnectionID string                    `json:"component_config_connection_id,omitzero" gorm:"notnull" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-" temporaljson:"component_config_connection,omitzero,omitempty"`

	VCSConnectionCommitID *string              `json:"-" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitzero" temporaljson:"vcs_connection_commit,omitzero,omitempty"`

	ComponentReleases []ComponentRelease `json:"releases,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_releases,omitzero,omitempty"`
	InstallDeploys    []InstallDeploy    `json:"install_deploys,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_deploys,omitzero,omitempty"`

	Status            ComponentBuildStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string               `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
	GitRef            *string              `json:"git_ref,omitzero" temporaljson:"git_ref,omitzero,omitempty"`

	// Read-only fields set on the object to de-nest data
	ComponentID            string `gorm:"-" json:"component_id,omitzero" temporaljson:"component_id,omitzero,omitempty"`
	ComponentName          string `gorm:"-" json:"component_name,omitzero" temporaljson:"component_name,omitzero,omitempty"`
	ComponentConfigVersion int    `gorm:"-" json:"component_config_version,omitzero" temporaljson:"component_config_version,omitzero,omitempty"`
}

func (c *ComponentBuild) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewBuildID()
	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *ComponentBuild) AfterQuery(tx *gorm.DB) error {
	c.ComponentID = c.ComponentConfigConnection.ComponentID
	c.ComponentName = c.ComponentConfigConnection.Component.Name
	c.ComponentConfigVersion = c.ComponentConfigConnection.Version
	return nil
}
