package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type ImageRepository struct {
	ID                string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt         time.Time             `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time             `gorm:"not null" json:"updated_at"`
	DeletedAt         soft_delete.DeletedAt `json:"-"`
	OrgID             string                `gorm:"not null;uniqueIndex:idx_image_repositories_org_key,where:deleted_at = 0" json:"org_id"`
	Org               Org                   `json:"-"`
	Platform          CloudPlatform         `gorm:"not null" json:"platform" swaggertype:"string" enums:"aws,gcp,azure"`
	Repository        string                `gorm:"not null" json:"repository"`
	CanonicalKey      string                `gorm:"not null;uniqueIndex:idx_image_repositories_org_key,where:deleted_at = 0" json:"-"`
	ImageURLPrefix    string                `gorm:"not null" json:"image_url_prefix"`
	CloudConnectionID string                `gorm:"not null;index" json:"cloud_connection_id"`
	CloudConnection   CloudConnection       `json:"cloud_connection"`
}

func (r *ImageRepository) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewImageRepositoryID()
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
