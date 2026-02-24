package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestUpdateAppComponentSuccess() {
	// Each subtest uses a different seeded component to avoid mutation interference.
	// Index: 0=helm, 1=terraform, 2=docker, 3=k8s, 4=external_image, 5=job

	s.Run("updates name", func() {
		// Use helm component (index 0)
		componentID := s.testAppConfig.ComponentConfigConnections[0].ComponentID

		reqBody := UpdateComponentRequest{
			Name: "updated_name",
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodPatch, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "updated_name", response.Name)

		var dbComp app.Component
		err = s.deps.DB.WithContext(s.ctx).First(&dbComp, "id = ?", response.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "updated_name", dbComp.Name)
		assert.Equal(s.T(), s.testApp.ID, dbComp.AppID)
	})

	s.Run("updates name and var_name", func() {
		// Use terraform component (index 1)
		componentID := s.testAppConfig.ComponentConfigConnections[1].ComponentID

		reqBody := UpdateComponentRequest{
			Name:    "new_name",
			VarName: "new_var_name",
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodPatch, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "new_name", response.Name)
		assert.Equal(s.T(), "new_var_name", response.VarName)

		var dbComp app.Component
		err = s.deps.DB.WithContext(s.ctx).First(&dbComp, "id = ?", response.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "new_name", dbComp.Name)
		assert.Equal(s.T(), "new_var_name", dbComp.VarName)
	})

	s.Run("updates dependencies", func() {
		// Use docker component (index 2), with k8s (index 3) and external_image (index 4) as deps.
		// Read current names from DB to avoid stale references.
		k8sComponentID := s.testAppConfig.ComponentConfigConnections[3].ComponentID
		extImageComponentID := s.testAppConfig.ComponentConfigConnections[4].ComponentID

		var k8sComp, extImageComp app.Component
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&k8sComp, "id = ?", k8sComponentID).Error)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&extImageComp, "id = ?", extImageComponentID).Error)

		componentID := s.testAppConfig.ComponentConfigConnections[2].ComponentID

		reqBody := UpdateComponentRequest{
			Name:         "updated_with_deps",
			Dependencies: []string{k8sComp.Name, extImageComp.Name},
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodPatch, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		var dependencies []app.ComponentDependency
		err = s.deps.DB.WithContext(s.ctx).
			Where("component_id = ?", response.ID).
			Find(&dependencies).Error
		require.NoError(s.T(), err)
		assert.Len(s.T(), dependencies, 2, "expected 2 dependency records")
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestUpdateAppComponentValidationErrors() {
	// Use a pre-seeded component from the full app config
	seededComponentID := s.testAppConfig.ComponentConfigConnections[0].ComponentID

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
			name: "missing name",
			body: map[string]interface{}{
				"var_name": "foo",
			},
		},
		{
			name: "empty name",
			body: map[string]interface{}{
				"name": "",
			},
		},
		{
			name: "name with spaces",
			body: map[string]interface{}{
				"name": "my component",
			},
		},
		{
			name: "name with hyphens",
			body: map[string]interface{}{
				"name": "my-component",
			},
		},
		{
			name:    "invalid JSON",
			rawBody: "{invalid json",
		},
		{
			name: "invalid var_name",
			body: map[string]interface{}{
				"name":     "valid_name",
				"var_name": "Bad Name",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, seededComponentID)

			var rr *httptest.ResponseRecorder
			if tc.rawBody != "" {
				rr = s.makeRawRequest(http.MethodPatch, path, tc.rawBody)
			} else {
				rr = s.makeRequest(http.MethodPatch, path, tc.body)
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

func (s *ComponentsServiceTestSuite) TestUpdateAppComponentNotFound() {
	s.Run("nonexistent component id", func() {
		reqBody := UpdateComponentRequest{
			Name: "new_name",
		}

		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, "cmp_nonexistent00000000000")
		rr := s.makeRequest(http.MethodPatch, path, reqBody)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
