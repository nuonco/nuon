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

func (s *ComponentsServiceTestSuite) TestCreateAppDockerBuildConfigSuccess() {
	s.Run("creates config with public git VCS and dockerfile", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/docker-build", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateDockerBuildComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Dockerfile:  "Dockerfile",
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

		var response app.DockerBuildComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "Dockerfile", response.Dockerfile)
		assert.NotNil(s.T(), response.PublicGitVCSConfig)
	})
}

func (s *ComponentsServiceTestSuite) TestCreateAppDockerBuildConfigWithOptionalFields() {
	s.Run("creates config with target, build_args, and env_vars", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)

		envVal := "bar"
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/docker-build", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateDockerBuildComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Dockerfile:  "Dockerfile.prod",
			Target:      "production",
			BuildArgs:   []string{"ARG1=val1"},
			EnvVars:     map[string]*string{"FOO": &envVal},
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

		var response app.DockerBuildComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "Dockerfile.prod", response.Dockerfile)
		assert.Equal(s.T(), "production", response.Target)
		assert.Equal(s.T(), "ARG1=val1", response.BuildArgs[0])
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppDockerBuildConfigValidationErrors() {
	comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)
	path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/docker-build", s.testApp.ID, comp.ID)

	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "missing dockerfile",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
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
				"dockerfile":    "Dockerfile",
			},
		},
		{
			name: "both public and connected VCS",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"dockerfile":    "Dockerfile",
				"public_git_vcs_config": map[string]interface{}{
					"repo":      "owner/repo",
					"directory": ".",
					"branch":    "main",
				},
				"connected_github_vcs_config": map[string]interface{}{
					"repo":      "owner/repo",
					"directory": ".",
					"branch":    "main",
				},
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

func (s *ComponentsServiceTestSuite) TestCreateAppDockerBuildConfigSignals() {
	s.Run("sends OperationConfigCreated and OperationUpdateComponentType signals", func() {

		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/docker-build", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateDockerBuildComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Dockerfile:  "Dockerfile",
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
		assert.Equal(s.T(), app.ComponentTypeDockerBuild, sig1.ComponentType)
	})
}
