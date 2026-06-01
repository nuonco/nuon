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

// UpdateAppActionDeprecatedTestService holds all fx-injected dependencies for update app action tests.
type UpdateAppActionDeprecatedTestService struct {
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

// UpdateAppActionDeprecatedTestSuite is the testify suite for UpdateAppAction endpoint.
type UpdateAppActionDeprecatedTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service UpdateAppActionDeprecatedTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestUpdateAppActionDeprecatedSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(UpdateAppActionDeprecatedTestSuite))
}

func (s *UpdateAppActionDeprecatedTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create fake event loop client for testing

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

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

func (s *UpdateAppActionDeprecatedTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test

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

func (s *UpdateAppActionDeprecatedTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateAppActionDeprecatedTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *UpdateAppActionDeprecatedTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UpdateAppActionDeprecatedTestSuite) TestUpdateAppActionSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		newName      string
		expectedCode int
		validateFunc func(*app.ActionWorkflow)
	}{
		{
			name: "update action name returns 201 with updated data",
			setupFunc: func() string {
				action := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   "original-name",
					Status: app.ActionWorkflowStatusActive,
				}
				err := s.service.DB.WithContext(s.ctx).Create(action).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
				})

				return action.ID
			},
			newName:      "updated-name",
			expectedCode: http.StatusCreated,
			validateFunc: func(action *app.ActionWorkflow) {
				assert.Equal(s.T(), "updated-name", action.Name)
				assert.Equal(s.T(), s.testOrg.ID, action.OrgID)
				assert.Equal(s.T(), s.testApp.ID, action.AppID)
			},
		},
		{
			name: "update action with different name returns 201",
			setupFunc: func() string {
				action := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   "deploy-staging",
					Status: app.ActionWorkflowStatusActive,
				}
				err := s.service.DB.WithContext(s.ctx).Create(action).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
				})

				return action.ID
			},
			newName:      "deploy-production",
			expectedCode: http.StatusCreated,
			validateFunc: func(action *app.ActionWorkflow) {
				assert.Equal(s.T(), "deploy-production", action.Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actionID := tc.setupFunc()

			req := UpdateActionRequest{
				Name: tc.newName,
			}
			rr := s.makeRequest(http.MethodPatch, "/v1/apps/"+s.testApp.ID+"/actions/"+actionID, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var response app.ActionWorkflow
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			if tc.validateFunc != nil {
				tc.validateFunc(&response)
			}

			// Verify action was updated in database
			var dbAction app.ActionWorkflow
			err = s.service.DB.First(&dbAction, "id = ?", actionID).Error
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tc.newName, dbAction.Name)
		})
	}
}

func (s *UpdateAppActionDeprecatedTestSuite) TestUpdateAppActionNonExistent() {
	req := UpdateActionRequest{
		Name: "updated-name",
	}

	nonExistentID := domains.NewActionWorkflowID()
	rr := s.makeRequest(http.MethodPatch, "/v1/apps/"+s.testApp.ID+"/actions/"+nonExistentID, req)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *UpdateAppActionDeprecatedTestSuite) TestUpdateAppActionInvalidRequest() {
	// Create action first
	action := &app.ActionWorkflow{
		ID:     domains.NewActionWorkflowID(),
		AppID:  s.testApp.ID,
		OrgID:  s.testOrg.ID,
		Name:   "test-action",
		Status: app.ActionWorkflowStatusActive,
	}
	err := s.service.DB.WithContext(s.ctx).Create(action).Error
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
	})

	testCases := []struct {
		name         string
		requestBody  interface{}
		expectedCode int
	}{
		{
			name:         "malformed JSON returns 400",
			requestBody:  "invalid json",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var reqBody *bytes.Buffer
			if str, ok := tc.requestBody.(string); ok {
				reqBody = bytes.NewBufferString(str)
			} else {
				jsonBytes, err := json.Marshal(tc.requestBody)
				require.NoError(s.T(), err)
				reqBody = bytes.NewBuffer(jsonBytes)
			}

			req, err := http.NewRequest(http.MethodPatch, "/v1/apps/"+s.testApp.ID+"/actions/"+action.ID, reqBody)
			require.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *UpdateAppActionDeprecatedTestSuite) TestUpdateAppActionDifferentOrg() {
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

	req := UpdateActionRequest{
		Name: "updated-name",
	}

	// Try to update action from different org
	rr := s.makeRequest(http.MethodPatch, "/v1/apps/"+s.testApp.ID+"/actions/"+otherAction.ID, req)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
