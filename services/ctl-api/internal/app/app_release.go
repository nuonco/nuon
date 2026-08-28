package app

import (
	"time"

	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppReleaseStatus string

const (
	AppReleaseStatusPreparing AppReleaseStatus = "preparing"
	AppReleaseStatusReady     AppReleaseStatus = "ready"
	AppReleaseStatusError     AppReleaseStatus = "error"
)

type AppReleaseRuntime struct {
	RunnerImageURL string                               `json:"runner_image_url"`
	RunnerImageTag string                               `json:"runner_image_tag"`
	Platforms      map[string]AppReleasePlatformRuntime `json:"platforms"`
}

type AppReleasePlatformRuntime struct {
	PortalBinaryURL string `json:"portal_binary_url"`
	RunnerBinaryURL string `json:"runner_binary_url"`
}

type AppRelease struct {
	ID          string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string    `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account   `json:"-"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"notnull"`

	OrgID string `json:"-" gorm:"notnull;uniqueIndex:idx_app_release_identity"`
	Org   Org    `json:"-"`

	AppID       string    `json:"app_id" gorm:"notnull;uniqueIndex:idx_app_release_identity"`
	App         App       `json:"-"`
	AppConfigID string    `json:"app_config_id" gorm:"notnull"`
	AppConfig   AppConfig `json:"-"`

	SandboxBuildID    string                            `json:"sandbox_build_id" gorm:"notnull"`
	ComponentBuildIDs map[string]string                 `json:"component_build_ids" gorm:"type:jsonb;serializer:json;notnull"`
	Runbooks          []customermanaged.RunbookTemplate `json:"-" gorm:"type:jsonb;serializer:json;<-:create"`
	Runtime           AppReleaseRuntime                 `json:"runtime" gorm:"type:jsonb;serializer:json;<-:create;notnull"`
	RuntimeDigest     string                            `json:"runtime_digest" gorm:"notnull"`
	DefinitionsBlob   *blobstore.Blob                   `json:"-" gorm:"<-:create"`

	SchemaVersion     int              `json:"schema_version" gorm:"notnull"`
	SemanticDigest    string           `json:"semantic_digest" gorm:"notnull;uniqueIndex:idx_app_release_identity"`
	Status            AppReleaseStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string           `json:"status_description" gorm:"notnull"`

	Members     []AppReleaseMember            `json:"members,omitempty" gorm:"foreignKey:ReleaseID;constraint:OnDelete:RESTRICT;"`
	Packages    []ReleasePackage              `json:"packages,omitempty" gorm:"foreignKey:ReleaseID;constraint:OnDelete:RESTRICT;"`
	SourceFiles []customermanaged.ReleaseFile `json:"source_files,omitempty" gorm:"-"`
}

func (r *AppRelease) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewAppReleaseID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if r.DefinitionsBlob != nil {
		if err := r.DefinitionsBlob.BeforeCreate(tx); err != nil {
			return err
		}
	}
	return nil
}

func (r *AppRelease) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{{Name: indexes.Name(db, &AppRelease{}, "org_id"), Columns: []string{"org_id"}}}
}

type AppReleaseMember struct {
	ID    string `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	OrgID string `json:"-" gorm:"notnull;index"`

	ReleaseID string     `json:"release_id" gorm:"notnull;uniqueIndex:idx_app_release_member_name"`
	Release   AppRelease `json:"-"`

	Kind                        string         `json:"kind" gorm:"notnull;uniqueIndex:idx_app_release_member_name"`
	LogicalName                 string         `json:"logical_name" gorm:"notnull;uniqueIndex:idx_app_release_member_name"`
	ComponentConfigConnectionID string         `json:"component_config_connection_id,omitempty"`
	ComponentID                 string         `json:"component_id,omitempty"`
	ActionWorkflowID            string         `json:"action_workflow_id,omitempty"`
	AppSandboxConfigID          string         `json:"app_sandbox_config_id,omitempty"`
	BuildID                     string         `json:"build_id,omitempty"`
	ConfigDigest                string         `json:"config_digest,omitempty"`
	ConfigTOML                  string         `json:"config_toml,omitempty" gorm:"-"`
	ContentDigest               string         `json:"content_digest" gorm:"notnull"`
	SourceType                  string         `json:"source_type,omitempty"`
	SourceIdentity              map[string]any `json:"source_identity,omitempty" gorm:"type:jsonb;serializer:json"`
}

func (m *AppReleaseMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewAppReleaseMemberID()
	}
	if m.OrgID == "" {
		m.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type ReleasePackageStatus string

const (
	ReleasePackageStatusQueued     ReleasePackageStatus = "queued"
	ReleasePackageStatusPublishing ReleasePackageStatus = "publishing"
	ReleasePackageStatusActive     ReleasePackageStatus = "active"
	ReleasePackageStatusError      ReleasePackageStatus = "error"
)

const ReleasePackageFormatPortableOCI = "portable-oci"

type ReleasePackage struct {
	ID          string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string    `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account   `json:"-"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"notnull"`

	OrgID string `json:"-" gorm:"notnull;index"`
	Org   Org    `json:"-"`

	ReleaseID string     `json:"release_id" gorm:"notnull;uniqueIndex:idx_release_package_identity"`
	Release   AppRelease `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`

	Format          string `json:"format" gorm:"notnull;uniqueIndex:idx_release_package_identity"`
	TargetPlatform  string `json:"target_platform" gorm:"notnull;uniqueIndex:idx_release_package_identity"`
	PackageDigest   string `json:"package_digest" gorm:"notnull"`
	SchemaVersion   int    `json:"schema_version" gorm:"notnull"`
	ManifestDigest  string `json:"manifest_digest" gorm:"notnull"`
	PlanDigest      string `json:"plan_digest" gorm:"notnull"`
	OCIRootDigest   string `json:"oci_root_digest" gorm:"notnull"`
	OCIIndexDigest  string `json:"oci_index_digest"`
	ArchiveChecksum string `json:"archive_checksum" gorm:"notnull"`
	ArchiveSize     int64  `json:"archive_size" gorm:"notnull;type:bigint"`

	Status            ReleasePackageStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string               `json:"status_description" gorm:"notnull"`

	Members  []ReleasePackageMember  `json:"members,omitempty" gorm:"foreignKey:PackageID;constraint:OnDelete:RESTRICT;"`
	Replicas []ReleasePackageReplica `json:"replicas,omitempty" gorm:"foreignKey:PackageID;constraint:OnDelete:RESTRICT;"`
}

func (p *ReleasePackage) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = domains.NewReleasePackageID()
	}
	if p.CreatedByID == "" {
		p.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if p.OrgID == "" {
		p.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type ReleasePackageMember struct {
	ID    string `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	OrgID string `json:"-" gorm:"notnull;index"`

	PackageID string         `json:"package_id" gorm:"notnull;uniqueIndex:idx_release_package_member_name"`
	Package   ReleasePackage `json:"-"`

	Kind                        string         `json:"kind" gorm:"notnull;uniqueIndex:idx_release_package_member_name"`
	LogicalName                 string         `json:"logical_name" gorm:"notnull;uniqueIndex:idx_release_package_member_name"`
	ComponentConfigConnectionID string         `json:"component_config_connection_id,omitempty"`
	ComponentID                 string         `json:"component_id,omitempty"`
	ActionWorkflowID            string         `json:"action_workflow_id,omitempty"`
	AppSandboxConfigID          string         `json:"app_sandbox_config_id,omitempty"`
	ConfigDigest                string         `json:"config_digest,omitempty"`
	SourceType                  string         `json:"source_type,omitempty"`
	SourceIdentity              map[string]any `json:"source_identity,omitempty" gorm:"type:jsonb;serializer:json"`
	Repository                  string         `json:"repository,omitempty"`
	Digest                      string         `json:"digest" gorm:"notnull"`
	MediaType                   string         `json:"media_type" gorm:"notnull"`
	Size                        int64          `json:"size" gorm:"notnull;type:bigint"`
	PlatformOS                  string         `json:"platform_os,omitempty"`
	PlatformArchitecture        string         `json:"platform_architecture,omitempty"`
}

func (m *ReleasePackageMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewReleasePackageMemberID()
	}
	if m.OrgID == "" {
		m.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type ReleasePackageReplica struct {
	ID        string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt time.Time `json:"created_at" gorm:"notnull"`
	OrgID     string    `json:"-" gorm:"notnull;index"`

	PackageID string         `json:"package_id" gorm:"notnull;index"`
	Package   ReleasePackage `json:"-"`

	Provider        string     `json:"provider" gorm:"notnull"`
	Region          string     `json:"region"`
	StorageRef      string     `json:"-" gorm:"notnull"`
	StorageVersion  string     `json:"storage_version" gorm:"notnull"`
	ArchiveChecksum string     `json:"archive_checksum" gorm:"notnull"`
	Size            int64      `json:"size" gorm:"notnull;type:bigint"`
	VerifiedAt      *time.Time `json:"verified_at" gorm:"notnull"`
}

func (r *ReleasePackageReplica) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewReleasePackageReplicaID()
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
