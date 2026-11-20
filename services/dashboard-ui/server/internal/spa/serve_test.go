package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

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

func TestPortFileIsNotServed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dist := t.TempDir()
	if err := os.WriteFile(filepath.Join(dist, ".port"), []byte("4010"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dist, "ping.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte("<head></head>"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := NewHandler(&internal.Config{
		DistDir:           dist,
		AppUrl:            "https://example.com",
		ServiceDeployment: "dashboard-prod",
	}, zap.NewNop())

	e := gin.New()
	if err := h.RegisterRoutes(e); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/.port", "/nested/.port"} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: got status %d, want 404", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "4010") {
			t.Errorf("%s: response included the port file", path)
		}
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping.txt", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Errorf("static file: got %d %q, want 200 ok", rec.Code, rec.Body.String())
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
