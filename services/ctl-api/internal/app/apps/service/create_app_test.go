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

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// CreateAppTestService holds all fx-injected dependencies for create app tests.
type CreateAppTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	MW              metrics.Writer
	VcsHelpers      *vcshelpers.Helpers
	AppsHelpers     *appshelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	AppsService     *service
}

// CreateAppTestSuite is the testify suite for CreateApp endpoint.
type CreateAppTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service CreateAppTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestCreateAppSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CreateAppTestSuite))
}

func (s *CreateAppTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)

	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *CreateAppTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares using helper
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *CreateAppTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateAppTestSuite) setupTestData() {
	// Create test account
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "user@example.com",
		Subject:     "subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:   domains.NewOrgID(),
		Name: "test-org",
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *CreateAppTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *CreateAppTestSuite) TestCreateAppSuccess() {
	req := CreateAppRequest{
		Name:        "test-app",
		Description: "Test app",
	}
	rr := s.makeRequest(http.MethodPost, "/v1/apps", req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Use OpenAPI-generated response type
	var response models.AppApp
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Verify response fields
	assert.NotEmpty(s.T(), response.ID)
	assert.Equal(s.T(), "test-app", response.Name)
	assert.Equal(s.T(), s.testOrg.ID, response.OrgID)

	// Verify app was created in database
	var dbApp app.App
	err = s.service.DB.First(&dbApp, "id = ?", response.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-app", dbApp.Name)
	assert.Equal(s.T(), s.testOrg.ID, dbApp.OrgID)
}

func (s *CreateAppTestSuite) TestCreateAppValidationError() {
	// entity_name validator allows: lowercase letters, numbers, underscores, hyphens
	// regex: ^[a-z0-9_-]*$
	testCases := []struct {
		appName  string
		testName string
	}{
		{appName: "", testName: "empty name"},
		{appName: "my app", testName: "name with spaces"},
		{appName: "MyApp", testName: "name with uppercase"},
		{appName: "my-app!@#", testName: "name with special chars"},
		{appName: "my.app", testName: "name with dots"},
		{appName: "my/app", testName: "name with slashes"},
	}

	for _, tc := range testCases {
		s.Run(tc.testName, func() {
			req := CreateAppRequest{
				Name: tc.appName,
			}
			rr := s.makeRequest(http.MethodPost, "/v1/apps", req)

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

func (s *CreateAppTestSuite) TestCreateAppDuplicateName() {
	appName := "test-app"

	// Create existing app
	existingApp := &app.App{
		Name:        appName,
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
	}
	err := s.service.DB.Create(existingApp).Error
	require.NoError(s.T(), err)

	s.Run("within org", func() {
		// Try to create duplicate app
		req := CreateAppRequest{Name: appName}
		rr := s.makeRequest(http.MethodPost, "/v1/apps", req)

		if rr.Code != http.StatusConflict {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}

		// Validate 409 within org
		require.Equal(s.T(), http.StatusConflict, rr.Code)
	})

	s.Run("across orgs", func() {
		// Create a second org with a different account
		acc2 := &app.Account{
			ID:          domains.NewAccountID(),
			Email:       "test2@example.com",
			Subject:     "subject",
			AccountType: app.AccountTypeAuth0,
		}
		err := s.service.DB.Create(acc2).Error
		require.NoError(s.T(), err)
		defer s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)

		ctx := context.Background()
		ctx = cctx.SetAccountContext(ctx, acc2)
		org2 := &app.Org{
			ID:   domains.NewOrgID(),
			Name: "test-org-2",
			NotificationsConfig: app.NotificationsConfig{
				InternalSlackWebhookURL: "https://hooks.slack.com/foo",
			},
		}
		err = s.service.DB.WithContext(ctx).Create(org2).Error
		require.NoError(s.T(), err)
		defer s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)

		// Recreate router with new org context
		router := tests.NewTestRouter(tests.RouterOptions{
			L:       s.service.L,
			DB:      s.service.DB,
			TestOrg: org2,
			TestAcc: acc2,
		})
		err = s.service.AppsService.RegisterPublicRoutes(router)
		require.NoError(s.T(), err)

		// Try to create duplicate app across orgs
		req := CreateAppRequest{Name: appName}

		var reqBody *bytes.Buffer
		jsonBytes, err := json.Marshal(req)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)

		httpReq, err := http.NewRequest(http.MethodPost, "/v1/apps", reqBody)
		require.NoError(s.T(), err)
		httpReq.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httpReq)

		// Validate 201
		require.Equal(s.T(), http.StatusCreated, rr.Code)
	})
}
