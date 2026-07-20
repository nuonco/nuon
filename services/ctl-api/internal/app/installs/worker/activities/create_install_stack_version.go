package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallStackVersionRequest struct {
	InstallID      string `validate:"required"`
	InstallStackID string `validate:"required"`
	AppConfigID    string `validate:"required"`
	Region         string `json:"region"`
	StackName      string `json:"stack_name"`
	Platform       string `json:"platform"`
	PublicAPIURL   string `json:"public_api_url"`
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// @temporal-gen-v2 activity
func (a *Activities) CreateInstallStackVersion(ctx context.Context, req *CreateInstallStackVersionRequest) (*app.InstallStackVersion, error) {
	phoneHomeID := domains.NewAWSAccountID()
	id := domains.NewInstallStackID()

	obj := app.InstallStackVersion{
		ID:             id,
		AppConfigID:    req.AppConfigID,
		InstallID:      req.InstallID,
		InstallStackID: req.InstallStackID,
		StackName:      req.StackName,
		PhoneHomeID:    phoneHomeID,
		PhoneHomeURL: fmt.Sprintf(
			"%s/v1/installs/%s/phone-home/%s",
			firstNonEmpty(req.PublicAPIURL, a.cfg.PublicAPIURL),
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
		// the template bucket is configured; otherwise the install is
		// Terraform-only and no template is uploaded.
		if a.cfg.AWSCloudFormationStackTemplateBaseURL != "" && a.cfg.AWSCloudFormationStackTemplateBucket != "" {
			templateURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(a.cfg.AWSCloudFormationStackTemplateBaseURL, "/"), bucketKey)
			obj.AWSBucketName = a.cfg.AWSCloudFormationStackTemplateBucket
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

	if res := a.db.WithContext(ctx).Create(&obj); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create cloudformation stack version")
	}

	// create service account for install stack updates
	_, err := a.accountsHelpers.CreateServiceAccount(ctx, obj.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create install stack service account")
	}

	return &obj, nil
}
