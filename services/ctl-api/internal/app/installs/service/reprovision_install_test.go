package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestReprovisionInstallSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/reprovision", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	captured := tests.GetQueueSignals(s.T(), s.deps.DB)
	require.Len(s.T(), captured, 1)
	_ = captured[0] // signal type check via .Type

	assert.Equal(s.T(), "ExecuteFlow-type", string(captured[0].Type))
}

func (s *InstallsServiceTestSuite) TestReprovisionInstallNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent_00000000/reprovision", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
