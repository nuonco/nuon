package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
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

	// Set means the template was rendered to authenticate, so an unauthenticated phone
	// home is rejected rather than skipped. Empty versions predate enforcement on their
	// cloud and are skipped, which is what lets this roll out per cloud.
	PhoneHomeIdentityName string `json:"-" temporaljson:"-"`

	// aws configuration parameters
	AWSBucketName string `json:"aws_bucket_name,omitzero" temporaljson:"aws_bucket_name,omitzero,omitempty"`
	AWSBucketKey  string `json:"aws_bucket_key,omitzero" temporaljson:"aws_bucket_key,omitzero,omitempty"`

	CustomStacksTemplateURL        string                       `json:"custom_stacks_template_url,omitzero" temporaljson:"custom_stacks_template_url,omitzero,omitempty"`
	CustomStacksAWSBucketKey       string                       `json:"custom_stacks_aws_bucket_key,omitzero" temporaljson:"custom_stacks_aws_bucket_key,omitzero,omitempty"`
	CustomStacksOutputMap          map[string]map[string]string `json:"custom_stacks_output_map,omitzero" gorm:"type:jsonb;serializer:json" swaggertype:"object" temporaljson:"custom_stacks_output_map,omitzero,omitempty"`
	CustomStacksInputParametersMap map[string]map[string]string `json:"custom_stacks_input_parameters_map,omitzero" gorm:"type:jsonb;serializer:json" swaggertype:"object" temporaljson:"custom_stacks_input_parameters_map,omitzero,omitempty"`

	// QuickLinkURL opens the cloud console pre-loaded with this version's stack:
	// CloudFormation quick-create on AWS, Deploy to Azure on Azure. Empty on GCP,
	// on any install whose template bucket is unconfigured, and on an Azure install
	// at resource group scope — the portal cannot create the resource group the
	// root template needs, so there is no link to offer.
	QuickLinkURL string `json:"quick_link_url,omitzero" temporaljson:"quick_link_url,omitzero,omitempty"`

	// QuickLinkBucketKey and QuickLinkUIDefBucketKey held the wrapper template and
	// createUiDefinition that an earlier Azure quick link pointed at, so that the
	// portal created a deployment stack rather than a plain deployment. Nothing
	// writes them now: the quick link addresses the stack template directly on both
	// platforms. Rows created while the wrapper shipped still carry their keys.
	QuickLinkBucketKey      string `json:"quick_link_bucket_key,omitzero" temporaljson:"quick_link_bucket_key,omitzero,omitempty"`
	QuickLinkUIDefBucketKey string `json:"quick_link_ui_def_bucket_key,omitzero" temporaljson:"quick_link_ui_def_bucket_key,omitzero,omitempty"`

	// On AWS, the install workflow renders BOTH a CloudFormation template and
	// a Terraform tfvars envelope. The CFN artifact lives in Contents/Checksum
	// (and is uploaded to S3 with TemplateURL/QuickLinkURL); the Terraform
	// artifact lives below. The dashboard shows both during the await step.
	TerraformContents []byte `json:"terraform_contents,omitzero" gorm:"type:jsonb" swaggertype:"string" temporaljson:"terraform_contents,omitzero,omitempty"`
	TerraformChecksum string `json:"terraform_checksum,omitzero" temporaljson:"terraform_checksum,omitzero,omitempty"`

	CallbackRef callback.Ref `json:"callback_ref,omitzero" gorm:"type:jsonb" temporaljson:"callback_ref,omitzero,omitempty"`

	// CompositeError holds a typed, structured error frozen at write time when
	// stack template generation fails due to a config or rendering problem. It
	// is nil for successful versions and for failures not attributed to template
	// rendering (e.g. transient infrastructure upload errors).
	CompositeError *compositeerrors.CompositeErrorData `json:"composite_error,omitempty" gorm:"type:jsonb" temporaljson:"composite_error,omitzero,omitempty"`
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
