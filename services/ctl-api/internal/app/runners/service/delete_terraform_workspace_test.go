package service

import (
	"context"
	"fmt"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type DeleteTerraformWorkspaceTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type DeleteTerraformWorkspaceTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeleteTerraformWorkspaceTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestDeleteTerraformWorkspaceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeleteTerraformWorkspaceTestSuite))
}

func (s *DeleteTerraformWorkspaceTestSuite) SetupSuite() {
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

func (s *DeleteTerraformWorkspaceTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes (needs org context)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeleteTerraformWorkspaceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeleteTerraformWorkspaceTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *DeleteTerraformWorkspaceTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeleteTerraformWorkspaceTestSuite) TestDeleteTerraformWorkspace() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "successfully delete workspace",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ws := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   domains.NewInstallID(),
					OwnerType: "install",
				}
				err := s.service.DB.WithContext(ctx).Create(ws).Error
				require.NoError(s.T(), err)

				return ws.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify workspace is deleted (soft delete)
				var ws app.TerraformWorkspace
				err := s.service.DB.Where("id = ?", workspaceID).First(&ws).Error
				assert.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
			},
		},
		{
			name: "workspace not found returns error",
			setupFunc: func() string {
				return "twsnonexistent123456789012"
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "workspace in different org not accessible",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/other",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

				// Create workspace in org2
				ws2 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     org2.ID,
					OwnerID:   domains.NewInstallID(),
					OwnerType: "install",
				}
				err = s.service.DB.WithContext(ctx).Create(ws2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return ws2.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify workspace still exists (wasn't deleted)
				var ws app.TerraformWorkspace
				err := s.service.DB.Unscoped().Where("id = ?", workspaceID).First(&ws).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), workspaceID, ws.ID)
			},
		},
		{
			name: "delete workspace with associated states succeeds",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ws := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   domains.NewInstallID(),
					OwnerType: "install",
				}
				err := s.service.DB.WithContext(ctx).Create(ws).Error
				require.NoError(s.T(), err)

				// Create associated state
				state := &app.TerraformWorkspaceState{
					ID:                   domains.NewTerraformWorkspaceStateID(),
					OrgID:                s.testOrg.ID,
					TerraformWorkspaceID: ws.ID,
					Contents:             []byte(`{"version": 4}`),
				}
				err = s.service.DB.WithContext(ctx).Create(state).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Where("terraform_workspace_id = ?", ws.ID).Delete(&app.TerraformWorkspaceState{})
				})

				return ws.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify workspace is soft-deleted (not visible in normal query)
				var ws app.TerraformWorkspace
				err := s.service.DB.Where("id = ?", workspaceID).First(&ws).Error
				assert.Error(s.T(), err)

				// Associated states are NOT cascade-deleted (handler only deletes workspace)
				var states []app.TerraformWorkspaceState
				err = s.service.DB.Where("terraform_workspace_id = ?", workspaceID).Find(&states).Error
				require.NoError(s.T(), err)
				assert.NotEmpty(s.T(), states, "associated states should still exist after workspace soft-delete")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID := tc.setupFunc()
			rr := s.makeRequest("DELETE", "/v1/terraform-workspaces/"+workspaceID)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(workspaceID)
			}
		})
	}
}
