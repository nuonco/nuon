package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppComponentBuildSuccess() {
	s.Run("create build with git_ref", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds", s.testApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateComponentBuildRequest{
			GitRef: generics.ToPtr("main"),
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.NotEmpty(s.T(), response.ID)
		assert.NotEmpty(s.T(), response.ComponentConfigConnectionID)
		assert.Equal(s.T(), app.ComponentBuildStatus("queued"), response.Status)

		// Verify persisted to DB
		var dbBuild app.ComponentBuild
		err = s.deps.DB.WithContext(s.ctx).First(&dbBuild, "id = ?", response.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), response.ID, dbBuild.ID)
	})
}

func (s *ComponentsServiceTestSuite) TestCreateAppComponentBuildUseLatest() {
	s.Run("create build with use_latest", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds", s.testApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateComponentBuildRequest{
			UseLatest: true,
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.NotEmpty(s.T(), response.ID)
		assert.Equal(s.T(), app.ComponentBuildStatus("queued"), response.Status)
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppComponentBuildValidationErrors() {
	cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
	path := fmt.Sprintf("/v1/apps/%s/components/%s/builds", s.testApp.ID, cmp.ID)

	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "nil body",
			body: nil,
		},
		{
			name: "empty JSON object",
			body: map[string]interface{}{},
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
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppComponentBuildNotFound() {
	s.Run("nonexistent component id", func() {
		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds", s.testApp.ID, "cmp_nonexistent00000000000")
		rr := s.makeRequest(http.MethodPost, path, CreateComponentBuildRequest{
			GitRef: generics.ToPtr("main"),
		})

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}

// ---------------------------------------------------------------------------
// Signals
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppComponentBuildSignals() {
	s.Run("sends OperationBuild signal", func() {
		s.mockEvClient.Reset()

		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds", s.testApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateComponentBuildRequest{
			GitRef: generics.ToPtr("main"),
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		capturedSignals := s.mockEvClient.GetSignals()
		require.Len(s.T(), capturedSignals, 1, "expected 1 signal")

		assert.Equal(s.T(), cmp.ID, capturedSignals[0].ID, "signal should target the component")

		sig, ok := capturedSignals[0].Signal.(*signals.Signal)
		require.True(s.T(), ok, "signal should be *signals.Signal")
		assert.Equal(s.T(), signals.OperationBuild, sig.Type)
		assert.Equal(s.T(), response.ID, sig.BuildID)
	})
}
