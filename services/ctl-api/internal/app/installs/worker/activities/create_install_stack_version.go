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
	// AWS/Azure still use S3-hosted templates with quick links.
	if req.Platform != "gcp" {
		bucketKey := fmt.Sprintf("templates/%s/%s.json", req.InstallID, id)
		templateURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(a.cfg.AWSCloudFormationStackTemplateBaseURL, "/"), bucketKey)
		quickLinkURL := fmt.Sprintf("https://%s.console.aws.amazon.com/cloudformation/home?region=%s#/stacks/quickcreate?templateUrl=%s&stackName=%s",
			req.Region, req.Region, templateURL, req.StackName,
		)

		obj.AWSBucketName = a.cfg.AWSCloudFormationStackTemplateBucket
		obj.AWSBucketKey = bucketKey
		obj.TemplateURL = templateURL
		obj.QuickLinkURL = quickLinkURL
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
