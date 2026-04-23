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

func (s *InstallsServiceTestSuite) TestRemoveInstallLabelsSuccess() {
	s.Run("removes specified keys", func() {
		install := s.createTestInstall()

		// Set initial labels
		install.Labels = labels.Labels{"env": "prod", "team": "platform", "region": "us-west-2"}
		err := s.deps.DB.WithContext(s.ctx).Model(&install).Select("labels").Updates(&install).Error
		require.NoError(s.T(), err)

		reqBody := RemoveInstallLabelsRequest{
			Keys: []string{"team"},
		}
		path := fmt.Sprintf("/v1/installs/%s/labels", install.ID)
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Install
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "us-west-2", response.Labels["region"])
		_, hasTeam := response.Labels["team"]
		assert.False(s.T(), hasTeam)

		// Verify in DB
		var dbInstall app.Install
		err = s.deps.DB.WithContext(s.ctx).First(&dbInstall, "id = ?", install.ID).Error
		require.NoError(s.T(), err)
		_, hasTeam = dbInstall.Labels["team"]
		assert.False(s.T(), hasTeam)
		assert.Equal(s.T(), "prod", dbInstall.Labels["env"])
	})

	s.Run("removing non-existent key succeeds silently", func() {
		install := s.createTestInstall()

		install.Labels = labels.Labels{"env": "prod"}
		err := s.deps.DB.WithContext(s.ctx).Model(&install).Select("labels").Updates(&install).Error
		require.NoError(s.T(), err)

		reqBody := RemoveInstallLabelsRequest{
			Keys: []string{"nonexistent"},
		}
		path := fmt.Sprintf("/v1/installs/%s/labels", install.ID)
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Install
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})

	s.Run("removes all labels", func() {
		install := s.createTestInstall()

		install.Labels = labels.Labels{"a": "1", "b": "2"}
		err := s.deps.DB.WithContext(s.ctx).Model(&install).Select("labels").Updates(&install).Error
		require.NoError(s.T(), err)

		reqBody := RemoveInstallLabelsRequest{
			Keys: []string{"a", "b"},
		}
		path := fmt.Sprintf("/v1/installs/%s/labels", install.ID)
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Install
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Empty(s.T(), response.Labels)
	})
}

func (s *InstallsServiceTestSuite) TestRemoveInstallLabelsValidationErrors() {
	install := s.createTestInstall()
	path := fmt.Sprintf("/v1/installs/%s/labels", install.ID)

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

func (s *InstallsServiceTestSuite) TestRemoveInstallLabelsNotFound() {
	reqBody := RemoveInstallLabelsRequest{
		Keys: []string{"env"},
	}
	rr := s.makeRequest(http.MethodDelete, "/v1/installs/ins_nonexistent00000000000/labels", reqBody)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
