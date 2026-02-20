package service

import (
	"bytes"
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

type UpdateTerraformStateJSONTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type UpdateTerraformStateJSONTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service UpdateTerraformStateJSONTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestUpdateTerraformStateJSONSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(UpdateTerraformStateJSONTestSuite))
}

func (s *UpdateTerraformStateJSONTestSuite) SetupSuite() {
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

func (s *UpdateTerraformStateJSONTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner routes - must include TestAcc for created_by_id context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *UpdateTerraformStateJSONTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateTerraformStateJSONTestSuite) setupTestData() {
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

func (s *UpdateTerraformStateJSONTestSuite) makeRequest(method, path string, body []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *UpdateTerraformStateJSONTestSuite) TestUpdateTerraformStateJSON() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, []byte)
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "successfully update state JSON",
			setupFunc: func() (string, []byte) {
				jsonData := []byte(`{"version": 4, "terraform_version": "1.2.0", "serial": 1}`)
				return s.testWS.ID, jsonData
			},
			expectedCode: http.StatusOK,
			validateFunc: func(workspaceID string) {
				// Verify state JSON was created
				var stateJSON app.TerraformWorkspaceStateJSON
				err := s.service.DB.Where("workspace_id = ?", workspaceID).
					Order("created_at DESC").
					First(&stateJSON).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), workspaceID, stateJSON.WorkspaceID)
				assert.NotNil(s.T(), stateJSON.Contents)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&stateJSON)
				})
			},
		},
		{
			name: "update state JSON with job_id",
			setupFunc: func() (string, []byte) {
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
					s.service.DB.Unscoped().Where("workspace_id = ?", ws.ID).Delete(&app.TerraformWorkspaceStateJSON{})
					s.service.DB.Unscoped().Delete(ws)
				})

				jsonData := []byte(`{"version": 4, "resources": []}`)
				return ws.ID, jsonData
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "missing workspace_id returns 400",
			setupFunc: func() (string, []byte) {
				jsonData := []byte(`{"version": 4}`)
				return "", jsonData
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "empty request body still succeeds",
			setupFunc: func() (string, []byte) {
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
					s.service.DB.Unscoped().Where("workspace_id = ?", ws.ID).Delete(&app.TerraformWorkspaceStateJSON{})
					s.service.DB.Unscoped().Delete(ws)
				})

				return ws.ID, []byte{}
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "update state JSON with complex data",
			setupFunc: func() (string, []byte) {
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
					s.service.DB.Unscoped().Where("workspace_id = ?", ws.ID).Delete(&app.TerraformWorkspaceStateJSON{})
					s.service.DB.Unscoped().Delete(ws)
				})

				jsonData := []byte(`{
					"version": 4,
					"terraform_version": "1.3.0",
					"serial": 5,
					"lineage": "test-lineage",
					"resources": [
						{"type": "aws_instance", "name": "example"}
					]
				}`)
				return ws.ID, jsonData
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID, body := tc.setupFunc()
			path := "/v1/terraform-workspaces/" + workspaceID + "/state-json"
			rr := s.makeRequest("POST", path, body)

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
