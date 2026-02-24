package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

func (s *InstallsServiceTestSuite) TestDeleteInstallSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s", install.ID)
	rr := s.makeRequest(http.MethodDelete, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	assert.NotEmpty(s.T(), rr.Header().Get(app.HeaderInstallWorkflowID))

	captured := s.mockEvClient.GetSignals()
	require.GreaterOrEqual(s.T(), len(captured), 2)

	var signalTypes []string
	for _, c := range captured {
		if sig, ok := c.Signal.(*signals.Signal); ok {
			signalTypes = append(signalTypes, string(sig.Type))
		}
	}
	assert.Contains(s.T(), signalTypes, string(signals.OperationExecuteFlow))
	assert.Contains(s.T(), signalTypes, string(signals.OperationForget))
}

func (s *InstallsServiceTestSuite) TestDeleteInstallNotFound() {
	rr := s.makeRequest(http.MethodDelete, "/v1/installs/ins_nonexistent_00000000", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
