package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ComponentBuildStatus string

const (
	ComponentBuildStatusPlanning     ComponentBuildStatus = "planning"
	ComponentBuildStatusError        ComponentBuildStatus = "error"
	ComponentBuildStatusBuilding     ComponentBuildStatus = "building"
	ComponentBuildStatusActive       ComponentBuildStatus = "active"
	ComponentBuildStatusDeleting     ComponentBuildStatus = "deleting"
	ComponentBuildStatusPolicyFailed ComponentBuildStatus = "policy_failed"
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

	PolicyReports []PolicyReport `json:"policy_reports,omitzero" gorm:"polymorphic:Owner;polymorphicValue:component_builds" temporaljson:"policy_reports,omitzero,omitempty"`

	// DEPRECATED: will retain the field to connect against the last component config connection that set this build
	ComponentConfigConnectionID string                    `json:"component_config_connection_id,omitzero" gorm:"notnull" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"component_config_connection,omitzero" temporaljson:"component_config_connection,omitzero,omitempty"`

	VCSConnectionCommitID *string              `json:"-" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitzero" temporaljson:"vcs_connection_commit,omitzero,omitempty"`

	ComponentReleases []ComponentRelease `json:"releases,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_releases,omitzero,omitempty"`
	InstallDeploys    []InstallDeploy    `json:"install_deploys,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_deploys,omitzero,omitempty"`

	Status            ComponentBuildStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string               `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus      `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`

	GitRef *string `json:"git_ref,omitzero" temporaljson:"git_ref,omitzero,omitempty"`

	// Read-only fields set on the object to de-nest data
	ComponentID            string `gorm:"-" json:"component_id,omitzero" temporaljson:"component_id,omitzero,omitempty"`
	ComponentName          string `gorm:"-" json:"component_name,omitzero" temporaljson:"component_name,omitzero,omitempty"`
	ComponentConfigVersion int    `gorm:"-" json:"component_config_version,omitzero" temporaljson:"component_config_version,omitzero,omitempty"`

	// checksum of our intermediate component config
	Checksum string `json:"checksum,omitzero" gorm:"default null" temporaljson:"checksum,omitzero,omitempty"`

	// Source identity for image-type builds.
	//
	// SourceRef is what the user wrote in the spec, e.g. "nginx:1.25.3" or
	// "myimage@sha256:...". Always populated for image-type builds so we have a
	// permanent record of what was requested at build time.
	SourceRef string `json:"source_ref,omitzero" gorm:"default null" temporaljson:"source_ref,omitzero,omitempty"`
	// SourceImage is the repository portion of SourceRef without tag/digest, e.g. "nginx".
	SourceImage string `json:"source_image,omitzero" gorm:"default null" temporaljson:"source_image,omitzero,omitempty"`
	// ResolvedTag is the tag the runner actually pulled from. For digest-pinned
	// refs this is empty. For mutable/semver refs this is the concrete tag the
	// runner selected (e.g. "1.25.5" even if SourceRef pinned "1.25.3" with a
	// "~1.25.0" update_policy constraint).
	ResolvedTag string `json:"resolved_tag,omitzero" gorm:"default null" temporaljson:"resolved_tag,omitzero,omitempty"`
	// SourceDigest is the manifest list digest of the resolved source ref,
	// e.g. "sha256:abc...". This is the canonical content address of what was
	// pulled and is used for build dedup.
	SourceDigest string `json:"source_digest,omitzero" gorm:"default null" temporaljson:"source_digest,omitzero,omitempty"`
	// SourceMediaType records the media type of the resolved manifest (image,
	// image index, OCI artifact, etc.) for downstream rendering decisions.
	SourceMediaType string `json:"source_media_type,omitzero" gorm:"default null" temporaljson:"source_media_type,omitzero,omitempty"`
	// ResolvedAt is when the runner resolved SourceRef to SourceDigest.
	ResolvedAt *time.Time `json:"resolved_at,omitzero" gorm:"default null" temporaljson:"resolved_at,omitzero,omitempty"`
	// NoOp is true when the runner detected SourceDigest matches the previous
	// build's SourceDigest and skipped the artifact push.
	//
	// Downstream contract:
	//   - The build is still marked Active because the bytes it represents
	//     are deployable (they live in the install registry under the prior
	//     build that pushed them).
	//   - No new install deploys are auto-queued for a NoOp build; the
	//     dep-aware deploy path handles fan-out for installs that depend
	//     on the underlying image.
	//   - pollForDeployableBuild treats NoOp builds as Active without any
	//     special-casing because the deployable artifact at the same
	//     SourceDigest is already present in the install registry from the
	//     prior build.
	NoOp bool `json:"no_op,omitzero" gorm:"default false" temporaljson:"no_op,omitzero,omitempty"`

	// QueueSignal is the signal enqueued when this build was created via the queue path
	QueueSignal *QueueSignal `json:"queue_signal,omitempty" gorm:"polymorphic:Owner;" temporaljson:"queue_signal,omitzero,omitempty"`
}

func (c *ComponentBuild) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &ComponentBuild{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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

	if c.StatusV2.Status != "" {
		c.Status = ComponentBuildStatus(c.StatusV2.Status)
		c.StatusDescription = c.StatusV2.StatusHumanDescription
	}

	// Surface NoOp on Active builds so consumers (CLI, dashboard, status
	// columns) can immediately tell a build was a content-address dedup hit
	// without having to inspect SourceDigest history themselves.
	if c.NoOp && c.Status == ComponentBuildStatusActive {
		c.StatusDescription = "no-op: source unchanged from previous build (reusing prior artifact)"
	}

	return nil
}

// IsNoOp returns true when the runner detected this build's resolved source
// digest matched the previous active build's digest and skipped the artifact
// push. NoOp builds are deployable by virtue of the prior build's artifact
// already living in the install registry at the same digest; they should
// never trigger new install deploys on their own.
func (c *ComponentBuild) IsNoOp() bool {
	return c.NoOp
}
