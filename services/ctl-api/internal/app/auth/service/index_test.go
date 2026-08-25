package service

import (
	"html/template"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseNuonBrandedLogin(t *testing.T) {
	tests := []struct {
		name        string
		flagEnabled bool
		appURL      string
		want        bool
	}{
		{
			name:   "nuon cloud app url enables the branded page",
			appURL: "https://app.nuon.co",
			want:   true,
		},
		{
			name:   "nuon cloud app url with path enables the branded page",
			appURL: "https://app.nuon.co/orgs",
			want:   true,
		},
		{
			name:   "byoc vendor app url keeps the default page",
			appURL: "https://dashboard.vendor-example.com",
			want:   false,
		},
		{
			name:   "vendor subdomain that merely contains nuon.co keeps the default page",
			appURL: "https://app.nuon.co.vendor-example.com",
			want:   false,
		},
		{
			name:   "local dev keeps the default page",
			appURL: "http://localhost:4000",
			want:   false,
		},
		{
			name:        "explicit flag enables the branded page on any app url",
			flagEnabled: true,
			appURL:      "http://localhost:4000",
			want:        true,
		},
		{
			name:   "unparseable app url keeps the default page",
			appURL: "://not-a-url",
			want:   false,
		},
		{
			name:   "empty app url keeps the default page",
			appURL: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, useNuonBrandedLogin(tt.flagEnabled, tt.appURL))
		})
	}
}

// Nothing parses the embedded templates outside a running server, so a syntax or field error in
// index_nuon would only surface at startup. Render both index templates here to catch it in CI.
func TestIndexTemplatesRender(t *testing.T) {
	sub, err := fs.Sub(tmplFS, "templates")
	require.NoError(t, err)

	tmpl, err := template.ParseFS(sub, "*.tmpl")
	require.NoError(t, err)

	providers := []ProviderOption{
		{ID: "default-google", Name: "Google", ProviderType: "google"},
		{ID: "gh", Name: "GitHub", ProviderType: "github"},
		{ID: "sso", Name: "Single Sign-On", Hint: "idp.example.com", ProviderType: "oidc"},
	}

	for _, name := range []string{"auth/index.tmpl", "auth/index_nuon.tmpl"} {
		t.Run(name+" signed out", func(t *testing.T) {
			var b strings.Builder
			err := tmpl.ExecuteTemplate(&b, name, map[string]any{
				"IsAuthenticated": false,
				"Email":           "",
				"Providers":       providers,
				"RedirectURL":     "https%3A%2F%2Fapp.nuon.co",
				"DashboardURL":    "https://app.nuon.co",
			})
			require.NoError(t, err)
			assert.Contains(t, b.String(), "Continue with Google")
			assert.NotContains(t, b.String(), "posthog.init", "no PostHogKey must mean no analytics snippet")
		})

		t.Run(name+" signed in", func(t *testing.T) {
			var b strings.Builder
			err := tmpl.ExecuteTemplate(&b, name, map[string]any{
				"IsAuthenticated": true,
				"Email":           "user@example.com",
				"Providers":       providers,
				"RedirectURL":     "",
				"DashboardURL":    "https://app.nuon.co",
			})
			require.NoError(t, err)
			assert.Contains(t, b.String(), "user@example.com")
		})
	}

	t.Run("auth/index_nuon.tmpl posthog snippet", func(t *testing.T) {
		var b strings.Builder
		err := tmpl.ExecuteTemplate(&b, "auth/index_nuon.tmpl", map[string]any{
			"IsAuthenticated": false,
			"Email":           "",
			"Providers":       providers,
			"RedirectURL":     "",
			"DashboardURL":    "https://app.nuon.co",
			"PostHogKey":      "test-project-key",
			"PostHogHost":     "https://us.i.posthog.com",
		})
		require.NoError(t, err)
		assert.Contains(t, b.String(), "posthog.init('test-project-key'")
		assert.Contains(t, b.String(), "us.i.posthog.com")
	})
}
