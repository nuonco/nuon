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

const azurePortalCustomDeployBaseURL = "https://portal.azure.com/#create/Microsoft.Template/uri/"

func escapeDataString(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func azureUIDefBucketKey(templateKey string) string {
	if !strings.HasSuffix(templateKey, ".json") {
		return ""
	}
	return strings.TrimSuffix(templateKey, ".json") + "-ui.json"
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

func stackTemplateLocations(configuredBaseURL, bucketKey string, req *CreateInstallStackVersionRequest) templateLocations {
	baseURL := strings.TrimSuffix(configuredBaseURL, "/")
	loc := templateLocations{templateURL: fmt.Sprintf("%s/%s", baseURL, bucketKey)}

	if req.Platform == string(app.AppRunnerTypeAzure) {
		if req.DeploymentScope != string(app.StackDeploymentScopeSubscription) {
			return loc
		}

		loc.quickLinkURL = azurePortalCustomDeployBaseURL + escapeDataString(loc.templateURL)
		if uiDefKey := azureUIDefBucketKey(bucketKey); uiDefKey != "" {
			uiDefURL := fmt.Sprintf("%s/%s", baseURL, uiDefKey)
			loc.quickLinkURL += "/createUIDefinitionUri/" + escapeDataString(uiDefURL)
		}
		return loc
	}

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

	if req.Platform != "gcp" {
		obj.AWSBucketKey = fmt.Sprintf("templates/%s/%s.json", req.InstallID, id)

		if a.cfg.AWSCloudFormationStackTemplateBaseURL != "" && a.cfg.AWSCloudFormationStackTemplateBucket != "" {
			obj.AWSBucketName = a.cfg.AWSCloudFormationStackTemplateBucket
			loc := stackTemplateLocations(a.cfg.AWSCloudFormationStackTemplateBaseURL, obj.AWSBucketKey, req)
			obj.TemplateURL = loc.templateURL
			obj.QuickLinkURL = loc.quickLinkURL
			if req.Platform == string(app.AppRunnerTypeAzure) && req.DeploymentScope == string(app.StackDeploymentScopeSubscription) {
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

	_, err := a.accountsHelpers.CreateServiceAccount(ctx, obj.ID, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create install stack service account")
	}

	return &obj, nil
}
