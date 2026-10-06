package app

import (
	"time"

	appbundle "github.com/nuonco/nuon/pkg/appbundle"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
)

type AppBundleStatus string

const (
	AppBundleStatusQueued     AppBundleStatus = "queued"
	AppBundleStatusPublishing AppBundleStatus = "publishing"
	AppBundleStatusActive     AppBundleStatus = "active"
	AppBundleStatusError      AppBundleStatus = "error"
)

type AppBundlePlatformRuntime struct {
	RunnerBinaryURL string `json:"runner_binary_url"`
}

type AppBundleRuntime struct {
	RunnerImageURL string                              `json:"runner_image_url"`
	RunnerImageTag string                              `json:"runner_image_tag"`
	Platforms      map[string]AppBundlePlatformRuntime `json:"platforms"`
}

type AppBundle struct {
	ID          string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string    `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account   `json:"-"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull"`

	OrgID       string    `json:"-" gorm:"notnull;uniqueIndex:idx_app_bundle_identity"`
	Org         Org       `json:"-"`
	AppID       string    `json:"app_id" gorm:"notnull;uniqueIndex:idx_app_bundle_identity"`
	App         App       `json:"-"`
	AppConfigID string    `json:"app_config_id" gorm:"notnull;uniqueIndex:idx_app_bundle_identity"`
	AppConfig   AppConfig `json:"-"`

	SandboxBuildID    string                      `json:"-" gorm:"notnull;default:''"`
	ComponentBuildIDs map[string]string           `json:"-" gorm:"type:jsonb;serializer:json"`
	Runbooks          []appbundle.RunbookTemplate `json:"-" gorm:"type:jsonb;serializer:json;<-:create"`
	RunbooksDigest    string                      `json:"-" gorm:"notnull;default:'';uniqueIndex:idx_app_bundle_identity"`
	Runtime           AppBundleRuntime            `json:"runtime" gorm:"type:jsonb;serializer:json;<-:create;notnull"`
	RuntimeDigest     string                      `json:"runtime_digest" gorm:"notnull;default:'';uniqueIndex:idx_app_bundle_identity"`

	TargetPlatform string `json:"target_platform" gorm:"notnull;uniqueIndex:idx_app_bundle_identity"`
	SchemaVersion  int    `json:"schema_version" gorm:"notnull"`

	ManifestDigest string `json:"manifest_digest" gorm:"notnull;default:'';uniqueIndex:idx_app_bundle_identity"`
	OCIRootDigest  string `json:"oci_root_digest" gorm:"notnull;default:''"`
	OCIIndexDigest string `json:"oci_index_digest" gorm:"notnull;default:''"`

	TransportChecksum string     `json:"transport_checksum" gorm:"notnull;default:''"`
	Size              int64      `json:"size" gorm:"notnull;default:0;type:bigint"`
	StorageProvider   string     `json:"-" gorm:"notnull;default:''"`
	StorageBucket     string     `json:"-" gorm:"notnull;default:''"`
	StorageRegion     string     `json:"-" gorm:"notnull;default:''"`
	StorageRef        string     `json:"-" gorm:"notnull;default:''"`
	StorageVersion    string     `json:"-" gorm:"notnull;default:''"`
	VerifiedAt        *time.Time `json:"verified_at" gorm:""`

	Status            AppBundleStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string          `json:"status_description" gorm:"notnull"`
}

func (b *AppBundle) HasVerifiedArchive() bool {
	return b.VerifiedAt != nil && b.ManifestDigest != "" && b.OCIRootDigest != "" && b.OCIIndexDigest != "" &&
		b.TransportChecksum != "" && b.StorageProvider != "" && b.StorageBucket != "" && b.StorageRef != "" && b.StorageVersion != ""
}

func (b *AppBundle) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = domains.NewAppBundleID()
	}
	if b.CreatedByID == "" {
		b.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if b.OrgID == "" {
		b.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (b *AppBundle) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{{Name: indexes.Name(db, &AppBundle{}, "org_id"), Columns: []string{"org_id"}}}
}
