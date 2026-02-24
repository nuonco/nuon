package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetWorkflowsSuccess() {
	result := s.createTestInstallViaAPI()

	path := fmt.Sprintf("/v1/installs/%s/workflows", result.Install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var workflows []app.Workflow
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &workflows))
	assert.GreaterOrEqual(s.T(), len(workflows), 1)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/workflows", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var workflows []app.Workflow
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &workflows))
	assert.Empty(s.T(), workflows)
}
