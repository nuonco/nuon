package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type ReleasePackageStatus string

const (
	ReleasePackageStatusQueued     ReleasePackageStatus = "queued"
	ReleasePackageStatusPublishing ReleasePackageStatus = "publishing"
	ReleasePackageStatusActive     ReleasePackageStatus = "active"
	ReleasePackageStatusError      ReleasePackageStatus = "error"
)

const ReleasePackageFormatPortableOCI = "portable-oci"

type ReleasePackage struct {
	ID          string     `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string     `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"notnull"`
	OrgID       string     `json:"-" gorm:"notnull;index"`
	ReleaseID   string     `json:"release_id" gorm:"notnull;uniqueIndex:idx_release_package_identity"`
	Release     AppRelease `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`

	Format            string               `json:"format" gorm:"notnull;uniqueIndex:idx_release_package_identity"`
	TargetPlatform    string               `json:"target_platform" gorm:"notnull;uniqueIndex:idx_release_package_identity"`
	PackageDigest     string               `json:"package_digest" gorm:"notnull"`
	SchemaVersion     int                  `json:"schema_version" gorm:"notnull"`
	ManifestDigest    string               `json:"manifest_digest" gorm:"notnull"`
	PlanDigest        string               `json:"plan_digest" gorm:"notnull"`
	OCIRootDigest     string               `json:"oci_root_digest" gorm:"notnull"`
	OCIIndexDigest    string               `json:"oci_index_digest"`
	ArchiveChecksum   string               `json:"archive_checksum" gorm:"notnull"`
	ArchiveSize       int64                `json:"archive_size" gorm:"notnull;type:bigint"`
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
	ID                          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	OrgID                       string         `json:"-" gorm:"notnull;index"`
	PackageID                   string         `json:"package_id" gorm:"notnull;uniqueIndex:idx_release_package_member_name"`
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
	ID              string     `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt       time.Time  `json:"created_at" gorm:"notnull"`
	OrgID           string     `json:"-" gorm:"notnull;index"`
	PackageID       string     `json:"package_id" gorm:"notnull;index"`
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
