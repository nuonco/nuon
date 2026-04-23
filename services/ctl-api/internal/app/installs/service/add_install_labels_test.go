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

func (s *InstallsServiceTestSuite) TestAddInstallLabelsSuccess() {
	s.Run("adds labels to install with no existing labels", func() {
		install := s.createTestInstall()

		reqBody := AddInstallLabelsRequest{
			Labels: map[string]string{"env": "prod", "team": "platform"},
		}
		path := fmt.Sprintf("/v1/installs/%s/labels", install.ID)
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Install
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "platform", response.Labels["team"])

		// Verify in DB
		var dbInstall app.Install
		err = s.deps.DB.WithContext(s.ctx).First(&dbInstall, "id = ?", install.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", dbInstall.Labels["env"])
		assert.Equal(s.T(), "platform", dbInstall.Labels["team"])
	})

	s.Run("merges labels with existing labels", func() {
		install := s.createTestInstall()

		// Set initial labels
		install.Labels = labels.Labels{"env": "staging"}
		err := s.deps.DB.WithContext(s.ctx).Model(&install).Select("labels").Updates(&install).Error
		require.NoError(s.T(), err)

		reqBody := AddInstallLabelsRequest{
			Labels: map[string]string{"team": "platform"},
		}
		path := fmt.Sprintf("/v1/installs/%s/labels", install.ID)
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Install
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "staging", response.Labels["env"])
		assert.Equal(s.T(), "platform", response.Labels["team"])
	})

	s.Run("overwrites existing key", func() {
		install := s.createTestInstall()

		// Set initial labels
		install.Labels = labels.Labels{"env": "staging"}
		err := s.deps.DB.WithContext(s.ctx).Model(&install).Select("labels").Updates(&install).Error
		require.NoError(s.T(), err)

		reqBody := AddInstallLabelsRequest{
			Labels: map[string]string{"env": "prod"},
		}
		path := fmt.Sprintf("/v1/installs/%s/labels", install.ID)
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Install
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})
}

func (s *InstallsServiceTestSuite) TestAddInstallLabelsValidationErrors() {
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
				rr = s.makeRawRequest(http.MethodPost, path, tc.rawBody)
			} else {
				rr = s.makeRequest(http.MethodPost, path, tc.body)
			}

			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

func (s *InstallsServiceTestSuite) TestAddInstallLabelsNotFound() {
	reqBody := AddInstallLabelsRequest{
		Labels: map[string]string{"env": "prod"},
	}
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent00000000000/labels", reqBody)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
