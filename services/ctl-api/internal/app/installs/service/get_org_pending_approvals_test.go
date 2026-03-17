package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestGetOrgPendingApprovalsEmpty() {
	rr := s.makeRequest(http.MethodGet, "/v1/workflows/pending-approvals", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var result []app.WorkflowStepApproval
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Empty(s.T(), result)
}

func (s *InstallsServiceTestSuite) TestGetOrgPendingApprovals() {
	type testCase struct {
		name          string
		setupApproval func() *app.WorkflowStepApproval
		setupResponse func(approvalID string)
		wantInResult  bool
	}

	tcs := []testCase{
		{
			name: "pending approval with no response is returned",
			setupApproval: func() *app.WorkflowStepApproval {
				install := s.createTestInstall()
				workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
				step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
				return s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "plan output")
			},
			setupResponse: nil,
			wantInResult:  true,
		},
		{
			name: "approval with active response is not returned",
			setupApproval: func() *app.WorkflowStepApproval {
				install := s.createTestInstall()
				workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
				step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
				return s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "plan output")
			},
			setupResponse: func(approvalID string) {
				resp := &app.WorkflowStepApprovalResponse{
					InstallWorkflowStepApprovalID: approvalID,
					Type:                          app.WorkflowStepApprovalResponseTypeApprove,
					Note:                          "approved",
				}
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(resp).Error)
			},
			wantInResult: false,
		},
		{
			name: "approval with soft-deleted response is returned",
			setupApproval: func() *app.WorkflowStepApproval {
				install := s.createTestInstall()
				workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
				step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
				return s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "plan output")
			},
			setupResponse: func(approvalID string) {
				resp := &app.WorkflowStepApprovalResponse{
					InstallWorkflowStepApprovalID: approvalID,
					Type:                          app.WorkflowStepApprovalResponseTypeApprove,
					Note:                          "approved then revoked",
				}
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(resp).Error)
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(resp).Error)
			},
			wantInResult: true,
		},
	}

	for _, tc := range tcs {
		s.Run(tc.name, func() {
			approval := tc.setupApproval()
			if tc.setupResponse != nil {
				tc.setupResponse(approval.ID)
			}

			rr := s.makeRequest(http.MethodGet, "/v1/workflows/pending-approvals", nil)
			require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

			var result []app.WorkflowStepApproval
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))

			found := false
			for _, a := range result {
				if a.ID == approval.ID {
					found = true
					assert.Equal(s.T(), approval.Type, a.Type)
					break
				}
			}
			assert.Equal(s.T(), tc.wantInResult, found, "approval %s: wantInResult=%v", approval.ID, tc.wantInResult)
		})
	}
}

func (s *InstallsServiceTestSuite) TestGetOrgPendingApprovalsPagination() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)

	var approvalIDs []string
	for i := 0; i < 3; i++ {
		step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
		approval := s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "plan output")
		approvalIDs = append(approvalIDs, approval.ID)
	}

	rr := s.makeRequest(http.MethodGet, "/v1/workflows/pending-approvals?limit=2", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())
	assert.Equal(s.T(), "true", rr.Header().Get("X-Nuon-Page-Next"))

	var page1 []app.WorkflowStepApproval
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &page1))
	assert.Len(s.T(), page1, 2)

	rr2 := s.makeRequest(http.MethodGet, "/v1/workflows/pending-approvals?limit=2&offset=2", nil)
	require.Equal(s.T(), http.StatusOK, rr2.Code, "body: %s", rr2.Body.String())
	assert.Equal(s.T(), "false", rr2.Header().Get("X-Nuon-Page-Next"))

	var page2 []app.WorkflowStepApproval
	require.NoError(s.T(), json.Unmarshal(rr2.Body.Bytes(), &page2))
	assert.Len(s.T(), page2, 1)

	allIDs := make([]string, 0, 3)
	for _, a := range append(page1, page2...) {
		allIDs = append(allIDs, a.ID)
	}
	for _, id := range approvalIDs {
		assert.Contains(s.T(), allIDs, id)
	}
}

func (s *InstallsServiceTestSuite) TestGetOrgPendingApprovalsIsolation() {
	install := s.createTestInstall()
	workflow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), install.ID, app.WorkflowTypeReprovision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), workflow.ID)
	otherOrgApproval := s.deps.Seeder.CreateWorkflowStepApproval(s.ctx, s.T(), step.ID, app.TerraformPlanApprovalType, "other org plan")

	ctx2, acc2 := s.deps.Seeder.EnsureAccount(context.Background(), s.T())
	_, org2 := s.deps.Seeder.EnsureOrg(ctx2, s.T())

	router2 := tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: org2,
		TestAcc: acc2,
	})
	require.NoError(s.T(), s.installsService.RegisterPublicRoutes(router2))

	req, err := http.NewRequest(http.MethodGet, "/v1/workflows/pending-approvals", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router2.ServeHTTP(rr, req)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var result []app.WorkflowStepApproval
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &result))

	for _, a := range result {
		assert.NotEqual(s.T(), otherOrgApproval.ID, a.ID, "approval from another org leaked into results")
	}
}
