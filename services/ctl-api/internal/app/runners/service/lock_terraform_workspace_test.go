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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type LockTerraformWorkspaceTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type LockTerraformWorkspaceTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service LockTerraformWorkspaceTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestLockTerraformWorkspaceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(LockTerraformWorkspaceTestSuite))
}

func (s *LockTerraformWorkspaceTestSuite) SetupSuite() {
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

func (s *LockTerraformWorkspaceTestSuite) SetupTest() {
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

func (s *LockTerraformWorkspaceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *LockTerraformWorkspaceTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create terraform workspace
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

func (s *LockTerraformWorkspaceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *LockTerraformWorkspaceTestSuite) TestLockTerraformWorkspace() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, interface{})
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully lock workspace",
			setupFunc: func() (string, interface{}) {
				lock := app.TerraformLock{
					ID:        "lock-123",
					Operation: "apply",
					Info:      "terraform apply",
					Who:       "test-user",
					Version:   "1.0.0",
					Created:   "2024-01-01T00:00:00Z",
					Path:      "/tmp/terraform",
				}
				return s.testWS.ID, lock
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify lock was created in DB
				var lock app.TerraformWorkspaceLock
				err := s.service.DB.Where("workspace_id = ?", workspaceID).First(&lock).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), workspaceID, lock.WorkspaceID)
				assert.Equal(s.T(), "lock-123", lock.Lock.ID)
				assert.Equal(s.T(), "apply", lock.Lock.Operation)
			},
		},
		{
			name: "lock workspace with minimal lock data",
			setupFunc: func() (string, interface{}) {
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Where("workspace_id = ?", ws.ID).Delete(&app.TerraformWorkspaceLock{})
					s.service.DB.Unscoped().Delete(ws)
				})

				lock := app.TerraformLock{
					ID:        "lock-456",
					Operation: "plan",
				}
				return ws.ID, lock
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "workspace not found returns error",
			setupFunc: func() (string, interface{}) {
				lock := app.TerraformLock{
					ID:        "lock-789",
					Operation: "apply",
				}
				return "twsnonexistent123456789012", lock
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "workspace in different org not accessible",
			setupFunc: func() (string, interface{}) {
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

				lock := app.TerraformLock{
					ID:        "lock-cross-org",
					Operation: "apply",
				}
				return ws2.ID, lock
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "empty workspace_id returns 400",
			setupFunc: func() (string, interface{}) {
				lock := app.TerraformLock{
					ID:        "lock-empty",
					Operation: "apply",
				}
				return "", lock
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid JSON body returns 400",
			setupFunc: func() (string, interface{}) {
				return s.testWS.ID, "invalid json"
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID, body := tc.setupFunc()
			path := "/v1/terraform-workspaces/" + workspaceID + "/lock"
			rr := s.makeRequest("POST", path, body)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(workspaceID)
			}
		})
	}
}
