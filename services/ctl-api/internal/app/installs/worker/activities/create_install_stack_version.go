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
	InstallID             string `validate:"required"`
	InstallStackID        string `validate:"required"`
	AppConfigID           string `validate:"required"`
	Region                string `json:"region"`
	StackName             string `json:"stack_name"`
	Platform              string `json:"platform"`
	PublicAPIURL          string `json:"public_api_url"`
	DeploymentScope       string `json:"deployment_scope"`
	HasCustomNestedStacks bool   `json:"has_custom_nested_stacks"`
}

// azurePortalCustomDeployBaseURL is the documented Deploy-to-Azure form. The
// encoded template URL is appended as the final path segment.
//
// Not Microsoft_Azure_CreateUIDef/CustomDeploymentBlade, which is undocumented and
// accepts the same shape while rendering none of a createUiDefinition's elements —
// it applies config.basics and drops basics/steps, so the form collapses to
// subscription and region with no way to tell that anything was ignored.
const azurePortalCustomDeployBaseURL = "https://portal.azure.com/#create/Microsoft.Template/uri/"

// escapeDataString URL-encodes a value the way Azure documents for portal deep
// links ([uri]::EscapeDataString). url.PathEscape leaves ':' unescaped and
// url.QueryEscape renders ' ' as '+'; neither matches on its own.
func escapeDataString(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func azureUIDefBucketKey(templateKey string) string {
	return azureSiblingBucketKey(templateKey, "-ui.json")
}

func azureQuickLinkBucketKey(templateKey string) string {
	return azureSiblingBucketKey(templateKey, "-quicklink.json")
}

func azureSiblingBucketKey(templateKey, suffix string) string {
	if !strings.HasSuffix(templateKey, ".json") {
		return ""
	}
	return strings.TrimSuffix(templateKey, ".json") + suffix
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
	templateURL  string
	quickLinkURL string
}

// stackTemplateLocations derives the S3 URL of a stack version's template and the
// console link that deploys it. CloudFormation's quick-create always has one and
// points at the template itself. The Azure portal can only deploy a root template
// that declares its own resource group, so only subscription scope gets a link,
// and it points at the deployment-stack wrapper so a role unticked on a later
// deploy is deleted rather than left behind.
func stackTemplateLocations(configuredBaseURL, bucketKey string, req *CreateInstallStackVersionRequest) templateLocations {
	baseURL := strings.TrimSuffix(configuredBaseURL, "/")
	loc := templateLocations{templateURL: fmt.Sprintf("%s/%s", baseURL, bucketKey)}

	if req.Platform == string(app.AppRunnerTypeAzure) {
		if req.DeploymentScope != string(app.StackDeploymentScopeSubscription) {
			return loc
		}

		wrapperURL := loc.templateURL
		if wrapperKey := azureQuickLinkBucketKey(bucketKey); wrapperKey != "" {
			wrapperURL = fmt.Sprintf("%s/%s", baseURL, wrapperKey)
		}
		loc.quickLinkURL = azurePortalCustomDeployBaseURL + escapeDataString(wrapperURL)
		if uiDefKey := azureUIDefBucketKey(bucketKey); uiDefKey != "" {
			uiDefURL := fmt.Sprintf("%s/%s", baseURL, uiDefKey)
			loc.quickLinkURL += "/createUIDefinitionUri/" + escapeDataString(uiDefURL)
		}
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
			if req.Platform == string(app.AppRunnerTypeAzure) && req.DeploymentScope == string(app.StackDeploymentScopeSubscription) {
				obj.QuickLinkBucketKey = azureQuickLinkBucketKey(obj.AWSBucketKey)
				obj.QuickLinkUIDefBucketKey = azureUIDefBucketKey(obj.AWSBucketKey)
			}

			if (req.Platform == "aws" || req.Platform == "azure") && req.HasCustomNestedStacks {
				obj.CustomStacksAWSBucketKey = fmt.Sprintf("templates/%s/%s-custom.json", req.InstallID, id)
				baseURL := strings.TrimSuffix(a.cfg.AWSCloudFormationStackTemplateBaseURL, "/")
				obj.CustomStacksTemplateURL = fmt.Sprintf("%s/%s", baseURL, obj.CustomStacksAWSBucketKey)
			}
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
