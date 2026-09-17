package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCreateAppTemplateUsesPublicAPIURLForSchema(t *testing.T) {
	currentApp := &app.App{Name: "sample"}

	t.Run("byoc host", func(t *testing.T) {
		s := &service{cfg: &internal.Config{PublicAPIURL: "https://api.acme.example.com/"}}
		tmpl, err := s.createAppTemplate(currentApp, AppConfigTemplateTypeHelm)
		require.NoError(t, err)
		assert.Contains(t, tmpl.Content, "#:schema https://api.acme.example.com/v1/general/config-schema/helm")
		assert.NotContains(t, tmpl.Content, cloudSchemaBaseURL)
	})

	t.Run("nuon cloud unchanged", func(t *testing.T) {
		s := &service{cfg: &internal.Config{PublicAPIURL: cloudSchemaBaseURL}}
		tmpl, err := s.createAppTemplate(currentApp, AppConfigTemplateTypeHelm)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(tmpl.Content, "#:schema https://api.nuon.co/v1/general/config-schema/helm"))
	})

	t.Run("empty config falls back to cloud", func(t *testing.T) {
		s := &service{}
		tmpl, err := s.createAppTemplate(currentApp, AppConfigTemplateTypeInstaller)
		require.NoError(t, err)
		assert.Contains(t, tmpl.Content, "#:schema https://api.nuon.co/v1/general/config-schema/installer")
	})
}
