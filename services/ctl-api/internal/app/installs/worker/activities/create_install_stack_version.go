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
		PhoneHomeID:    phoneHomeID,
		PhoneHomeURL: fmt.Sprintf(
			"%s/v1/installs/%s/phone-home/%s",
			a.cfg.PublicAPIURL,
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
		if a.cfg.AWSCloudFormationStackTemplateBaseURL != "" {
			templateURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(a.cfg.AWSCloudFormationStackTemplateBaseURL, "/"), bucketKey)
			obj.AWSBucketName = a.cfg.AWSCloudFormationStackTemplateBucket
			obj.TemplateURL = templateURL
			// Quick-launch URL embeds the region; only build it when the
			// install carries one. When region is empty, the dashboard surfaces
			// the template URL + AWS-CLI snippet instead and the user picks a
			// region in the AWS console.
			if req.Region != "" {
				obj.QuickLinkURL = fmt.Sprintf(
					"https://%s.console.aws.amazon.com/cloudformation/home?region=%s#/stacks/quickcreate?templateUrl=%s&stackName=%s",
					req.Region, req.Region, templateURL, req.StackName,
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
