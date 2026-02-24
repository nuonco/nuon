package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

func (s *InstallsServiceTestSuite) TestForgetInstallSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/forget", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	captured := s.mockEvClient.GetSignals()
	require.Len(s.T(), captured, 1)
	sig, ok := captured[0].Signal.(*signals.Signal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), signals.OperationForget, sig.Type)

	// Verify the install is gone
	getPath := fmt.Sprintf("/v1/installs/%s", install.ID)
	getRR := s.makeRequest(http.MethodGet, getPath, nil)
	require.Equal(s.T(), http.StatusNotFound, getRR.Code)
}

func (s *InstallsServiceTestSuite) TestForgetInstallNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent_00000000/forget", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
