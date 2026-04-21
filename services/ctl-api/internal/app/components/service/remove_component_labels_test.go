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

func (s *ComponentsServiceTestSuite) TestRemoveComponentLabelsSuccess() {
	s.Run("removes specified keys", func() {
		componentID := s.testAppConfig.ComponentConfigConnections[3].ComponentID

		// Set initial labels
		err := s.deps.DB.WithContext(s.ctx).
			Model(&app.Component{}).
			Where("id = ?", componentID).
			Update("labels", labels.Labels{"env": "prod", "team": "platform", "region": "us-west-2"}).Error
		require.NoError(s.T(), err)

		reqBody := RemoveComponentLabelsRequest{
			Keys: []string{"team"},
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "us-west-2", response.Labels["region"])
		_, hasTeam := response.Labels["team"]
		assert.False(s.T(), hasTeam)

		// Verify in DB
		var dbComp app.Component
		err = s.deps.DB.WithContext(s.ctx).First(&dbComp, "id = ?", componentID).Error
		require.NoError(s.T(), err)
		_, hasTeam = dbComp.Labels["team"]
		assert.False(s.T(), hasTeam)
	})

	s.Run("removing non-existent key succeeds silently", func() {
		componentID := s.testAppConfig.ComponentConfigConnections[4].ComponentID

		err := s.deps.DB.WithContext(s.ctx).
			Model(&app.Component{}).
			Where("id = ?", componentID).
			Update("labels", labels.Labels{"env": "prod"}).Error
		require.NoError(s.T(), err)

		reqBody := RemoveComponentLabelsRequest{
			Keys: []string{"nonexistent"},
		}
		path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, componentID)
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Component
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})
}

func (s *ComponentsServiceTestSuite) TestRemoveComponentLabelsValidationErrors() {
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
				rr = s.makeRawRequest(http.MethodDelete, path, tc.rawBody)
			} else {
				rr = s.makeRequest(http.MethodDelete, path, tc.body)
			}

			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

func (s *ComponentsServiceTestSuite) TestRemoveComponentLabelsNotFound() {
	reqBody := RemoveComponentLabelsRequest{
		Keys: []string{"env"},
	}
	path := fmt.Sprintf("/v1/apps/%s/components/%s/labels", s.testApp.ID, "cmp_nonexistent00000000000")
	rr := s.makeRequest(http.MethodDelete, path, reqBody)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
