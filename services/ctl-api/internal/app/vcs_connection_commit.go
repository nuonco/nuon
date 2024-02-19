package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type VCSConnectionCommit struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	VCSConnection   VCSConnection `json:"-"`
	VCSConnectionID string        `json:"component_config_connection_id" gorm:"notnull"`

	ComponentBuilds []ComponentBuild `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	SHA         string `json:"sha" gorm:"notnull"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	Message     string `json:"message"`
}

func (v *VCSConnectionCommit) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSConnectionID()

	if v.CreatedByID != "" {
		v.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if v.OrgID == "" {
		v.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
