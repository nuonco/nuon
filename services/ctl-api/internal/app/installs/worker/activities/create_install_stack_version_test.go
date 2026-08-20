package activities

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	testBaseURL   = "https://templates.example.com"
	testBucketKey = "templates/inl-test/ist-test.json"
)

func TestStackTemplateLocations_AzureSubscriptionScope(t *testing.T) {
	loc := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
		Platform:        string(app.AppRunnerTypeAzure),
		Region:          "centralus",
		StackName:       "test-stack",
		DeploymentScope: string(app.StackDeploymentScopeSubscription),
	})

	if got, want := loc.templateURL, testBaseURL+"/"+testBucketKey; got != want {
		t.Errorf("templateURL = %q, want %q", got, want)
	}

	want := azurePortalCustomDeployBaseURL + escapeDataString(loc.templateURL)
	if loc.quickLinkURL != want {
		t.Errorf("quickLinkURL = %q, want %q", loc.quickLinkURL, want)
	}

	// Pinned to the literal rather than the constant. Expressing this test only in
	// terms of azurePortalCustomDeployBaseURL is what let the undocumented
	// Microsoft_Azure_CreateUIDef/CustomDeploymentBlade route ship: that blade
	// accepts the identical shape and silently renders nothing, so nothing failed
	// until someone opened the link.
	if !strings.HasPrefix(loc.quickLinkURL, "https://portal.azure.com/#create/Microsoft.Template/uri/") {
		t.Errorf("quick link does not use the documented Deploy-to-Azure route: %q", loc.quickLinkURL)
	}

	// The template URL sits in a path segment, so its separators must be escaped.
	// An unescaped one silently truncates the segment and the deployment fails to
	// load.
	segment := strings.TrimPrefix(loc.quickLinkURL, azurePortalCustomDeployBaseURL)
	if strings.Contains(segment, "/") {
		t.Errorf("quick link has unescaped separators: %q", loc.quickLinkURL)
	}
}

// The portal cannot deploy a resource-group-scoped root template: it has no way
// to create the group first. Rather than hand the customer a link that fails,
// emit none — the dashboard hides the button when the URL is empty.
func TestStackTemplateLocations_AzureResourceGroupScopeHasNoQuickLink(t *testing.T) {
	for _, scope := range []string{"", string(app.StackDeploymentScopeResourceGroup)} {
		loc := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
			Platform:        string(app.AppRunnerTypeAzure),
			StackName:       "test-stack",
			DeploymentScope: scope,
		})

		if got, want := loc.templateURL, testBaseURL+"/"+testBucketKey; got != want {
			t.Errorf("scope %q: templateURL = %q, want %q", scope, got, want)
		}
		if loc.quickLinkURL != "" {
			t.Errorf("scope %q: expected no quick link, got %q", scope, loc.quickLinkURL)
		}
	}
}

// Azure documents portal deep-link encoding as [uri]::EscapeDataString, which
// escapes ':' — url.PathEscape does not, and url.QueryEscape turns ' ' into '+'.
func TestEscapeDataString(t *testing.T) {
	got := escapeDataString("https://templates.example.com/a b/c.json")
	want := "https%3A%2F%2Ftemplates.example.com%2Fa%20b%2Fc.json"
	if got != want {
		t.Errorf("escapeDataString = %q, want %q", got, want)
	}
}

// Azure ignores Region for the quick link: the portal blade prompts for
// subscription, resource group, and location itself.
func TestStackTemplateLocations_AzureIgnoresRegion(t *testing.T) {
	withRegion := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
		Platform:        string(app.AppRunnerTypeAzure),
		Region:          "centralus",
		DeploymentScope: string(app.StackDeploymentScopeSubscription),
	})
	withoutRegion := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
		Platform:        string(app.AppRunnerTypeAzure),
		DeploymentScope: string(app.StackDeploymentScopeSubscription),
	})

	if withRegion.quickLinkURL != withoutRegion.quickLinkURL {
		t.Errorf("region changed the azure quick link: %q vs %q", withRegion.quickLinkURL, withoutRegion.quickLinkURL)
	}
}

func TestStackTemplateLocations_AWSWithRegion(t *testing.T) {
	loc := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
		Platform:  string(app.AppRunnerTypeAWS),
		Region:    "us-west-2",
		StackName: "test-stack",
	})

	if !strings.HasPrefix(loc.quickLinkURL, "https://us-west-2.console.aws.amazon.com/cloudformation/home?region=us-west-2") {
		t.Errorf("unexpected aws quick link: %q", loc.quickLinkURL)
	}
	if !strings.Contains(loc.quickLinkURL, "stackName=test-stack") {
		t.Errorf("aws quick link missing stack name: %q", loc.quickLinkURL)
	}
}

func TestStackTemplateLocations_AWSWithoutRegion(t *testing.T) {
	loc := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
		Platform:  string(app.AppRunnerTypeAWS),
		StackName: "test-stack",
	})

	if !strings.HasPrefix(loc.quickLinkURL, "https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate") {
		t.Errorf("unexpected region-less aws quick link: %q", loc.quickLinkURL)
	}
	if strings.Contains(loc.quickLinkURL, "region=") {
		t.Errorf("region-less aws quick link should not pin a region: %q", loc.quickLinkURL)
	}
}

func TestStackTemplateLocations_TrimsBaseURLSlash(t *testing.T) {
	loc := stackTemplateLocations(testBaseURL+"/", testBucketKey, &CreateInstallStackVersionRequest{
		Platform:        string(app.AppRunnerTypeAzure),
		DeploymentScope: string(app.StackDeploymentScopeSubscription),
	})

	if got, want := loc.templateURL, testBaseURL+"/"+testBucketKey; got != want {
		t.Errorf("templateURL = %q, want %q", got, want)
	}
	// The doubled separator survives escaping as %2F%2F, so it would reach the
	// portal rather than being collapsed by an HTTP client.
	if strings.Contains(loc.quickLinkURL, escapeDataString("//"+testBucketKey)) {
		t.Errorf("quick link has a doubled separator: %q", loc.quickLinkURL)
	}
}
