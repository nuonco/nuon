package service

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetWorkflowStepApprovalContentsSuccess() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)

	originalContents := "terraform plan output\n+ resource aws_instance.foo"
	approval := s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, originalContents)

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/approvals/%s/contents", workflow.ID, step.ID, approval.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())
	assert.Equal(s.T(), "gzip", rr.Header().Get("Content-Encoding"))

	// Decompress and verify contents match.
	reader, err := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	require.NoError(s.T(), err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), originalContents, string(decompressed))
}

func (s *InstallsServiceTestSuite) TestGetWorkflowStepApprovalContentsNotFound() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)

	path := fmt.Sprintf("/v1/workflows/%s/steps/%s/approvals/nonexistent/contents", workflow.ID, step.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}
