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

type DeleteAppActionTestService struct {
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

type DeleteAppActionTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service DeleteAppActionTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestDeleteAppActionSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(DeleteAppActionTestSuite))
}

func (s *DeleteAppActionTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *DeleteAppActionTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.ActionsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeleteAppActionTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeleteAppActionTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *DeleteAppActionTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *DeleteAppActionTestSuite) TestDeleteAppActionSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "delete action by ID returns 200 and marks as delete_queued",
			setupFunc: func() string {
				action := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   "action-to-delete",
					Status: app.ActionWorkflowStatusActive,
				}
				err := s.service.DB.WithContext(s.ctx).Create(action).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
				})

				return action.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(actionID string) {
				var dbAction app.ActionWorkflow
				err := s.service.DB.First(&dbAction, "id = ?", actionID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.ActionWorkflowStatusDeleteQueued, dbAction.Status)
				assert.Equal(s.T(), "Delete Queued", dbAction.StatusDescription)
			},
		},
		{
			name: "delete action by name returns 200",
			setupFunc: func() string {
				uniqueName := "action-by-name-delete-" + domains.NewActionWorkflowID()
				action := &app.ActionWorkflow{
					ID:     domains.NewActionWorkflowID(),
					AppID:  s.testApp.ID,
					OrgID:  s.testOrg.ID,
					Name:   uniqueName,
					Status: app.ActionWorkflowStatusActive,
				}
				err := s.service.DB.WithContext(s.ctx).Create(action).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
				})

				return action.Name
			},
			expectedCode: http.StatusOK,
			validateFunc: func(actionName string) {
				var dbAction app.ActionWorkflow
				err := s.service.DB.Where("name = ? AND org_id = ?", actionName, s.testOrg.ID).First(&dbAction).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.ActionWorkflowStatusDeleteQueued, dbAction.Status)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			actionIdentifier := tc.setupFunc()
			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+s.testApp.ID+"/actions/"+actionIdentifier, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var response bool
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)
			assert.True(s.T(), response)

			if tc.validateFunc != nil {
				tc.validateFunc(actionIdentifier)
			}

			queueSignals := tests.GetQueueSignals(s.T(), s.service.DB)
			assert.Len(s.T(), queueSignals, 0, "delete action workflow does not send signals")
		})
	}
}

func (s *DeleteAppActionTestSuite) TestDeleteAppActionNonExistent() {
	nonExistentID := domains.NewActionWorkflowID()
	rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+s.testApp.ID+"/actions/"+nonExistentID, nil)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *DeleteAppActionTestSuite) TestDeleteAppActionDifferentOrg() {
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

	rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+s.testApp.ID+"/actions/"+otherAction.ID, nil)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)

	var dbAction app.ActionWorkflow
	err = s.service.DB.First(&dbAction, "id = ?", otherAction.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), app.ActionWorkflowStatusActive, dbAction.Status)
}

func (s *DeleteAppActionTestSuite) TestDeleteAppActionAlreadyDeleted() {
	action := &app.ActionWorkflow{
		ID:                domains.NewActionWorkflowID(),
		AppID:             s.testApp.ID,
		OrgID:             s.testOrg.ID,
		Name:              "already-deleted-action",
		Status:            app.ActionWorkflowStatusDeleteQueued,
		StatusDescription: "Delete Queued",
	}
	err := s.service.DB.WithContext(s.ctx).Create(action).Error
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
	})

	rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+s.testApp.ID+"/actions/"+action.ID, nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var dbAction app.ActionWorkflow
	err = s.service.DB.First(&dbAction, "id = ?", action.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), app.ActionWorkflowStatusDeleteQueued, dbAction.Status)
}
