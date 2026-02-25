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

type AdminLockWorkspaceTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminLockWorkspaceTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service AdminLockWorkspaceTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestAdminLockWorkspaceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminLockWorkspaceTestSuite))
}

func (s *AdminLockWorkspaceTestSuite) SetupSuite() {
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

func (s *AdminLockWorkspaceTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// CRITICAL: TestAcc needed because handler creates lock records with created_by_id
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminLockWorkspaceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminLockWorkspaceTestSuite) setupTestData() {
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

func (s *AdminLockWorkspaceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminLockWorkspaceTestSuite) TestAdminLockWorkspace() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, interface{})
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully lock workspace by ID",
			setupFunc: func() (string, interface{}) {
				return s.testWS.ID, AdminLockWorkspace{}
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify lock was created in DB with fake data
				var lock app.TerraformWorkspaceLock
				err := s.service.DB.Where("workspace_id = ?", workspaceID).First(&lock).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), workspaceID, lock.WorkspaceID)
				assert.Equal(s.T(), workspaceID, lock.Lock.ID)
			},
		},
		{
			name: "lock workspace by owner_id",
			setupFunc: func() (string, interface{}) {
				// Clean up any existing locks for this workspace from previous subtests
				s.service.DB.Unscoped().Where("workspace_id = ?", s.testWS.ID).Delete(&app.TerraformWorkspaceLock{})

				// Use owner_id instead of workspace_id
				return s.testOrg.ID, AdminLockWorkspace{}
			},
			expectedCode: http.StatusOK,
			validateFunc: func(ownerID string) {
				// Verify lock was created using workspace found by owner_id
				var lock app.TerraformWorkspaceLock
				err := s.service.DB.Where("workspace_id = ?", s.testWS.ID).First(&lock).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testWS.ID, lock.WorkspaceID)
			},
		},
		{
			name: "lock workspace in different org",
			setupFunc: func() (string, interface{}) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

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

				ws2 := &app.TerraformWorkspace{
					ID:        domains.NewTerraformWorkspaceID(),
					OrgID:     org2.ID,
					OwnerID:   org2.ID,
					OwnerType: "org",
				}
				err = s.service.DB.WithContext(ctx).Create(ws2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Where("workspace_id = ?", ws2.ID).Delete(&app.TerraformWorkspaceLock{})
					s.service.DB.Unscoped().Delete(ws2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return ws2.ID, AdminLockWorkspace{}
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Admin routes have no org scoping - can lock any workspace
				var lock app.TerraformWorkspaceLock
				err := s.service.DB.Where("workspace_id = ?", workspaceID).First(&lock).Error
				require.NoError(s.T(), err)
			},
		},
		{
			name: "workspace not found returns error",
			setupFunc: func() (string, interface{}) {
				return "twsnonexistent123456789012", AdminLockWorkspace{}
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
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
