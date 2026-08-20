package activities

import (
	"context"
	"fmt"
	"net/url"
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

// azurePortalCustomDeployBaseURL is the portal's Custom Deployment blade, in the
// form that takes a createUiDefinition alongside the template. The plain
// `#create/Microsoft.Template/uri/<template>` form renders an uncontrolled
// Basics step, where a customer picking a resource group other than
// <install-id>-rg silently creates a second deployment stack rather than
// updating the install's. The UI definition constrains that step.
const azurePortalCustomDeployBaseURL = "https://portal.azure.com/#blade/Microsoft_Azure_CreateUIDef/CustomDeploymentBlade/uri/"

// escapeDataString URL-encodes a value the way Azure documents for portal deep
// links ([uri]::EscapeDataString). url.PathEscape leaves ':' unescaped and
// url.QueryEscape renders ' ' as '+'; neither matches on its own.
func escapeDataString(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

type templateLocations struct {
	templateURL        string
	quickLinkURL       string
	quickLinkBucketKey string
	quickLinkUIDefKey  string
}

// stackTemplateLocations derives the S3 URL of a stack version's template and the
// console link that deploys it. The two platforms differ in what the link points
// at: CloudFormation's quick-create takes the template itself, while the Azure
// portal takes a wrapper template stored under its own key, so that the portal
// creates a deployment stack rather than a bare deployment.
func stackTemplateLocations(configuredBaseURL, bucketKey string, req *CreateInstallStackVersionRequest) templateLocations {
	baseURL := strings.TrimSuffix(configuredBaseURL, "/")
	loc := templateLocations{templateURL: fmt.Sprintf("%s/%s", baseURL, bucketKey)}

	if req.Platform == string(app.AppRunnerTypeAzure) {
		keyStem := strings.TrimSuffix(bucketKey, ".json")
		loc.quickLinkBucketKey = keyStem + "-quicklink.json"
		loc.quickLinkUIDefKey = keyStem + "-uidef.json"

		wrapperURL := fmt.Sprintf("%s/%s", baseURL, loc.quickLinkBucketKey)
		uiDefURL := fmt.Sprintf("%s/%s", baseURL, loc.quickLinkUIDefKey)
		loc.quickLinkURL = fmt.Sprintf("%s%s/createUIDefinitionUri/%s",
			azurePortalCustomDeployBaseURL,
			escapeDataString(wrapperURL),
			escapeDataString(uiDefURL),
		)
		return loc
	}

	// When the install pins a region we embed it in the quick-launch URL;
	// otherwise emit a region-less variant that opens in whatever region the user
	// currently has selected in the AWS console.
	if req.Region != "" {
		loc.quickLinkURL = fmt.Sprintf(
			"https://%s.console.aws.amazon.com/cloudformation/home?region=%s#/stacks/quickcreate?templateUrl=%s&stackName=%s",
			req.Region, req.Region, loc.templateURL, req.StackName,
		)
		return loc
	}

	loc.quickLinkURL = fmt.Sprintf(
		"https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?templateUrl=%s&stackName=%s",
		loc.templateURL, req.StackName,
	)
	return loc
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
		obj.AWSBucketKey = fmt.Sprintf("templates/%s/%s.json", req.InstallID, id)

		// Only generate S3-based template URL and quick link when the template
		// bucket is configured; otherwise the install is Terraform-only and no
		// template is uploaded.
		if a.cfg.AWSCloudFormationStackTemplateBaseURL != "" && a.cfg.AWSCloudFormationStackTemplateBucket != "" {
			obj.AWSBucketName = a.cfg.AWSCloudFormationStackTemplateBucket
			loc := stackTemplateLocations(a.cfg.AWSCloudFormationStackTemplateBaseURL, obj.AWSBucketKey, req)
			obj.TemplateURL = loc.templateURL
			obj.QuickLinkURL = loc.quickLinkURL
			obj.QuickLinkBucketKey = loc.quickLinkBucketKey
			obj.QuickLinkUIDefBucketKey = loc.quickLinkUIDefKey
		}
	}

	if res := a.db.WithContext(ctx).Create(&obj); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create cloudformation stack version")
	}

	// create service account for install stack updates
	_, err := a.accountsHelpers.CreateServiceAccount(ctx, obj.ID, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create install stack service account")
	}

	return &obj, nil
}
