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

type GetTerraformWorkspaceLockTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetTerraformWorkspaceLockTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service GetTerraformWorkspaceLockTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestGetTerraformWorkspaceLockSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetTerraformWorkspaceLockTestSuite))
}

func (s *GetTerraformWorkspaceLockTestSuite) SetupSuite() {
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

func (s *GetTerraformWorkspaceLockTestSuite) SetupTest() {
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

func (s *GetTerraformWorkspaceLockTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetTerraformWorkspaceLockTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create terraform workspace (use unique OwnerID to avoid unique constraint on owner_id+owner_type)
	s.testWS = &app.TerraformWorkspace{
		ID:        domains.NewTerraformWorkspaceID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   domains.NewInstallID(),
		OwnerType: "install",
	}
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	err := s.service.DB.WithContext(ctx).Create(s.testWS).Error
	require.NoError(s.T(), err)
}

func (s *GetTerraformWorkspaceLockTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetTerraformWorkspaceLockTestSuite) TestGetTerraformWorkspaceLock() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "workspace exists but no lock returns null",
			setupFunc: func() string {
				return s.testWS.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(body string) {
				// Should return null or empty when no lock exists
				if body != "null" && body != "" {
					var lock app.TerraformWorkspaceLock
					err := json.Unmarshal([]byte(body), &lock)
					require.NoError(s.T(), err)
					// If not null, verify it's empty
					assert.Empty(s.T(), lock.ID)
				}
			},
		},
		{
			name: "workspace with lock returns lock data",
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

				// Create lock
				lock := &app.TerraformWorkspaceLock{
					ID:          domains.NewTerraformWorkspaceLockID(),
					WorkspaceID: ws.ID,
					Lock: &app.TerraformLock{
						ID:        "lock-123",
						Operation: "apply",
						Info:      "test-info",
						Who:       "test-user",
						Version:   "1.0.0",
						Created:   "2024-01-01T00:00:00Z",
						Path:      "/tmp/test",
					},
				}
				err = s.service.DB.WithContext(ctx).Create(lock).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(lock)
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(body string) {
				var lock app.TerraformWorkspaceLock
				err := json.Unmarshal([]byte(body), &lock)
				require.NoError(s.T(), err)
				assert.NotEmpty(s.T(), lock.ID)
				assert.Equal(s.T(), "lock-123", lock.Lock.ID)
				assert.Equal(s.T(), "apply", lock.Lock.Operation)
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
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "empty workspace_id returns 400",
			setupFunc: func() string {
				return ""
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID := tc.setupFunc()
			path := "/v1/terraform-workspaces/" + workspaceID + "/lock"
			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(rr.Body.String())
			}
		})
	}
}
