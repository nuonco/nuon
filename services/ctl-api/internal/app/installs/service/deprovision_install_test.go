package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

func (s *InstallsServiceTestSuite) TestDeprovisionInstallSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/deprovision", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	assert.NotEmpty(s.T(), rr.Header().Get(app.HeaderInstallWorkflowID))

	captured := s.mockEvClient.GetSignals()
	require.Len(s.T(), captured, 1)
	sig, ok := captured[0].Signal.(*signals.Signal)
	require.True(s.T(), ok)
	assert.Equal(s.T(), signals.OperationExecuteFlow, sig.Type)
}

func (s *InstallsServiceTestSuite) TestDeprovisionInstallPlanOnly() {
	install := s.createTestInstall()

	body := DeprovisionInstallRequest{PlanOnly: true}

	path := fmt.Sprintf("/v1/installs/%s/deprovision", install.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)
	assert.NotEmpty(s.T(), rr.Header().Get(app.HeaderInstallWorkflowID))
}

func (s *InstallsServiceTestSuite) TestDeprovisionInstallNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent_00000000/deprovision", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
