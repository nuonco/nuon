package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppHelmConfigSuccess() {
	s.Run("creates config with public git VCS and chart_name", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeHelmChart)

		valVal := "bar"
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/helm", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateHelmComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			ChartName:   "my-chart",
			Values:      map[string]*string{"foo": &valVal},
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

		var response app.HelmComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.NotNil(s.T(), response.PublicGitVCSConfig)
		assert.NotNil(s.T(), response.HelmConfig)
		assert.Equal(s.T(), "my-chart", response.HelmConfig.ChartName)
	})
}

func (s *ComponentsServiceTestSuite) TestCreateAppHelmConfigWithHelmRepo() {
	s.Run("creates config with helm_repo_config instead of VCS", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeHelmChart)

		valVal := "bar"
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/helm", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateHelmComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			ChartName:   "my-chart",
			Values:      map[string]*string{"foo": &valVal},
			HelmRepoConfig: &HelmRepoConfigRequest{
				RepoURL: "https://charts.example.com",
				Chart:   "my-chart",
				Version: "1.0.0",
			},
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.HelmComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.NotNil(s.T(), response.HelmConfig)
		assert.Equal(s.T(), "my-chart", response.HelmConfig.ChartName)
		assert.NotNil(s.T(), response.HelmConfig.HelmRepoConfig)
		assert.Equal(s.T(), "https://charts.example.com", response.HelmConfig.HelmRepoConfig.RepoURL)
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppHelmConfigValidationErrors() {
	comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeHelmChart)
	path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/helm", s.testApp.ID, comp.ID)

	valVal := "bar"

	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "missing chart_name",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"values":        map[string]*string{"foo": &valVal},
				"public_git_vcs_config": map[string]interface{}{
					"repo":      "owner/repo",
					"directory": ".",
					"branch":    "main",
				},
			},
		},
		{
			name: "chart_name too short",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"chart_name":    "ab",
				"values":        map[string]*string{"foo": &valVal},
				"public_git_vcs_config": map[string]interface{}{
					"repo":      "owner/repo",
					"directory": ".",
					"branch":    "main",
				},
			},
		},
		{
			name: "missing values",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"chart_name":    "my-chart",
				"public_git_vcs_config": map[string]interface{}{
					"repo":      "owner/repo",
					"directory": ".",
					"branch":    "main",
				},
			},
		},
		{
			name: "no VCS and no helm_repo",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"chart_name":    "my-chart",
				"values":        map[string]*string{"foo": &valVal},
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

func (s *ComponentsServiceTestSuite) TestCreateAppHelmConfigSignals() {
	s.Run("sends OperationConfigCreated and OperationUpdateComponentType signals", func() {
		s.mockEvClient.Reset()

		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeHelmChart)

		valVal := "bar"
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/helm", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateHelmComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			ChartName:   "my-chart",
			Values:      map[string]*string{"foo": &valVal},
			basicVCSConfigRequest: basicVCSConfigRequest{
				PublicGitVCSConfig: &PublicGitVCSConfigRequest{
					Repo:      "owner/repo",
					Directory: ".",
					Branch:    "main",
				},
			},
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		capturedSignals := s.mockEvClient.GetSignals()
		require.Len(s.T(), capturedSignals, 2, "expected 2 signals")

		sig0, ok := capturedSignals[0].Signal.(*signals.Signal)
		require.True(s.T(), ok)
		assert.Equal(s.T(), signals.OperationConfigCreated, sig0.Type)

		sig1, ok := capturedSignals[1].Signal.(*signals.Signal)
		require.True(s.T(), ok)
		assert.Equal(s.T(), signals.OperationUpdateComponentType, sig1.Type)
		assert.Equal(s.T(), app.ComponentTypeHelmChart, sig1.ComponentType)
	})
}
