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

type GetTerraformWorkspaceTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetTerraformWorkspaceTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service GetTerraformWorkspaceTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestGetTerraformWorkspaceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetTerraformWorkspaceTestSuite))
}

func (s *GetTerraformWorkspaceTestSuite) SetupSuite() {
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

func (s *GetTerraformWorkspaceTestSuite) SetupTest() {
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

func (s *GetTerraformWorkspaceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetTerraformWorkspaceTestSuite) setupTestData() {
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
}

func (s *GetTerraformWorkspaceTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetTerraformWorkspaceTestSuite) TestGetTerraformWorkspace() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.TerraformWorkspace)
		expectedNotFound bool
	}{
		{
			name: "successfully get workspace by ID",
			setupFunc: func() string {
				return s.testWS.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(ws *app.TerraformWorkspace) {
				assert.Equal(s.T(), s.testWS.ID, ws.ID)
				assert.Equal(s.T(), s.testOrg.ID, ws.OrgID)
				assert.Equal(s.T(), s.testOrg.ID, ws.OwnerID)
				assert.Equal(s.T(), "org", ws.OwnerType)
			},
		},
		{
			name: "workspace not found returns error",
			setupFunc: func() string {
				return "twsnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
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
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "multiple workspaces exist, correct one returned by ID",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create additional workspaces
				ws2 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   "apptest123456789012345678",
					OwnerType: "app",
				}
				err := s.service.DB.WithContext(ctx).Create(ws2).Error
				require.NoError(s.T(), err)

				ws3 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   "instest123456789012345678",
					OwnerType: "install",
				}
				err = s.service.DB.WithContext(ctx).Create(ws3).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws2)
					s.service.DB.Unscoped().Delete(ws3)
				})

				return ws2.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(ws *app.TerraformWorkspace) {
				assert.Equal(s.T(), "apptest123456789012345678", ws.OwnerID)
				assert.Equal(s.T(), "app", ws.OwnerType)
			},
		},
		{
			name: "workspace with different owner type",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ws := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   "cmptest123456789012345678",
					OwnerType: "component",
				}
				err := s.service.DB.WithContext(ctx).Create(ws).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(ws *app.TerraformWorkspace) {
				assert.Equal(s.T(), "component", ws.OwnerType)
				assert.Equal(s.T(), s.testOrg.ID, ws.OrgID)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/terraform-workspaces/"+workspaceID)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var workspace app.TerraformWorkspace
				err := json.Unmarshal(rr.Body.Bytes(), &workspace)
				require.NoError(s.T(), err)
				tc.validateFunc(&workspace)
			}
		})
	}
}
