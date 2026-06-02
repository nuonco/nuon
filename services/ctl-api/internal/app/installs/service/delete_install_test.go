package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestDeleteInstallSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s", install.ID)
	rr := s.makeRequest(http.MethodDelete, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	captured := tests.GetQueueSignals(s.T(), s.deps.DB)
	require.GreaterOrEqual(s.T(), len(captured), 2)

	var signalTypes []string
	for _, c := range captured {
		signalTypes = append(signalTypes, string(c.Type))
	}
	assert.Contains(s.T(), signalTypes, "execute-workflow")
	assert.Contains(s.T(), signalTypes, "forgotten")
}

func (s *InstallsServiceTestSuite) TestDeleteInstallNotFound() {
	rr := s.makeRequest(http.MethodDelete, "/v1/installs/ins_nonexistent_00000000", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
