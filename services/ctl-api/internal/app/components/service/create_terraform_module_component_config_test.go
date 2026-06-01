package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	configcreated "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/configcreated"
	updatecomptype "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/updatecomponenttype"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppTerraformModuleConfigSuccess() {
	s.Run("creates config with public git VCS", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)

		varVal := "bar"
		envVal := "baz"
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/terraform-module", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateTerraformModuleComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Variables:   map[string]*string{"foo": &varVal},
			EnvVars:     map[string]*string{"ENV": &envVal},
			basicVCSConfigRequest: basicVCSConfigRequest{
				PublicGitVCSConfig: &PublicGitVCSConfigRequest{
					Repo:      "owner/repo",
					Directory: ".",
					Branch:    "main",
				},
			},
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.TerraformModuleComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.NotNil(s.T(), response.PublicGitVCSConfig)
		assert.Equal(s.T(), "1.14.6", response.Version)
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppTerraformModuleConfigValidationErrors() {
	comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)
	path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/terraform-module", s.testApp.ID, comp.ID)

	varVal := "bar"
	envVal := "baz"

	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "missing variables",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"env_vars":      map[string]*string{"ENV": &envVal},
				"public_git_vcs_config": map[string]interface{}{
					"repo":      "owner/repo",
					"directory": ".",
					"branch":    "main",
				},
			},
		},
		{
			name: "missing env_vars",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"variables":     map[string]*string{"foo": &varVal},
				"public_git_vcs_config": map[string]interface{}{
					"repo":      "owner/repo",
					"directory": ".",
					"branch":    "main",
				},
			},
		},
		{
			name: "no VCS config",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"variables":     map[string]*string{"foo": &varVal},
				"env_vars":      map[string]*string{"ENV": &envVal},
			},
		},
		{
			name:    "invalid JSON",
			rawBody: "{invalid json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var rr *httptest.ResponseRecorder
			if tc.rawBody != "" {
				rr = s.makeRawRequest(http.MethodPost, path, tc.rawBody)
			} else {
				rr = s.makeRequest(http.MethodPost, path, tc.body)
			}

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// Signals
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppTerraformModuleConfigSignals() {
	s.Run("sends OperationConfigCreated and OperationUpdateComponentType signals", func() {

		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)

		varVal := "bar"
		envVal := "baz"
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/terraform-module", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateTerraformModuleComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Variables:   map[string]*string{"foo": &varVal},
			EnvVars:     map[string]*string{"ENV": &envVal},
			basicVCSConfigRequest: basicVCSConfigRequest{
				PublicGitVCSConfig: &PublicGitVCSConfigRequest{
					Repo:      "owner/repo",
					Directory: ".",
					Branch:    "main",
				},
			},
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		capturedSignals := tests.GetQueueSignals(s.T(), s.deps.DB)
		require.Len(s.T(), capturedSignals, 2, "expected 2 signals")

		assert.Equal(s.T(), configcreated.SignalType, capturedSignals[0].Type)

		assert.Equal(s.T(), updatecomptype.SignalType, capturedSignals[1].Type)
		sig1, ok := capturedSignals[1].Signal.Signal.(*updatecomptype.Signal)
		require.True(s.T(), ok)
		assert.Equal(s.T(), app.ComponentTypeTerraformModule, sig1.ComponentType)
	})
}
