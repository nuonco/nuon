package service

import (
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetTerraformWorkspacesTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetTerraformWorkspacesTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service GetTerraformWorkspacesTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestGetTerraformWorkspacesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetTerraformWorkspacesTestSuite))
}

func (s *GetTerraformWorkspacesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetTerraformWorkspacesTestSuite) SetupTest() {
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

func (s *GetTerraformWorkspacesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetTerraformWorkspacesTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *GetTerraformWorkspacesTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetTerraformWorkspacesTestSuite) TestGetTerraformWorkspaces() {
	testCases := []struct {
		name          string
		setupFunc     func() []string
		expectedCount int
		expectedCode  int
		validateFunc  func([]app.TerraformWorkspace)
	}{
		{
			name: "returns empty array when no workspaces",
			setupFunc: func() []string {
				return []string{}
			},
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "returns workspaces for org",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ws1 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   domains.NewInstallID(),
					OwnerType: "install",
				}
				err := s.service.DB.WithContext(ctx).Create(ws1).Error
				require.NoError(s.T(), err)

				ws2 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   domains.NewInstallID(),
					OwnerType: "install",
				}
				err = s.service.DB.WithContext(ctx).Create(ws2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws1)
					s.service.DB.Unscoped().Delete(ws2)
				})

				return []string{ws1.ID, ws2.ID}
			},
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(workspaces []app.TerraformWorkspace) {
				for _, ws := range workspaces {
					assert.Equal(s.T(), s.testOrg.ID, ws.OrgID)
					assert.NotEmpty(s.T(), ws.ID)
				}
			},
		},
		{
			name: "doesn't return other org's workspaces",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create workspace in test org
				ws1 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   domains.NewInstallID(),
					OwnerType: "install",
				}
				err := s.service.DB.WithContext(ctx).Create(ws1).Error
				require.NoError(s.T(), err)

				// Create second org
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/other",
					},
				}
				err = s.service.DB.WithContext(ctx).Create(org2).Error
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
					s.service.DB.Unscoped().Delete(ws1)
					s.service.DB.Unscoped().Delete(org2)
				})

				return []string{ws1.ID}
			},
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(workspaces []app.TerraformWorkspace) {
				// Should only see workspace from test org
				assert.Equal(s.T(), 1, len(workspaces))
				assert.Equal(s.T(), s.testOrg.ID, workspaces[0].OrgID)
			},
		},
		{
			name: "multiple workspaces with different owner types",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ws1 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   domains.NewInstallID(),
					OwnerType: "install",
				}
				err := s.service.DB.WithContext(ctx).Create(ws1).Error
				require.NoError(s.T(), err)

				ws2 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   "apptest123456789012345678",
					OwnerType: "app",
				}
				err = s.service.DB.WithContext(ctx).Create(ws2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws1)
					s.service.DB.Unscoped().Delete(ws2)
				})

				return []string{ws1.ID, ws2.ID}
			},
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(workspaces []app.TerraformWorkspace) {
				assert.Equal(s.T(), 2, len(workspaces))
				ownerTypes := make(map[string]bool)
				for _, ws := range workspaces {
					ownerTypes[ws.OwnerType] = true
				}
				assert.True(s.T(), ownerTypes["install"])
				assert.True(s.T(), ownerTypes["app"])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_ = tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/terraform-workspaces")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var workspaces []app.TerraformWorkspace
			err := json.Unmarshal(rr.Body.Bytes(), &workspaces)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tc.expectedCount, len(workspaces))

			if tc.validateFunc != nil {
				tc.validateFunc(workspaces)
			}
		})
	}
}
