package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetAppInstallsReturnsList() {
	s.createTestInstall()

	path := fmt.Sprintf("/v1/apps/%s/installs", s.testApp.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 1)
	assert.Equal(s.T(), s.testApp.ID, resp[0].AppID)
}

func (s *InstallsServiceTestSuite) TestGetAppInstallsEmpty() {
	otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())

	path := fmt.Sprintf("/v1/apps/%s/installs", otherApp.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(s.T(), resp)
}
