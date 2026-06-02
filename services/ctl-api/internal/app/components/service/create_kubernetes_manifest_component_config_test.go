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

func (s *ComponentsServiceTestSuite) TestCreateAppK8sManifestConfigInlineSuccess() {
	s.Run("creates config with inline manifest", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeKubernetesManifest)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/kubernetes-manifest", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateKubernetesManifestComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Manifest:    "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test",
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.KubernetesManifestComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Contains(s.T(), response.Manifest, "kind: Pod")
	})
}

func (s *ComponentsServiceTestSuite) TestCreateAppK8sManifestConfigWithNamespace() {
	s.Run("creates config with namespace", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeKubernetesManifest)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/kubernetes-manifest", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateKubernetesManifestComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Manifest:    "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test",
			Namespace:   "my-namespace",
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.KubernetesManifestComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "my-namespace", response.Namespace)
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppK8sManifestConfigValidationErrors() {
	comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeKubernetesManifest)
	path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/kubernetes-manifest", s.testApp.ID, comp.ID)

	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "neither manifest nor kustomize",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
			},
		},
		{
			name: "both manifest and kustomize",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"manifest":      "apiVersion: v1\nkind: Pod",
				"kustomize": map[string]interface{}{
					"path": "./overlay",
				},
			},
		},
		{
			name: "inline manifest with VCS config",
			body: map[string]interface{}{
				"app_config_id": s.testAppConfig.ID,
				"manifest":      "apiVersion: v1\nkind: Pod",
				"public_git_vcs_config": map[string]interface{}{
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

func (s *ComponentsServiceTestSuite) TestCreateAppK8sManifestConfigSignals() {
	s.Run("sends OperationConfigCreated and OperationUpdateComponentType signals", func() {

		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeKubernetesManifest)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/kubernetes-manifest", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateKubernetesManifestComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Manifest:    "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test",
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		capturedSignals := tests.GetQueueSignals(s.T(), s.deps.DB)
		require.Len(s.T(), capturedSignals, 2, "expected 2 signals")

		assert.Equal(s.T(), configcreated.SignalType, capturedSignals[0].Type)

		assert.Equal(s.T(), updatecomptype.SignalType, capturedSignals[1].Type)
		sig1, ok := capturedSignals[1].Signal.Signal.(*updatecomptype.Signal)
		require.True(s.T(), ok)
		assert.Equal(s.T(), app.ComponentTypeKubernetesManifest, sig1.ComponentType)
	})
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppK8sManifestConfigNotFound() {
	s.Run("nonexistent component id", func() {
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/kubernetes-manifest", s.testApp.ID, "cmp_nonexistent00000000000")
		rr := s.makeRequest(http.MethodPost, path, CreateKubernetesManifestComponentConfigRequest{
			AppConfigID: s.testAppConfig.ID,
			Manifest:    "apiVersion: v1\nkind: Pod",
		})

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
