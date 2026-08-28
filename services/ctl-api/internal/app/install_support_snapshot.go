package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/supportsnapshot"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type InstallSupportSnapshot struct {
	ID          string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string    `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull"`

	OrgID     string  `json:"-" gorm:"notnull;uniqueIndex:idx_customer_managed_support_snapshot_archive"`
	Org       Org     `json:"-"`
	InstallID string  `json:"install_id" gorm:"notnull;index"`
	Install   Install `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`

	ArchiveSHA256 string    `json:"archive_sha256" gorm:"notnull;uniqueIndex:idx_customer_managed_support_snapshot_archive"`
	ArchiveSize   int64     `json:"archive_size" gorm:"notnull;type:bigint"`
	SchemaVersion int       `json:"schema_version" gorm:"notnull"`
	CapturedAt    time.Time `json:"captured_at" gorm:"notnull"`

	StorageProvider string `json:"-" gorm:"notnull"`
	StorageRegion   string `json:"-"`
	StorageRef      string `json:"-" gorm:"notnull"`
	StorageVersion  string `json:"-" gorm:"notnull"`

	IntegrityStatus   string                   `json:"integrity_status" gorm:"notnull"`
	AssociationStatus string                   `json:"association_status" gorm:"notnull"`
	Manifest          supportsnapshot.Manifest `json:"manifest" gorm:"type:jsonb;serializer:json;<-:create"`
	SnapshotBlob      *blobstore.Blob          `json:"-" gorm:"<-:create"`
}

func (s *InstallSupportSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = domains.NewInstallSupportSnapshotID()
	}
	if s.CreatedByID == "" {
		s.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if s.OrgID == "" {
		s.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if s.SnapshotBlob != nil {
		if err := s.SnapshotBlob.BeforeCreate(tx); err != nil {
			return err
		}
	}
	return nil
}
