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

type DeleteTerraformStateJSONTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type DeleteTerraformStateJSONTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeleteTerraformStateJSONTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestDeleteTerraformStateJSONSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeleteTerraformStateJSONTestSuite))
}

func (s *DeleteTerraformStateJSONTestSuite) SetupSuite() {
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

func (s *DeleteTerraformStateJSONTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner routes (this endpoint doesn't use getWorkspace validation)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeleteTerraformStateJSONTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeleteTerraformStateJSONTestSuite) setupTestData() {
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

func (s *DeleteTerraformStateJSONTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeleteTerraformStateJSONTestSuite) TestDeleteTerraformStateJSON() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "successfully delete state JSON",
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

				// Create state JSON
				stateJSON := &app.TerraformWorkspaceStateJSON{
					ID:          domains.NewTerraformWorkspaceStateJSONID(),
					WorkspaceID: ws.ID,
					Contents:    []byte(`{"version": 4}`),
				}
				err = s.service.DB.WithContext(ctx).Create(stateJSON).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify state JSON was deleted
				var stateJSON []app.TerraformWorkspaceStateJSON
				err := s.service.DB.Where("workspace_id = ?", workspaceID).Find(&stateJSON).Error
				require.NoError(s.T(), err)
				assert.Empty(s.T(), stateJSON)
			},
		},
		{
			name: "delete state JSON when none exists",
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "delete multiple state JSON entries",
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

				// Create multiple state JSON entries
				for i := 0; i < 3; i++ {
					stateJSON := &app.TerraformWorkspaceStateJSON{
						ID:          domains.NewTerraformWorkspaceStateJSONID(),
						WorkspaceID: ws.ID,
						Contents:    []byte(`{"version": 4, "serial": ` + string(rune(i)) + `}`),
					}
					err = s.service.DB.WithContext(ctx).Create(stateJSON).Error
					require.NoError(s.T(), err)
				}

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify all state JSON entries were deleted
				var stateJSON []app.TerraformWorkspaceStateJSON
				err := s.service.DB.Where("workspace_id = ?", workspaceID).Find(&stateJSON).Error
				require.NoError(s.T(), err)
				assert.Empty(s.T(), stateJSON)
			},
		},
		{
			name: "missing workspace_id returns 400",
			setupFunc: func() string {
				return ""
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "workspace ID with no associated states",
			setupFunc: func() string {
				return s.testWS.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify no state JSON exists
				var stateJSON []app.TerraformWorkspaceStateJSON
				err := s.service.DB.Where("workspace_id = ?", workspaceID).Find(&stateJSON).Error
				require.NoError(s.T(), err)
				assert.Empty(s.T(), stateJSON)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID := tc.setupFunc()
			path := "/v1/terraform-workspaces/" + workspaceID + "/states"
			rr := s.makeRequest("DELETE", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(workspaceID)
			}
		})
	}
}
