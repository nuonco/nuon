package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallStackVersion struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID      string `json:"install_id" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`
	InstallStackID string `json:"install_stack_id" temporaljson:"install_stack_id,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id" temporaljson:"app_config_id,omitzero,omitempty"`

	Status CompositeStatus `json:"composite_status" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	Runs []InstallStackVersionRun `json:"runs" temporaljson:"runs,omitzero,omitempty"`

	Contents     []byte `json:"contents" gorm:"type:jsonb" swaggertype:"string" temporaljson:"contents,omitzero,omitempty"`
	Checksum     string `json:"checksum" temporaljson:"checksum,omitzero,omitempty"`
	TemplateURL  string `json:"template_url" temporaljson:"template_url,omitzero,omitempty"`
	PhoneHomeID  string `json:"phone_home_id" temporaljson:"phone_home_id,omitzero,omitempty"`
	PhoneHomeURL string `json:"phone_home_url" temporaljson:"phone_home_url,omitzero,omitempty"`

	// aws configuration parameters
	AWSBucketName string `json:"aws_bucket_name" temporaljson:"aws_bucket_name,omitzero,omitempty"`
	AWSBucketKey  string `json:"aws_bucket_key" temporaljson:"aws_bucket_key,omitzero,omitempty"`
	QuickLinkURL  string `json:"quick_link_url" temporaljson:"quick_link_url,omitzero,omitempty"`
}

func (a *InstallStackVersion) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewInstallStackVersionID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
