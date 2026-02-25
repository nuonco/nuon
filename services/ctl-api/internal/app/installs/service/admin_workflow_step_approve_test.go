package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminWorkflowStepApproveTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminWorkflowStepApproveTestSuite struct {
	tests.BaseDBTestSuite
	app         *fxtest.App
	service     AdminWorkflowStepApproveTestService
	router      *gin.Engine
	ctx         context.Context
	testOrg     *app.Org
	testAcc     *app.Account
	testApp     *app.App
	testInstall *app.Install
}

func TestAdminWorkflowStepApproveSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminWorkflowStepApproveTestSuite))
}

func (s *AdminWorkflowStepApproveTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminWorkflowStepApproveTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Admin routes need account context for created_by_id on workflow-related records
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})
	err := s.service.InstallsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminWorkflowStepApproveTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminWorkflowStepApproveTestSuite) setupTestData() {
	s.ctx = context.Background()

	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
	s.service.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(s.ctx, s.T(), s.testApp)
}

func (s *AdminWorkflowStepApproveTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(s.T(), err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminWorkflowStepApproveTestSuite) TestAdminInstallWorkflowStepApprove() {
	testCases := []struct {
		name             string
		setupFunc        func() AdminWorkflowStepApproveRequest
		expectedCode     int
		validateFunc     func(*app.WorkflowStepApprovalResponse)
		expectedNotFound bool
	}{
		{
			name: "successfully approve workflow step",
			setupFunc: func() AdminWorkflowStepApproveRequest {
				ctx := s.ctx

				// Create workflow
				workflow := &app.Workflow{
					ID:        domains.NewWorkflowID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testInstall.ID,
					OwnerType: "installs",
					Type:      app.WorkflowTypeReprovision,
					Status:    app.NewCompositeStatus(ctx, app.StatusPending),
					PlanOnly:  false,
				}
				err := s.service.DB.WithContext(ctx).Create(workflow).Error
				require.NoError(s.T(), err)

				// Create workflow step
				step := &app.WorkflowStep{
					ID:                domains.NewWorkflowStepID(),
					OrgID:             s.testOrg.ID,
					InstallWorkflowID: workflow.ID,
					Name:              "test-step",
					Status:            app.NewCompositeStatus(ctx, app.StatusPending),
				}
				err = s.service.DB.WithContext(ctx).Create(step).Error
				require.NoError(s.T(), err)

				// Create approval
				approval := &app.WorkflowStepApproval{
					OrgID:                 s.testOrg.ID,
					InstallWorkflowStepID: step.ID,
					Type:                  app.TerraformPlanApprovalType,
				}
				err = s.service.DB.WithContext(ctx).Create(approval).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(approval)
					s.service.DB.Unscoped().Delete(step)
					s.service.DB.Unscoped().Delete(workflow)
				})

				return AdminWorkflowStepApproveRequest{
					StepID: step.ID,
				}
			},
			expectedCode: http.StatusOK,
			validateFunc: func(resp *app.WorkflowStepApprovalResponse) {
				assert.NotEmpty(s.T(), resp.ID)
				assert.Equal(s.T(), s.testOrg.ID, resp.OrgID)
				assert.Equal(s.T(), app.WorkflowStepApprovalResponseTypeApprove, resp.Type)
				assert.Equal(s.T(), "Admin approved the step", resp.Note)
			},
		},
		{
			name: "step already approved returns error",
			setupFunc: func() AdminWorkflowStepApproveRequest {
				ctx := s.ctx

				// Create workflow
				workflow := &app.Workflow{
					ID:        domains.NewWorkflowID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testInstall.ID,
					OwnerType: "installs",
					Type:      app.WorkflowTypeReprovision,
					Status:    app.NewCompositeStatus(ctx, app.StatusPending),
					PlanOnly:  false,
				}
				err := s.service.DB.WithContext(ctx).Create(workflow).Error
				require.NoError(s.T(), err)

				// Create workflow step
				step := &app.WorkflowStep{
					ID:                domains.NewWorkflowStepID(),
					OrgID:             s.testOrg.ID,
					InstallWorkflowID: workflow.ID,
					Name:              "test-step",
					Status:            app.NewCompositeStatus(ctx, app.StatusPending),
				}
				err = s.service.DB.WithContext(ctx).Create(step).Error
				require.NoError(s.T(), err)

				// Create approval
				approval := &app.WorkflowStepApproval{
					OrgID:                 s.testOrg.ID,
					InstallWorkflowStepID: step.ID,
					Type:                  app.TerraformPlanApprovalType,
				}
				err = s.service.DB.WithContext(ctx).Create(approval).Error
				require.NoError(s.T(), err)

				// Create existing response
				response := &app.WorkflowStepApprovalResponse{
					ID:                            domains.NewWorkflowStepApprovalResponseID(),
					OrgID:                         s.testOrg.ID,
					InstallWorkflowStepApprovalID: approval.ID,
					Type:                          app.WorkflowStepApprovalResponseTypeApprove,
					Note:                          "Already approved",
				}
				err = s.service.DB.WithContext(ctx).Create(response).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(response)
					s.service.DB.Unscoped().Delete(approval)
					s.service.DB.Unscoped().Delete(step)
					s.service.DB.Unscoped().Delete(workflow)
				})

				return AdminWorkflowStepApproveRequest{
					StepID: step.ID,
				}
			},
			expectedCode:     http.StatusInternalServerError,
			expectedNotFound: true,
		},
		{
			name: "step without approval returns error",
			setupFunc: func() AdminWorkflowStepApproveRequest {
				ctx := s.ctx

				// Create workflow
				workflow := &app.Workflow{
					ID:        domains.NewWorkflowID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testInstall.ID,
					OwnerType: "installs",
					Type:      app.WorkflowTypeReprovision,
					Status:    app.NewCompositeStatus(ctx, app.StatusPending),
					PlanOnly:  false,
				}
				err := s.service.DB.WithContext(ctx).Create(workflow).Error
				require.NoError(s.T(), err)

				// Create workflow step without approval
				step := &app.WorkflowStep{
					ID:                domains.NewWorkflowStepID(),
					OrgID:             s.testOrg.ID,
					InstallWorkflowID: workflow.ID,
					Name:              "test-step",
					Status:            app.NewCompositeStatus(ctx, app.StatusPending),
				}
				err = s.service.DB.WithContext(ctx).Create(step).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(step)
					s.service.DB.Unscoped().Delete(workflow)
				})

				return AdminWorkflowStepApproveRequest{
					StepID: step.ID,
				}
			},
			expectedCode:     http.StatusInternalServerError,
			expectedNotFound: true,
		},
		{
			name: "nonexistent step returns error",
			setupFunc: func() AdminWorkflowStepApproveRequest {
				return AdminWorkflowStepApproveRequest{
					StepID: "wks000000000000000000000000",
				}
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "empty step ID returns not found",
			setupFunc: func() AdminWorkflowStepApproveRequest {
				return AdminWorkflowStepApproveRequest{
					StepID: "",
				}
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/admin-install-workflow-step-approve", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				var resp app.WorkflowStepApprovalResponse
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
				tc.validateFunc(&resp)
			}
		})
	}
}
