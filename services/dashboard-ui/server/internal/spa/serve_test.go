package spa

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const shellHTML = `<head>
    <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
    <link rel="stylesheet" href="/assets/styles.css" data-shell="default" />
    <link rel="stylesheet" href="/assets/app.css" data-shell="default" />
    <link rel="stylesheet" href="/assets/lite.css" data-shell="lite" />
  </head>`

func TestSelectShellLinksDefault(t *testing.T) {
	got := string(selectShellLinks([]byte(shellHTML), shellDefault))

	for _, want := range []string{"/assets/styles.css", "/assets/app.css", "/favicon.svg"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q to be kept, got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "/assets/lite.css") {
		t.Errorf("expected lite stylesheet to be stripped, got:\n%s", got)
	}
}

func TestSelectShellLinksLite(t *testing.T) {
	got := string(selectShellLinks([]byte(shellHTML), shellLite))

	for _, want := range []string{"/assets/lite.css", "/favicon.svg"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q to be kept, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"/assets/styles.css", "/assets/app.css"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("expected %q to be stripped, got:\n%s", unwanted, got)
		}
	}
}

func TestSelectShellLinksHashedFilenames(t *testing.T) {
	hashed := `<link rel="stylesheet" href="/assets/lite-a1b2c3d4.css" data-shell="lite" />
<link rel="stylesheet" href="/assets/styles-e5f6a7b8.css" data-shell="default" />`

	got := string(selectShellLinks([]byte(hashed), shellLite))

	if !strings.Contains(got, "/assets/lite-a1b2c3d4.css") {
		t.Errorf("expected hashed lite stylesheet to be kept, got:\n%s", got)
	}
	if strings.Contains(got, "/assets/styles-e5f6a7b8.css") {
		t.Errorf("expected hashed default stylesheet to be stripped, got:\n%s", got)
	}
}

func TestBuildClientConfigIncludesDashboardDefaults(t *testing.T) {
	got := buildClientConfig(&internal.Config{
		StatusBarAutoEnabled:   true,
		InstallsTabAutoEnabled: true,
	})

	if !got.StatusBarAutoEnabled {
		t.Error("expected status bar auto-enabled setting in client config")
	}
	if !got.InstallsTabAutoEnabled {
		t.Error("expected installs tab auto-enabled setting in client config")
	}
}

func TestBuildClientConfigPostHog(t *testing.T) {
	withKey := buildClientConfig(&internal.Config{
		PostHogKey:  "phc_test",
		PostHogHost: "https://us.i.posthog.com",
	})
	if withKey.PostHogKey != "phc_test" || withKey.PostHogHost != "https://us.i.posthog.com" {
		t.Errorf("expected posthog key and host in client config when configured, got %+v", withKey)
	}

	byocWithKey := buildClientConfig(&internal.Config{
		IsBYOC:     true,
		PostHogKey: "phc_test",
	})
	if byocWithKey.PostHogKey != "phc_test" {
		t.Errorf("expected posthog enabled on BYOC when key is configured, got %+v", byocWithKey)
	}

	for _, cfg := range []*internal.Config{
		{},
		{IsBYOC: true},
	} {
		got := buildClientConfig(cfg)
		if got.PostHogKey != "" || got.PostHogHost != "" {
			t.Errorf("expected posthog omitted from client config when no key is configured, got %+v", got)
		}
	}
}
