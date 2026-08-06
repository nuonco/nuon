package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

const (
	AirgapBundleStatusQueued     = "queued"
	AirgapBundleStatusPublishing = "publishing"
	AirgapBundleStatusActive     = "active"
	AirgapBundleStatusError      = "error"
)

type AirgapBundle struct {
	ID          string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string    `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account   `json:"-"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull"`

	OrgID             string            `json:"-" gorm:"notnull;uniqueIndex:idx_airgap_bundle_identity"`
	Org               Org               `json:"-"`
	AppID             string            `json:"app_id" gorm:"notnull;uniqueIndex:idx_airgap_bundle_identity"`
	App               App               `json:"-"`
	AppConfigID       string            `json:"app_config_id" gorm:"notnull;uniqueIndex:idx_airgap_bundle_identity"`
	AppConfig         AppConfig         `json:"-"`
	SandboxBuildID    string            `json:"-"`
	ComponentBuildIDs map[string]string `json:"-" gorm:"type:jsonb;serializer:json"`

	TargetPlatform    string `json:"target_platform" gorm:"notnull;uniqueIndex:idx_airgap_bundle_identity"`
	SchemaVersion     int    `json:"schema_version" gorm:"notnull"`
	ManifestDigest    string `json:"manifest_digest" gorm:"notnull;uniqueIndex:idx_airgap_bundle_identity"`
	OCIRootDigest     string `json:"oci_root_digest" gorm:"notnull"`
	TransportChecksum string `json:"transport_checksum" gorm:"notnull"`
	Size              int64  `json:"size" gorm:"notnull;type:bigint"`
	Status            string `json:"status" gorm:"notnull"`
	StatusDescription string `json:"status_description" gorm:"notnull"`

	Artifacts []AirgapBundleArtifact         `json:"artifacts,omitempty" gorm:"foreignKey:BundleID;constraint:OnDelete:RESTRICT;"`
	Replicas  []AirgapBundleTransportReplica `json:"replicas,omitempty" gorm:"foreignKey:BundleID;constraint:OnDelete:RESTRICT;"`
}

func (b *AirgapBundle) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = domains.NewAirgapBundleID()
	}
	if b.CreatedByID == "" {
		b.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if b.OrgID == "" {
		b.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (b *AirgapBundle) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{{Name: indexes.Name(db, &AirgapBundle{}, "org_id"), Columns: []string{"org_id"}}}
}

type AirgapBundleTransportReplica struct {
	ID        string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt time.Time `json:"created_at" gorm:"notnull"`
	OrgID     string    `json:"-" gorm:"notnull;index"`

	BundleID string       `json:"bundle_id" gorm:"notnull;index"`
	Bundle   AirgapBundle `json:"-"`

	Provider          string     `json:"provider" gorm:"notnull"`
	Region            string     `json:"region"`
	StorageRef        string     `json:"-" gorm:"notnull"`
	StorageVersion    string     `json:"storage_version" gorm:"notnull"`
	TransportChecksum string     `json:"transport_checksum" gorm:"notnull"`
	Size              int64      `json:"size" gorm:"notnull;type:bigint"`
	VerifiedAt        *time.Time `json:"verified_at" gorm:"notnull"`
}

func (r *AirgapBundleTransportReplica) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewAirgapBundleReplicaID()
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type AirgapBundleArtifact struct {
	ID    string `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	OrgID string `json:"-" gorm:"notnull;index"`

	BundleID string       `json:"bundle_id" gorm:"notnull;uniqueIndex:idx_airgap_bundle_artifact_name"`
	Bundle   AirgapBundle `json:"-"`

	Kind                        string         `json:"kind" gorm:"notnull;uniqueIndex:idx_airgap_bundle_artifact_name"`
	LogicalName                 string         `json:"logical_name" gorm:"notnull;uniqueIndex:idx_airgap_bundle_artifact_name"`
	ComponentConfigConnectionID string         `json:"component_config_connection_id,omitempty"`
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

func (a *AirgapBundleArtifact) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAirgapBundleArtifactID()
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
