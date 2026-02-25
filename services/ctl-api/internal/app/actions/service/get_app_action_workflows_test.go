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

// GetAppActionsTestService holds all fx-injected dependencies for get app actions tests.
type GetAppActionsTestService struct {
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

// GetAppActionsTestSuite is the testify suite for GetAppActions endpoint.
type GetAppActionsTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      GetAppActionsTestService
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	testApp      *app.App
	mockEvClient *tests.MockEventLoopClient
}

func TestGetAppActionsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetAppActionsTestSuite))
}

func (s *GetAppActionsTestSuite) SetupSuite() {
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

func (s *GetAppActionsTestSuite) SetupTest() {
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

func (s *GetAppActionsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetAppActionsTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *GetAppActionsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *GetAppActionsTestSuite) TestGetAppActionsSuccess() {
	testCases := []struct {
		name          string
		setupFunc     func() []string
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]app.ActionWorkflow)
	}{
		{
			name: "get all actions returns 200 with correct count",
			setupFunc: func() []string {
				actionIDs := make([]string, 0)
				for i := 0; i < 3; i++ {
					action := &app.ActionWorkflow{
						ID:     domains.NewActionWorkflowID(),
						AppID:  s.testApp.ID,
						OrgID:  s.testOrg.ID,
						Name:   "action-" + domains.NewActionWorkflowID(),
						Status: app.ActionWorkflowStatusActive,
					}
					err := s.service.DB.WithContext(s.ctx).Create(action).Error
					require.NoError(s.T(), err)
					actionIDs = append(actionIDs, action.ID)
				}
				return actionIDs
			},
			queryParams:   "",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
			validateFunc: func(actions []app.ActionWorkflow) {
				assert.Len(s.T(), actions, 3)
				for _, action := range actions {
					assert.Equal(s.T(), s.testApp.ID, action.AppID)
					assert.Equal(s.T(), s.testOrg.ID, action.OrgID)
				}
			},
		},
		{
			name: "get actions with search query filters correctly",
			setupFunc: func() []string {
				actionIDs := make([]string, 0)
				// Create action with specific name
				action1 := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   "deploy-staging",
					Status: app.ActionWorkflowStatusActive,
				}
				err := s.service.DB.WithContext(s.ctx).Create(action1).Error
				require.NoError(s.T(), err)
				actionIDs = append(actionIDs, action1.ID)

				// Create action with different name
				action2 := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   "test-action",
					Status: app.ActionWorkflowStatusActive,
				}
				err = s.service.DB.WithContext(s.ctx).Create(action2).Error
				require.NoError(s.T(), err)
				actionIDs = append(actionIDs, action2.ID)

				return actionIDs
			},
			queryParams:   "?q=deploy",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(actions []app.ActionWorkflow) {
				assert.Len(s.T(), actions, 1)
				assert.Contains(s.T(), actions[0].Name, "deploy")
			},
		},
		{
			name: "get actions with pagination returns correct subset",
			setupFunc: func() []string {
				actionIDs := make([]string, 0)
				for i := 0; i < 5; i++ {
					action := &app.ActionWorkflow{
						ID:     domains.NewActionWorkflowID(),
						AppID:  s.testApp.ID,
						OrgID:  s.testOrg.ID,
						Name:   "action-" + domains.NewActionWorkflowID(),
						Status: app.ActionWorkflowStatusActive,
					}
					err := s.service.DB.WithContext(s.ctx).Create(action).Error
					require.NoError(s.T(), err)
					actionIDs = append(actionIDs, action.ID)
				}
				return actionIDs
			},
			queryParams:   "?limit=2",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(actions []app.ActionWorkflow) {
				assert.Len(s.T(), actions, 2)
			},
		},
		{
			name: "get actions with no results returns empty array",
			setupFunc: func() []string {
				return []string{}
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
			validateFunc: func(actions []app.ActionWorkflow) {
				assert.Len(s.T(), actions, 0)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actionIDs := tc.setupFunc()

			s.T().Cleanup(func() {
				for _, actionID := range actionIDs {
					s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", actionID)
				}
			})

			rr := s.makeRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/actions"+tc.queryParams, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var response []app.ActionWorkflow
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			assert.Len(s.T(), response, tc.expectedCount)

			if tc.validateFunc != nil {
				tc.validateFunc(response)
			}
		})
	}
}

func (s *GetAppActionsTestSuite) TestGetAppActionsNonExistentApp() {
	rr := s.makeRequest(http.MethodGet, "/v1/apps/non-existent-app/actions", nil)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *GetAppActionsTestSuite) TestGetAppActionsDifferentOrg() {
	// Create app in different org
	ctx2 := context.Background()
	ctx2, _ = s.service.Seeder.EnsureAccount(ctx2, s.T())
	ctx2, _ = s.service.Seeder.EnsureOrg(ctx2, s.T())
	otherApp := s.service.Seeder.CreateApp(ctx2, s.T())

	// Note: Current handler behavior has security issue - findApp uses Or("id = ?", appID)
	// without org_id check. However, findActionWorkflows filters by org_id from context,
	// so it returns empty array (no actions in requesting org for that app).
	rr := s.makeRequest(http.MethodGet, "/v1/apps/"+otherApp.ID+"/actions", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Verify empty array is returned (no actions in requesting org for this app)
	var response []app.ActionWorkflow
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Len(s.T(), response, 0)
}
