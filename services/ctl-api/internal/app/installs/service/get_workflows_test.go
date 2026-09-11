package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
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

func (s *InstallsServiceTestSuite) seedWorkflow(installID string, workflowType app.WorkflowType, status app.Status) *app.Workflow {
	return s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), installID, workflowType,
		testseed.WithWorkflowStatus(app.NewCompositeStatus(s.ctx, status)))
}

func (s *InstallsServiceTestSuite) fetchWorkflows(installID, query string) ([]app.Workflow, *httptest.ResponseRecorder) {
	path := fmt.Sprintf("/v1/installs/%s/workflows%s", installID, query)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var workflows []app.Workflow
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &workflows))
	return workflows, rr
}

func workflowIDs(workflows []app.Workflow) []string {
	ids := make([]string, 0, len(workflows))
	for _, workflow := range workflows {
		ids = append(ids, workflow.ID)
	}
	return ids
}

func (s *InstallsServiceTestSuite) TestGetWorkflowsFilterByStatus() {
	install := s.createTestInstall()
	failed := s.seedWorkflow(install.ID, app.WorkflowTypeManualDeploy, app.StatusError)
	succeeded := s.seedWorkflow(install.ID, app.WorkflowTypeManualDeploy, app.StatusSuccess)
	pending := s.seedWorkflow(install.ID, app.WorkflowTypeDriftRun, app.StatusPending)

	s.Run("single status returns only that status", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?status=error")
		require.Equal(s.T(), []string{failed.ID}, workflowIDs(workflows))
	})

	s.Run("several statuses return the union", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?status=error,success")
		require.ElementsMatch(s.T(), []string{failed.ID, succeeded.ID}, workflowIDs(workflows))
	})

	s.Run("unset status returns everything", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "")
		require.ElementsMatch(s.T(), []string{failed.ID, succeeded.ID, pending.ID}, workflowIDs(workflows))
	})

	s.Run("empty status is the same as unset", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?status=%20,")
		require.ElementsMatch(s.T(), []string{failed.ID, succeeded.ID, pending.ID}, workflowIDs(workflows))
	})

	s.Run("unknown status returns an empty list", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?status=not-a-real-status")
		require.Empty(s.T(), workflows)
	})
}

func (s *InstallsServiceTestSuite) TestGetWorkflowsFilterByType() {
	install := s.createTestInstall()
	deploy := s.seedWorkflow(install.ID, app.WorkflowTypeManualDeploy, app.StatusSuccess)
	drift := s.seedWorkflow(install.ID, app.WorkflowTypeDriftRun, app.StatusSuccess)
	sandboxDrift := s.seedWorkflow(install.ID, app.WorkflowTypeDriftRunReprovisionSandbox, app.StatusSuccess)

	s.Run("single type behaves as before", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?type=drift_run")
		require.Equal(s.T(), []string{drift.ID}, workflowIDs(workflows))
	})

	s.Run("several types return the union", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?type=drift_run,drift_run_reprovision_sandbox")
		require.ElementsMatch(s.T(), []string{drift.ID, sandboxDrift.ID}, workflowIDs(workflows))
	})

	s.Run("unknown type returns an empty list", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?type=not_a_real_type")
		require.Empty(s.T(), workflows)
	})

	s.Run("unset type returns everything", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "")
		require.ElementsMatch(s.T(), []string{deploy.ID, drift.ID, sandboxDrift.ID}, workflowIDs(workflows))
	})
}

func (s *InstallsServiceTestSuite) TestGetWorkflowsFiltersIntersect() {
	install := s.createTestInstall()
	failedDrift := s.seedWorkflow(install.ID, app.WorkflowTypeDriftRun, app.StatusError)
	s.seedWorkflow(install.ID, app.WorkflowTypeDriftRun, app.StatusSuccess)
	s.seedWorkflow(install.ID, app.WorkflowTypeManualDeploy, app.StatusError)

	s.Run("status and type intersect", func() {
		workflows, _ := s.fetchWorkflows(install.ID, "?status=error&type=drift_run")
		require.Equal(s.T(), []string{failedDrift.ID}, workflowIDs(workflows))
	})

	s.Run("status and search intersect", func() {
		failedSecrets := s.seedWorkflow(install.ID, app.WorkflowTypeSyncSecrets, app.StatusError)
		s.seedWorkflow(install.ID, app.WorkflowTypeSyncSecrets, app.StatusSuccess)

		workflows, _ := s.fetchWorkflows(install.ID, "?status=error&search=secrets")
		require.Equal(s.T(), []string{failedSecrets.ID}, workflowIDs(workflows))
	})
}

func (s *InstallsServiceTestSuite) TestGetWorkflowsFilteredPagination() {
	install := s.createTestInstall()

	failedIDs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		failedIDs = append(failedIDs, s.seedWorkflow(install.ID, app.WorkflowTypeDriftRun, app.StatusError).ID)
		s.seedWorkflow(install.ID, app.WorkflowTypeDriftRun, app.StatusSuccess)
	}

	firstPage, firstRR := s.fetchWorkflows(install.ID, "?status=error&limit=2")
	require.Len(s.T(), firstPage, 2)
	require.Equal(s.T(), "true", firstRR.Header().Get("X-Nuon-Page-Next"))

	secondPage, secondRR := s.fetchWorkflows(install.ID, "?status=error&limit=2&offset=2")
	require.Len(s.T(), secondPage, 1)
	require.Equal(s.T(), "false", secondRR.Header().Get("X-Nuon-Page-Next"))

	paged := append(workflowIDs(firstPage), workflowIDs(secondPage)...)
	require.ElementsMatch(s.T(), failedIDs, paged)
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
