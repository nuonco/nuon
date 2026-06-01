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

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	comphelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// ActionLabelsTestService holds all fx-injected dependencies for action labels tests.
type ActionLabelsTestService struct {
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

// ActionLabelsTestSuite is the testify suite for action label endpoints.
type ActionLabelsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service ActionLabelsTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestActionLabelsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(ActionLabelsTestSuite))
}

func (s *ActionLabelsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *ActionLabelsTestSuite) SetupTest() {
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

func (s *ActionLabelsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *ActionLabelsTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
}

func (s *ActionLabelsTestSuite) createTestAction(name string) *app.ActionWorkflow {
	action := &app.ActionWorkflow{
		ID:     domains.NewActionWorkflowID(),
		AppID:  s.testApp.ID,
		OrgID:  s.testOrg.ID,
		Name:   name,
		Status: app.ActionWorkflowStatusActive,
	}
	err := s.service.DB.WithContext(s.ctx).Create(action).Error
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.ActionWorkflow{}, "id = ?", action.ID)
	})

	return action
}

func (s *ActionLabelsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *ActionLabelsTestSuite) makeRawRequest(method, path string, rawBody string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(rawBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// Add action labels tests
// ---------------------------------------------------------------------------

func (s *ActionLabelsTestSuite) TestAddActionLabelsSuccess() {
	s.Run("adds labels to action with no existing labels", func() {
		action := s.createTestAction("test-add-labels")

		reqBody := AddActionLabelsRequest{
			Labels: map[string]string{"env": "prod", "team": "platform"},
		}
		path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ActionWorkflow
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "platform", response.Labels["team"])

		// Verify in DB
		var dbAction app.ActionWorkflow
		err = s.service.DB.First(&dbAction, "id = ?", action.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", dbAction.Labels["env"])
		assert.Equal(s.T(), "platform", dbAction.Labels["team"])
	})

	s.Run("merges labels with existing labels", func() {
		action := s.createTestAction("test-merge-labels")

		// Set initial labels
		err := s.service.DB.WithContext(s.ctx).
			Model(&app.ActionWorkflow{}).
			Where("id = ?", action.ID).
			Update("labels", labels.Labels{"env": "staging"}).Error
		require.NoError(s.T(), err)

		reqBody := AddActionLabelsRequest{
			Labels: map[string]string{"team": "platform"},
		}
		path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ActionWorkflow
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "staging", response.Labels["env"])
		assert.Equal(s.T(), "platform", response.Labels["team"])
	})

	s.Run("overwrites existing key", func() {
		action := s.createTestAction("test-overwrite-labels")

		err := s.service.DB.WithContext(s.ctx).
			Model(&app.ActionWorkflow{}).
			Where("id = ?", action.ID).
			Update("labels", labels.Labels{"env": "staging"}).Error
		require.NoError(s.T(), err)

		reqBody := AddActionLabelsRequest{
			Labels: map[string]string{"env": "prod"},
		}
		path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"
		rr := s.makeRequest(http.MethodPost, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ActionWorkflow
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})
}

func (s *ActionLabelsTestSuite) TestAddActionLabelsValidationErrors() {
	action := s.createTestAction("test-validation")
	path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"

	s.Run("empty body", func() {
		rr := s.makeRequest(http.MethodPost, path, map[string]interface{}{})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})

	s.Run("invalid JSON", func() {
		rr := s.makeRawRequest(http.MethodPost, path, "{invalid json")
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})
}

func (s *ActionLabelsTestSuite) TestAddActionLabelsNotFound() {
	reqBody := AddActionLabelsRequest{
		Labels: map[string]string{"env": "prod"},
	}
	nonExistentID := domains.NewActionWorkflowID()
	path := "/v1/apps/" + s.testApp.ID + "/actions/" + nonExistentID + "/labels"
	rr := s.makeRequest(http.MethodPost, path, reqBody)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

// ---------------------------------------------------------------------------
// Remove action labels tests
// ---------------------------------------------------------------------------

func (s *ActionLabelsTestSuite) TestRemoveActionLabelsSuccess() {
	s.Run("removes specified keys", func() {
		action := s.createTestAction("test-remove-labels")

		err := s.service.DB.WithContext(s.ctx).
			Model(&app.ActionWorkflow{}).
			Where("id = ?", action.ID).
			Update("labels", labels.Labels{"env": "prod", "team": "platform", "region": "us-west-2"}).Error
		require.NoError(s.T(), err)

		reqBody := RemoveActionLabelsRequest{
			Keys: []string{"team"},
		}
		path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ActionWorkflow
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "us-west-2", response.Labels["region"])
		_, hasTeam := response.Labels["team"]
		assert.False(s.T(), hasTeam)

		// Verify in DB
		var dbAction app.ActionWorkflow
		err = s.service.DB.First(&dbAction, "id = ?", action.ID).Error
		require.NoError(s.T(), err)
		_, hasTeam = dbAction.Labels["team"]
		assert.False(s.T(), hasTeam)
	})

	s.Run("removing non-existent key succeeds silently", func() {
		action := s.createTestAction("test-remove-nonexistent")

		err := s.service.DB.WithContext(s.ctx).
			Model(&app.ActionWorkflow{}).
			Where("id = ?", action.ID).
			Update("labels", labels.Labels{"env": "prod"}).Error
		require.NoError(s.T(), err)

		reqBody := RemoveActionLabelsRequest{
			Keys: []string{"nonexistent"},
		}
		path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ActionWorkflow
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})

	s.Run("removes all labels", func() {
		action := s.createTestAction("test-remove-all")

		err := s.service.DB.WithContext(s.ctx).
			Model(&app.ActionWorkflow{}).
			Where("id = ?", action.ID).
			Update("labels", labels.Labels{"a": "1", "b": "2"}).Error
		require.NoError(s.T(), err)

		reqBody := RemoveActionLabelsRequest{
			Keys: []string{"a", "b"},
		}
		path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"
		rr := s.makeRequest(http.MethodDelete, path, reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ActionWorkflow
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Empty(s.T(), response.Labels)
	})
}

func (s *ActionLabelsTestSuite) TestRemoveActionLabelsValidationErrors() {
	action := s.createTestAction("test-remove-validation")
	path := "/v1/apps/" + s.testApp.ID + "/actions/" + action.ID + "/labels"

	s.Run("empty body", func() {
		rr := s.makeRequest(http.MethodDelete, path, map[string]interface{}{})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})

	s.Run("invalid JSON", func() {
		rr := s.makeRawRequest(http.MethodDelete, path, "{invalid json")
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})
}

func (s *ActionLabelsTestSuite) TestRemoveActionLabelsNotFound() {
	reqBody := RemoveActionLabelsRequest{
		Keys: []string{"env"},
	}
	nonExistentID := domains.NewActionWorkflowID()
	path := "/v1/apps/" + s.testApp.ID + "/actions/" + nonExistentID + "/labels"
	rr := s.makeRequest(http.MethodDelete, path, reqBody)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
