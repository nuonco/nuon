package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallStackVersion struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	InstallID      string `json:"install_id" gorm:"notnull;default null"`
	InstallStackID string `json:"install_stack_id"`

	AppConfigID string `json:"app_config_id"`

	Status CompositeStatus `json:"composite_status" gorm:"type:jsonb"`

	Runs []InstallStackVersionRun `json:"runs"`

	Contents     []byte `json:"contents" gorm:"type:jsonb" swaggertype:"string"`
	Checksum     string `json:"checksum"`
	TemplateURL  string `json:"template_url"`
	PhoneHomeID  string `json:"phone_home_id"`
	PhoneHomeURL string `json:"phone_home_url"`

	// aws configuration parameters
	AWSBucketName string `json:"aws_bucket_name"`
	AWSBucketKey  string `json:"aws_bucket_key"`
	QuickLinkURL  string `json:"quick_link_url"`
}

func (a *InstallStackVersion) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppCfgID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
