package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
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

	// RunnerAPIURL is the externally-reachable runner-API host the stack-manager
	// SDK should POST to. Populated transiently on read via AfterFind from
	// Install.RunnerGroup.Settings.RunnerAPIURL — the runner API is the
	// surface vendors expose, so the customer's workstation can hit it even
	// when ctl-api itself is private.
	RunnerAPIURL string `json:"runner_api_url,omitzero" gorm:"-" temporaljson:"-"`

	// aws configuration parameters
	AWSBucketName string `json:"aws_bucket_name,omitzero" temporaljson:"aws_bucket_name,omitzero,omitempty"`
	AWSBucketKey  string `json:"aws_bucket_key,omitzero" temporaljson:"aws_bucket_key,omitzero,omitempty"`
	QuickLinkURL  string `json:"quick_link_url,omitzero" temporaljson:"quick_link_url,omitzero,omitempty"`

	// On AWS, the install workflow renders BOTH a CloudFormation template and
	// a Terraform tfvars envelope. The CFN artifact lives in Contents/Checksum
	// (and is uploaded to S3 with TemplateURL/QuickLinkURL); the Terraform
	// artifact lives below. The dashboard shows both during the await step.
	TerraformContents []byte `json:"terraform_contents,omitzero" gorm:"type:jsonb" swaggertype:"string" temporaljson:"terraform_contents,omitzero,omitempty"`
	TerraformChecksum string `json:"terraform_checksum,omitzero" temporaljson:"terraform_checksum,omitzero,omitempty"`
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

// AfterFind hydrates the transient RunnerAPIURL field by joining through
// the install's runner group. Failure is non-fatal — the dashboard falls
// back to a placeholder host in the rendered CLI command.
func (a *InstallStackVersion) AfterFind(tx *gorm.DB) error {
	if a.InstallID == "" {
		return nil
	}
	var url string
	// installs ←(polymorphic owner)→ runner_groups ←→ runner_group_settings
	if err := tx.Session(&gorm.Session{NewDB: true}).
		Table("installs").
		Select("runner_group_settings.runner_api_url").
		Joins("JOIN runner_groups ON runner_groups.owner_id = installs.id AND runner_groups.owner_type = 'installs'").
		Joins("JOIN runner_group_settings ON runner_group_settings.runner_group_id = runner_groups.id").
		Where("installs.id = ?", a.InstallID).
		Limit(1).
		Scan(&url).Error; err == nil {
		a.RunnerAPIURL = url
	}
	return nil
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
