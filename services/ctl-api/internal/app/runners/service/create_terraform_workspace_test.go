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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type CreateTerraformWorkspaceTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type CreateTerraformWorkspaceTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service CreateTerraformWorkspaceTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestCreateTerraformWorkspaceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(CreateTerraformWorkspaceTestSuite))
}

func (s *CreateTerraformWorkspaceTestSuite) SetupSuite() {
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

func (s *CreateTerraformWorkspaceTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes (V2 endpoint needs org context)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *CreateTerraformWorkspaceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateTerraformWorkspaceTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *CreateTerraformWorkspaceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *CreateTerraformWorkspaceTestSuite) TestCreateTerraformWorkspace() {
	testCases := []struct {
		name         string
		reqBody      interface{}
		expectedCode int
		validateFunc func(*app.TerraformWorkspace)
	}{
		{
			name: "successfully create workspace",
			reqBody: CreateTerraformWorkspaceRequest{
				OwnerID:   domains.NewAppID(),
				OwnerType: "app",
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(ws *app.TerraformWorkspace) {
				assert.NotEmpty(s.T(), ws.ID)
				assert.Equal(s.T(), s.testOrg.ID, ws.OrgID)
				assert.Equal(s.T(), s.testAcc.ID, ws.CreatedByID)
				assert.Equal(s.T(), "app", ws.OwnerType)

				// Verify in database
				var dbWS app.TerraformWorkspace
				err := s.service.DB.Where("id = ?", ws.ID).First(&dbWS).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), ws.ID, dbWS.ID)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&dbWS)
				})
			},
		},
		{
			name: "create workspace with org owner type",
			reqBody: CreateTerraformWorkspaceRequest{
				OwnerID:   s.testOrg.ID,
				OwnerType: "org",
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(ws *app.TerraformWorkspace) {
				assert.Equal(s.T(), "org", ws.OwnerType)
				assert.Equal(s.T(), s.testOrg.ID, ws.OwnerID)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Where("id = ?", ws.ID).Delete(&app.TerraformWorkspace{})
				})
			},
		},
		{
			name: "create workspace with install owner type",
			reqBody: CreateTerraformWorkspaceRequest{
				OwnerID:   domains.NewInstallID(),
				OwnerType: "install",
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(ws *app.TerraformWorkspace) {
				assert.Equal(s.T(), "install", ws.OwnerType)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Where("id = ?", ws.ID).Delete(&app.TerraformWorkspace{})
				})
			},
		},
		{
			name:         "missing owner_id returns 400",
			reqBody:      CreateTerraformWorkspaceRequest{OwnerType: "app"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing owner_type returns 400",
			reqBody:      CreateTerraformWorkspaceRequest{OwnerID: domains.NewAppID()},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty request body returns 400",
			reqBody:      CreateTerraformWorkspaceRequest{},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid JSON returns 400",
			reqBody:      "invalid json",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest("POST", "/v1/terraform-workspaces", tc.reqBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil && rr.Code == http.StatusCreated {
				var workspace app.TerraformWorkspace
				err := json.Unmarshal(rr.Body.Bytes(), &workspace)
				require.NoError(s.T(), err)
				tc.validateFunc(&workspace)
			}
		})
	}
}
