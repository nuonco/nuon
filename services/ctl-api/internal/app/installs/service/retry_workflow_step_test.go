package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestRetryWorkflowStepNotFound() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	path := fmt.Sprintf("/v1/workflows/%s/steps/nonexistent/retry", workflow.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}
