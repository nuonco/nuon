package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestReprovisionStackSuccess() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/reprovision-stack", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	workflow := s.latestWorkflowOfType(install.ID, app.WorkflowTypeReprovisionStack)
	assert.False(s.T(), workflow.PlanOnly)

	var found bool
	captured := tests.GetQueueSignals(s.T(), s.deps.DB)
	for _, c := range captured {
		if string(c.Type) == "execute-workflow" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "expected execute-workflow signal")
}

func (s *InstallsServiceTestSuite) TestReprovisionStackNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent_00000000/reprovision-stack", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) latestWorkflowOfType(installID string, typ app.WorkflowType) app.Workflow {
	var workflow app.Workflow
	err := s.deps.DB.Where(app.Workflow{
		OwnerID:   installID,
		OwnerType: "installs",
		Type:      typ,
	}).Order("created_at DESC").First(&workflow).Error
	require.NoError(s.T(), err)
	return workflow
}
