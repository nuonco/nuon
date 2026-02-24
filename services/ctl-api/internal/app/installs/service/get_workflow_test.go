package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetWorkflowSuccess() {
	result := s.createTestInstallViaAPI()
	require.NotEmpty(s.T(), result.WorkflowID)

	path := fmt.Sprintf("/v1/workflows/%s", result.WorkflowID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var workflow app.Workflow
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &workflow))
	assert.Equal(s.T(), result.WorkflowID, workflow.ID)
}

func (s *InstallsServiceTestSuite) TestGetWorkflowNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/workflows/iwf_nonexistent_00000000", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
