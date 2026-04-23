package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *ComponentsServiceTestSuite) TestAddComponentLabelsSuccess() {
	s.Run("adds labels to component with no existing labels", func() {
		componentID := s.testAppConfig.ComponentConfigConnections[0].ComponentID

		reqBody := AddComponentLabelsRequest{
			Labels: map[string]string{"env": "prod", "tier": "frontend"},
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "frontend", response.Labels["tier"])

		// Verify in DB
		var dbComp app.Component
		err = s.deps.DB.WithContext(s.ctx).First(&dbComp, "id = ?", componentID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", dbComp.Labels["env"])
		assert.Equal(s.T(), "frontend", dbComp.Labels["tier"])
	})

	s.Run("merges labels with existing labels", func() {
		componentID := s.testAppConfig.ComponentConfigConnections[1].ComponentID

		// Set initial labels
		err := s.deps.DB.WithContext(s.ctx).
			Model(&app.Component{}).
			Where("id = ?", componentID).
			Update("labels", labels.Labels{"env": "staging"}).Error
		require.NoError(s.T(), err)

		reqBody := AddComponentLabelsRequest{
			Labels: map[string]string{"tier": "backend"},
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "staging", response.Labels["env"])
		assert.Equal(s.T(), "backend", response.Labels["tier"])
	})

	s.Run("overwrites existing key", func() {
		componentID := s.testAppConfig.ComponentConfigConnections[2].ComponentID

		err := s.deps.DB.WithContext(s.ctx).
			Model(&app.Component{}).
			Where("id = ?", componentID).
			Update("labels", labels.Labels{"env": "staging"}).Error
		require.NoError(s.T(), err)

		reqBody := AddComponentLabelsRequest{
			Labels: map[string]string{"env": "prod"},
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})
}

func (s *ComponentsServiceTestSuite) TestAddComponentLabelsValidationErrors() {
	componentID := s.testAppConfig.ComponentConfigConnections[0].ComponentID
	path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, componentID)

	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "empty body",
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

			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

func (s *ComponentsServiceTestSuite) TestAddComponentLabelsNotFound() {
	reqBody := AddComponentLabelsRequest{
		Labels: map[string]string{"env": "prod"},
	}
	path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, "cmp_nonexistent00000000000")
	rr := s.makeRequest(http.MethodPost, path, reqBody)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
