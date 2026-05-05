package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// CreateInstallStackVersionRequest is the input for CreateInstallStackVersion.
// Mirrors the activity request shape so the activity can call this helper.
type CreateInstallStackVersionRequest struct {
	InstallID      string
	InstallStackID string
	AppConfigID    string
	Region         string
	StackName      string
	Platform       string
}

// CreateInstallStackVersion inserts an InstallStackVersion row and creates the
// associated service account used to update the stack later. Shared by the
// generate-stack-version Temporal activity and by install-creation paths that
// need to produce a stack version inline (e.g. native AWS provisioner toggle).
func (h *Helpers) CreateInstallStackVersion(ctx context.Context, req *CreateInstallStackVersionRequest) (*app.InstallStackVersion, error) {
	phoneHomeID := domains.NewAWSAccountID()
	id := domains.NewInstallStackID()

	obj := app.InstallStackVersion{
		ID:             id,
		AppConfigID:    req.AppConfigID,
		InstallID:      req.InstallID,
		InstallStackID: req.InstallStackID,
		PhoneHomeID:    phoneHomeID,
		PhoneHomeURL: fmt.Sprintf(
			"%s/v1/installs/%s/phone-home/%s",
			h.cfg.PublicAPIURL,
			req.InstallID,
			phoneHomeID,
		),
		Status: app.NewCompositeStatus(ctx, app.InstallStackVersionStatusGenerating),
	}

	// GCP uses static Terraform modules with tfvars, no S3 upload needed.
	// AWS/Azure render both a CloudFormation template (S3-hosted, with a
	// quick link) and — for AWS — a Terraform tfvars envelope stored on the
	// row. The user picks one to apply during the await step.
	if req.Platform != "gcp" {
		bucketKey := fmt.Sprintf("templates/%s/%s.json", req.InstallID, id)
		obj.AWSBucketKey = bucketKey

		// Only generate S3-based template URL and CloudFormation quick link when
		// running on AWS BYOC (S3 base URL is configured). On GCP BYOC, the
		// template is uploaded to GCS and the URL is set after upload.
		if h.cfg.AWSCloudFormationStackTemplateBaseURL != "" {
			templateURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(h.cfg.AWSCloudFormationStackTemplateBaseURL, "/"), bucketKey)
			obj.AWSBucketName = h.cfg.AWSCloudFormationStackTemplateBucket
			obj.TemplateURL = templateURL
			// When the install pins a region we embed it in the quick-launch
			// URL; otherwise emit a region-less variant that opens in whatever
			// region the user currently has selected in the AWS console.
			if req.Region != "" {
				obj.QuickLinkURL = fmt.Sprintf(
					"https://%s.console.aws.amazon.com/cloudformation/home?region=%s#/stacks/quickcreate?templateUrl=%s&stackName=%s",
					req.Region, req.Region, templateURL, req.StackName,
				)
			} else {
				obj.QuickLinkURL = fmt.Sprintf(
					"https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?templateUrl=%s&stackName=%s",
					templateURL, req.StackName,
				)
			}
		}
	}

	if res := h.db.WithContext(ctx).Create(&obj); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install stack version")
	}

	if _, err := h.accountsClient.CreateServiceAccount(ctx, obj.ID); err != nil {
		return nil, errors.Wrap(err, "unable to create install stack service account")
	}

	return &obj, nil
}

// CreateInstallStackVersionForInstall is the entry point used by install
// creation when the OrgFeatureNativeAWSProvisioner toggle is on. It looks up
// the install's stack and app config, derives region + platform, and creates
// the stack version inline (no workflow involved).
func (h *Helpers) CreateInstallStackVersionForInstall(ctx context.Context, install *app.Install) (*app.InstallStackVersion, error) {
	stack, err := h.getInstallStack(ctx, install.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get install stack")
	}
	cfg, err := h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
	if err != nil {
		return nil, errors.Wrap(err, "get app config")
	}
	if cfg == nil {
		return nil, errors.New("no app config found for install")
	}

	region := ""
	switch {
	case install.AWSAccount != nil:
		region = install.AWSAccount.Region
	case install.AzureAccount != nil:
		region = install.AzureAccount.Location
	case install.GCPAccount != nil:
		region = install.GCPAccount.Region
	}

	return h.CreateInstallStackVersion(ctx, &CreateInstallStackVersionRequest{
		InstallID:      install.ID,
		InstallStackID: stack.ID,
		AppConfigID:    cfg.ID,
		StackName:      cfg.StackConfig.Name,
		Region:         region,
		Platform:       string(cfg.RunnerConfig.Type),
	})
}
