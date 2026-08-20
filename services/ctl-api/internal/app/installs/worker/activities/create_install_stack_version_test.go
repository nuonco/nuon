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

func TestStackTemplateLocations_Azure(t *testing.T) {
	loc := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
		Platform:  string(app.AppRunnerTypeAzure),
		Region:    "centralus",
		StackName: "test-stack",
	})

	if got, want := loc.templateURL, testBaseURL+"/"+testBucketKey; got != want {
		t.Errorf("templateURL = %q, want %q", got, want)
	}
	if got, want := loc.quickLinkBucketKey, "templates/inl-test/ist-test-quicklink.json"; got != want {
		t.Errorf("quickLinkBucketKey = %q, want %q", got, want)
	}
	if got, want := loc.quickLinkUIDefKey, "templates/inl-test/ist-test-uidef.json"; got != want {
		t.Errorf("quickLinkUIDefKey = %q, want %q", got, want)
	}

	wrapperURL := testBaseURL + "/" + loc.quickLinkBucketKey
	uiDefURL := testBaseURL + "/" + loc.quickLinkUIDefKey
	want := azurePortalCustomDeployBaseURL + escapeDataString(wrapperURL) + "/createUIDefinitionUri/" + escapeDataString(uiDefURL)
	if loc.quickLinkURL != want {
		t.Errorf("quickLinkURL = %q, want %q", loc.quickLinkURL, want)
	}

	// Pinned to the literal rather than the constant. Expressing this test only in
	// terms of azurePortalCustomDeployBaseURL is what let the undocumented
	// Microsoft_Azure_CreateUIDef/CustomDeploymentBlade route ship: that blade
	// accepts the identical two segments and silently renders none of the UI
	// definition's fields, so nothing failed until someone opened the link.
	if !strings.HasPrefix(loc.quickLinkURL, "https://portal.azure.com/#create/Microsoft.Template/uri/") {
		t.Errorf("quick link does not use the documented Deploy-to-Azure route: %q", loc.quickLinkURL)
	}

	// Both URLs sit in path segments, so their separators must be escaped. An
	// unescaped one silently truncates the segment and the deployment fails to
	// load. The only slashes left after the prefix are the two bounding
	// /createUIDefinitionUri/.
	segments := strings.TrimPrefix(loc.quickLinkURL, azurePortalCustomDeployBaseURL)
	if strings.Count(segments, "/") != 2 {
		t.Errorf("quick link has unescaped separators: %q", loc.quickLinkURL)
	}

	// The quick link must address the wrapper, never the stack template — pointing
	// the portal at the template produces a plain deployment with no deny settings.
	if strings.Contains(loc.quickLinkURL, escapeDataString(loc.templateURL)) {
		t.Errorf("quick link points at the stack template rather than the wrapper: %q", loc.quickLinkURL)
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
		Platform: string(app.AppRunnerTypeAzure),
		Region:   "centralus",
	})
	withoutRegion := stackTemplateLocations(testBaseURL, testBucketKey, &CreateInstallStackVersionRequest{
		Platform: string(app.AppRunnerTypeAzure),
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

	if loc.quickLinkBucketKey != "" {
		t.Errorf("aws must not allocate a wrapper key, got %q", loc.quickLinkBucketKey)
	}
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
		Platform: string(app.AppRunnerTypeAzure),
	})

	if got, want := loc.templateURL, testBaseURL+"/"+testBucketKey; got != want {
		t.Errorf("templateURL = %q, want %q", got, want)
	}
	if strings.Contains(loc.quickLinkBucketKey, "//") {
		t.Errorf("wrapper key has a doubled separator: %q", loc.quickLinkBucketKey)
	}
}
