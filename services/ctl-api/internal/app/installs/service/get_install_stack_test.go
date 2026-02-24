package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetInstallStackSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/stack", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var stack app.InstallStack
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &stack))
	assert.NotEmpty(s.T(), stack.ID)
}

func (s *InstallsServiceTestSuite) TestGetInstallStackRunsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/stack-runs", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)
}
