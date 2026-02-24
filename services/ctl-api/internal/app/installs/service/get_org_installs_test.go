package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetOrgInstallsEmpty() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(s.T(), resp)
}

func (s *InstallsServiceTestSuite) TestGetOrgInstallsReturnsList() {
	s.createTestInstall()
	s.createTestInstall()

	rr := s.makeRequest(http.MethodGet, "/v1/installs", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)
}

func (s *InstallsServiceTestSuite) TestGetOrgInstallsSearch() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs?q=%s", install.Name)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), install.ID, resp[0].ID)
}
