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
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	comphelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// GetAppActionTestService holds all fx-injected dependencies for get app action tests.
type GetAppActionTestService struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	VcsHelpers     *vcshelpers.Helpers
	CompHelpers    *comphelpers.Helpers
	ActionsHelpers *actionshelpers.Helpers
	InstallHelpers *installhelpers.Helpers
	ActionsService *service
	Seeder         *testseed.Seeder
}

// GetAppActionTestSuite is the testify suite for GetAppAction endpoint.
type GetAppActionTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      GetAppActionTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestGetAppActionSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetAppActionTestSuite))
}

func (s *GetAppActionTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create fake event loop client for testing
	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

			CustomValidator: true,
		}),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GetAppActionTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

	// Create test router with standard middlewares using helper
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.ActionsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetAppActionTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetAppActionTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetAppActionTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetAppActionTestSuite) TestGetAppActionSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		lookupBy     string // "id" or "name"
		expectedCode int
		validateFunc func(*app.ActionWorkflow)
	}{
		{
			name: "get action by ID returns 200 with correct data",
			setupFunc: func() string {
				action := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   "test-action-by-id",
					Status: app.ActionWorkflowStatusActive,
				}
				err := s.service.DB.WithContext(s.ctx).Create(action).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
				})

				return action.ID
			},
			lookupBy:     "id",
			expectedCode: http.StatusOK,
			validateFunc: func(action *app.ActionWorkflow) {
				assert.Equal(s.T(), "test-action-by-id", action.Name)
				assert.Equal(s.T(), s.testOrg.ID, action.OrgID)
				assert.Equal(s.T(), s.testApp.ID, action.AppID)
				assert.Equal(s.T(), app.ActionWorkflowStatusActive, action.Status)
			},
		},
		{
			name: "get action by name returns 200",
			setupFunc: func() string {
				action := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   "test-action-by-name",
					Status: app.ActionWorkflowStatusActive,
				}
				err := s.service.DB.WithContext(s.ctx).Create(action).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
				})

				return action.Name
			},
			lookupBy:     "name",
			expectedCode: http.StatusOK,
			validateFunc: func(action *app.ActionWorkflow) {
				assert.Equal(s.T(), "test-action-by-name", action.Name)
				assert.Equal(s.T(), s.testApp.ID, action.AppID)
			},
		},
		{
			name: "get non-existent action returns 404",
			setupFunc: func() string {
				return domains.NewActionWorkflowID()
			},
			lookupBy:     "id",
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actionIdentifier := tc.setupFunc()
			rr := s.makeRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/actions/"+actionIdentifier, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				var response app.ActionWorkflow
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				tc.validateFunc(&response)
			}
		})
	}
}

func (s *GetAppActionTestSuite) TestGetAppActionDifferentOrg() {
	// Create action in different org
	ctx2 := context.Background()
	ctx2, _ = s.service.Seeder.EnsureAccount(ctx2, s.T())
	ctx2, org2 := s.service.Seeder.EnsureOrg(ctx2, s.T())

	otherAction := &app.ActionWorkflow{
		ID:     domains.NewActionWorkflowID(),
		AppID:  s.testApp.ID,
		OrgID:  org2.ID,
		Name:   "other-org-action",
		Status: app.ActionWorkflowStatusActive,
	}
	err := s.service.DB.WithContext(ctx2).Create(otherAction).Error
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", otherAction.ID)
	})

	// Try to get action from different org
	rr := s.makeRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/actions/"+otherAction.ID, nil)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
