package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallStackVersion struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID      string `json:"install_id,omitzero" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`
	InstallStackID string `json:"install_stack_id,omitzero" temporaljson:"install_stack_id,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	Status CompositeStatus `json:"composite_status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	Runs []InstallStackVersionRun `json:"runs,omitzero" temporaljson:"runs,omitzero,omitempty"`

	Contents     []byte `json:"contents,omitzero" gorm:"type:jsonb" swaggertype:"string" temporaljson:"contents,omitzero,omitempty"`
	Checksum     string `json:"checksum,omitzero" temporaljson:"checksum,omitzero,omitempty"`
	TemplateURL  string `json:"template_url,omitzero" temporaljson:"template_url,omitzero,omitempty"`
	PhoneHomeID  string `json:"phone_home_id,omitzero" temporaljson:"phone_home_id,omitzero,omitempty"`
	PhoneHomeURL string `json:"phone_home_url,omitzero" temporaljson:"phone_home_url,omitzero,omitempty"`

	// aws configuration parameters
	AWSBucketName string `json:"aws_bucket_name,omitzero" temporaljson:"aws_bucket_name,omitzero,omitempty"`
	AWSBucketKey  string `json:"aws_bucket_key,omitzero" temporaljson:"aws_bucket_key,omitzero,omitempty"`
	QuickLinkURL  string `json:"quick_link_url,omitzero" temporaljson:"quick_link_url,omitzero,omitempty"`
}

func (a *InstallStackVersion) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallStackVersion{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
