package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

func (s *InstallsServiceTestSuite) TestSyncSecretsSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/sync-secrets", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	captured := s.mockEvClient.GetSignals()
	var found bool
	for _, c := range captured {
		if sig, ok := c.Signal.(*signals.Signal); ok && sig.Type == signals.OperationExecuteFlow {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "expected OperationExecuteFlow signal")
}

func (s *InstallsServiceTestSuite) TestSyncSecretsNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent_00000000/sync-secrets", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
