package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestDeployInstallComponentsSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/components/deploy-all", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var found bool
	captured := tests.GetQueueSignals(s.T(), s.deps.DB)
	for _, c := range captured {
		if string(c.Type) == "execute-workflow" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "expected OperationExecuteFlow signal")
}

func (s *InstallsServiceTestSuite) TestDeployInstallComponentsPlanOnly() {
	install := s.createTestInstall()

	body := DeployInstallComponentsRequest{PlanOnly: true}
	path := fmt.Sprintf("/v1/installs/%s/components/deploy-all", install.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)
}

func (s *InstallsServiceTestSuite) TestDeployInstallComponentsNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent_00000000/components/deploy-all", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
