package service

import (
	"bytes"
	"context"
	"encoding/json"
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

type UnlockTerraformWorkspaceTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type UnlockTerraformWorkspaceTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service UnlockTerraformWorkspaceTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestUnlockTerraformWorkspaceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(UnlockTerraformWorkspaceTestSuite))
}

func (s *UnlockTerraformWorkspaceTestSuite) SetupSuite() {
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

func (s *UnlockTerraformWorkspaceTestSuite) SetupTest() {
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

func (s *UnlockTerraformWorkspaceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UnlockTerraformWorkspaceTestSuite) setupTestData() {
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

func (s *UnlockTerraformWorkspaceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UnlockTerraformWorkspaceTestSuite) TestUnlockTerraformWorkspace() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, interface{})
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully unlock workspace",
			setupFunc: func() (string, interface{}) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create workspace with lock
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
					},
				}
				err = s.service.DB.WithContext(ctx).Create(lock).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(lock)
					s.service.DB.Unscoped().Delete(ws)
				})

				unlockData := app.TerraformLock{
					ID: "lock-123",
				}
				return ws.ID, unlockData
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify lock was removed from DB
				var lock app.TerraformWorkspaceLock
				err := s.service.DB.Where("workspace_id = ?", workspaceID).First(&lock).Error
				assert.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
			},
		},
		{
			name: "unlock workspace without existing lock",
			setupFunc: func() (string, interface{}) {
				unlockData := app.TerraformLock{
					ID: "nonexistent-lock",
				}
				return s.testWS.ID, unlockData
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "workspace not found returns error",
			setupFunc: func() (string, interface{}) {
				unlockData := app.TerraformLock{
					ID: "lock-456",
				}
				return "twsnonexistent123456789012", unlockData
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

				unlockData := app.TerraformLock{
					ID: "lock-cross-org",
				}
				return ws2.ID, unlockData
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "empty workspace_id returns 400",
			setupFunc: func() (string, interface{}) {
				unlockData := app.TerraformLock{
					ID: "lock-empty",
				}
				return "", unlockData
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
			path := "/v1/terraform-workspaces/" + workspaceID + "/unlock"
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
