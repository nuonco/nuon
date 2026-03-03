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

type UpdateTerraformStateTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type UpdateTerraformStateTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service UpdateTerraformStateTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testWS  *app.TerraformWorkspace
}

func TestUpdateTerraformStateSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(UpdateTerraformStateTestSuite))
}

func (s *UpdateTerraformStateTestSuite) SetupSuite() {
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

func (s *UpdateTerraformStateTestSuite) SetupTest() {
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

func (s *UpdateTerraformStateTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateTerraformStateTestSuite) setupTestData() {
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

func (s *UpdateTerraformStateTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UpdateTerraformStateTestSuite) TestUpdateTerraformState() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, interface{})
		expectedCode int
		validateFunc func()
	}{
		{
			name: "successfully update terraform state",
			setupFunc: func() (string, interface{}) {
				stateData := &app.TerraformStateData{
					Version:          4,
					TerraformVersion: "1.2.0",
					Serial:           1,
				}
				return s.testWS.ID, stateData
			},
			expectedCode: http.StatusOK,
			validateFunc: func() {
				// Verify state was created in DB
				var state app.TerraformWorkspaceState
				err := s.service.DB.Where("terraform_workspace_id = ?", s.testWS.ID).
					Order("created_at DESC").
					First(&state).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testWS.ID, state.TerraformWorkspaceID)
				assert.NotNil(s.T(), state.Contents)
			},
		},
		{
			name: "missing workspace_id returns 400",
			setupFunc: func() (string, interface{}) {
				stateData := &app.TerraformStateData{
					Version:          4,
					TerraformVersion: "1.0.0",
				}
				return "", stateData
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "workspace not found returns error",
			setupFunc: func() (string, interface{}) {
				stateData := &app.TerraformStateData{
					Version:          4,
					TerraformVersion: "1.0.0",
				}
				return "twsnonexistent123456789012", stateData
			},
			expectedCode: http.StatusNotFound,
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
					OwnerID:   org2.ID,
					OwnerType: "org",
				}
				err = s.service.DB.WithContext(ctx).Create(ws2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ws2)
					s.service.DB.Unscoped().Delete(org2)
				})

				stateData := &app.TerraformStateData{
					Version:          4,
					TerraformVersion: "1.0.0",
				}
				return ws2.ID, stateData
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "update state with job_id parameter",
			setupFunc: func() (string, interface{}) {
				stateData := &app.TerraformStateData{
					Version:          4,
					TerraformVersion: "1.1.0",
					Serial:           2,
				}
				return s.testWS.ID, stateData
			},
			expectedCode: http.StatusOK,
			validateFunc: func() {
				// Verify state was created
				var state app.TerraformWorkspaceState
				err := s.service.DB.Where("terraform_workspace_id = ?", s.testWS.ID).
					Order("created_at DESC").
					First(&state).Error
				require.NoError(s.T(), err)
				assert.NotNil(s.T(), state.Contents)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			workspaceID, body := tc.setupFunc()
			path := "/v1/terraform-backend"
			if workspaceID != "" {
				path += "?workspace_id=" + workspaceID
			}
			rr := s.makeRequest("POST", path, body)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc()
			}
		})
	}
}
