package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
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
	StackName    string `json:"stack_name,omitzero" temporaljson:"stack_name,omitzero,omitempty"`
	PhoneHomeID  string `json:"phone_home_id,omitzero" temporaljson:"phone_home_id,omitzero,omitempty"`
	PhoneHomeURL string `json:"phone_home_url,omitzero" temporaljson:"phone_home_url,omitzero,omitempty"`

	// PhoneHomeTokenID is the tokens row minted for this stack version's phone home,
	// issued to the version's own service account. Never serialized: it identifies a
	// live credential.
	PhoneHomeTokenID string `json:"-" gorm:"index" temporaljson:"-"`

	// PhoneHomeTokenRevokedAt tombstones a token that was deliberately killed while
	// its stack version is still live — the revoke-on-successor case. Status cannot
	// express this, because a superseded version is left Active until its successor
	// has actually phoned home. Without the tombstone, an empty PhoneHomeTokenID is
	// indistinguishable from "never minted" and the reconciler resurrects the
	// credential on its next run.
	PhoneHomeTokenRevokedAt *time.Time `json:"-" temporaljson:"-"`

	// aws configuration parameters
	AWSBucketName string `json:"aws_bucket_name,omitzero" temporaljson:"aws_bucket_name,omitzero,omitempty"`
	AWSBucketKey  string `json:"aws_bucket_key,omitzero" temporaljson:"aws_bucket_key,omitzero,omitempty"`

	// QuickLinkURL opens the cloud console pre-loaded with this version's stack:
	// CloudFormation quick-create on AWS, the portal's Custom Deployment blade on
	// Azure. Empty on GCP, and on any install whose template bucket is
	// unconfigured.
	QuickLinkURL string `json:"quick_link_url,omitzero" temporaljson:"quick_link_url,omitzero,omitempty"`

	// QuickLinkBucketKey is the Azure-only second S3 object behind QuickLinkURL: a
	// wrapper template whose sole resource is a deployment stack pointing at
	// AWSBucketKey's template. The portal cannot create a deployment stack
	// directly, and the template cannot be inlined into the wrapper — see
	// arm.QuickLinkWrapper. Empty on AWS, where the quick link addresses the
	// template itself.
	QuickLinkBucketKey string `json:"quick_link_bucket_key,omitzero" temporaljson:"quick_link_bucket_key,omitzero,omitempty"`

	// QuickLinkUIDefBucketKey is the Azure-only createUiDefinition accompanying the
	// wrapper. It constrains the portal's Basics step to the install's resource
	// group and location, so that a reprovision updates the install's stack instead
	// of silently creating a second one alongside it.
	QuickLinkUIDefBucketKey string `json:"quick_link_ui_def_bucket_key,omitzero" temporaljson:"quick_link_ui_def_bucket_key,omitzero,omitempty"`

	// On AWS, the install workflow renders BOTH a CloudFormation template and
	// a Terraform tfvars envelope. The CFN artifact lives in Contents/Checksum
	// (and is uploaded to S3 with TemplateURL/QuickLinkURL); the Terraform
	// artifact lives below. The dashboard shows both during the await step.
	TerraformContents []byte `json:"terraform_contents,omitzero" gorm:"type:jsonb" swaggertype:"string" temporaljson:"terraform_contents,omitzero,omitempty"`
	TerraformChecksum string `json:"terraform_checksum,omitzero" temporaljson:"terraform_checksum,omitzero,omitempty"`

	CallbackRef callback.Ref `json:"callback_ref,omitzero" gorm:"type:jsonb" temporaljson:"callback_ref,omitzero,omitempty"`
}

// PhoneHomeTokenEligibleStatuses are the statuses for which a stack version should
// hold a live phone-home token. Retired versions (cancelled, expired, outdated) want
// no credential at all — the handler rejects an expired version outright, and an
// outdated one has been superseded by a version that already phoned home.
//
// Exported as a slice because the reconciler needs the same rule in SQL, where
// status is jsonb: Where("(status->>'status') IN ?", PhoneHomeTokenEligibleStatuses).
var PhoneHomeTokenEligibleStatuses = []Status{
	InstallStackVersionStatusGenerating,
	InstallStackVersionStatusPendingUser,
	InstallStackVersionStatusProvisioning,
	InstallStackVersionStatusActive,
}

// PhoneHomeTokenEligible reports whether this version should hold a live phone-home
// token. A tombstoned token is never reissued: that is what distinguishes a
// deliberate revocation from a version that has simply never been minted for.
func (a *InstallStackVersion) PhoneHomeTokenEligible() bool {
	if a.PhoneHomeTokenRevokedAt != nil {
		return false
	}

	for _, status := range PhoneHomeTokenEligibleStatuses {
		if a.Status.Status == status {
			return true
		}
	}

	return false
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
