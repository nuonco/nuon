package service

import (
	"context"
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

type GetTerraformStateTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetTerraformStateTestSuite struct {
	tests.BaseDBTestSuite
	app       *fxtest.App
	service   GetTerraformStateTestService
	router    *gin.Engine
	testOrg   *app.Org
	testAcc   *app.Account
	testWS    *app.TerraformWorkspace
	testState *app.TerraformWorkspaceState
}

func TestGetTerraformStateSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetTerraformStateTestSuite))
}

func (s *GetTerraformStateTestSuite) SetupSuite() {
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

func (s *GetTerraformStateTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes (terraform backend needs org context)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetTerraformStateTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetTerraformStateTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create terraform workspace
	s.testWS = &app.TerraformWorkspace{
		ID:        domains.NewTerraformWorkspaceID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testOrg.ID,
		OwnerType: "org",
	}
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	err := s.service.DB.WithContext(ctx).Create(s.testWS).Error
	require.NoError(s.T(), err)

	// Create terraform state
	stateContents := []byte(`{"version": 4, "terraform_version": "1.0.0"}`)
	s.testState = &app.TerraformWorkspaceState{
		ID:                   domains.NewTerraformWorkspaceStateID(),
		OrgID:                s.testOrg.ID,
		TerraformWorkspaceID: s.testWS.ID,
		Contents:             stateContents,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testState).Error
	require.NoError(s.T(), err)
}

func (s *GetTerraformStateTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetTerraformStateTestSuite) TestGetTerraformState() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "successfully get terraform state",
			setupFunc: func() string {
				return s.testWS.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(body string) {
				assert.Contains(s.T(), body, "terraform_version")
				assert.Contains(s.T(), body, "1.0.0")
			},
		},
		{
			name: "workspace not found returns error",
			setupFunc: func() string {
				return "twsnonexistent123456789012"
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "workspace in different org not accessible",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
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
					OwnerID:   org2.ID,
					OwnerType: "org",
				}
				err = s.service.DB.WithContext(ctx).Create(ws2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return ws2.ID
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "workspace exists but no state returns 204",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create workspace without state - use unique owner
				ownerID := domains.NewAppID()
				ws := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   ownerID,
					OwnerType: "app",
				}
				err := s.service.DB.WithContext(ctx).Create(ws).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name: "workspace with empty state contents returns 204",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Use unique owner ID
				ownerID := domains.NewInstallID()
				ws := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   ownerID,
					OwnerType: "install",
				}
				err := s.service.DB.WithContext(ctx).Create(ws).Error
				require.NoError(s.T(), err)

				// Create state with empty contents
				state := &app.TerraformWorkspaceState{
					ID:                   domains.NewTerraformWorkspaceStateID(),
					OrgID:                s.testOrg.ID,
					TerraformWorkspaceID: ws.ID,
					Contents:             []byte{},
				}
				err = s.service.DB.WithContext(ctx).Create(state).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(state)
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID
			},
			expectedCode: http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/terraform-backend?workspace_id="+workspaceID)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(rr.Body.String())
			}
		})
	}
}
