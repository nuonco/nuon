package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestGetInstallSandboxRunsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/sandbox-runs", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallSandboxRunNotFound() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/sandbox-runs/isr_nonexistent_00000000", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
