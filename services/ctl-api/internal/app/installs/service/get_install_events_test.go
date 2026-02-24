package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestGetInstallEventsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/events", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallEventsNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000/events", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallEventNotFound() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/events/iev_nonexistent_00000000", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
