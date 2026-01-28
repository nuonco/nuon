package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/pagination"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"

	ghclient "github.com/google/go-github/v50/github"
	enumsv1 "go.temporal.io/api/enums/v1"
)

// Test helper functions for mocking external dependencies

// newMockGitHubClient returns a stub GitHub client for testing.
// This client is not connected to any real GitHub instance.
func newMockGitHubClient() (*ghclient.Client, error) {
	return ghclient.NewClient(nil), nil
}

// mockEventLoopClient is a no-op implementation of eventloop.Client for testing.
type mockEventLoopClient struct{}

func (m *mockEventLoopClient) Send(ctx context.Context, id string, signal eventloop.Signal) {}

func (m *mockEventLoopClient) Cancel(ctx context.Context, namespace, id string) error {
	return nil
}

func (m *mockEventLoopClient) GetWorkflowStatus(ctx context.Context, namespace string, workflowID string) (enumsv1.WorkflowExecutionStatus, error) {
	return enumsv1.WORKFLOW_EXECUTION_STATUS_RUNNING, nil
}
func (m *mockEventLoopClient) GetWorkflowCount(ctx context.Context, namespace string, workflowID string) (int64, error) {
	return 0, nil
}

func newMockEventLoopClient() eventloop.Client {
	return &mockEventLoopClient{}
}

// TestService holds all fx-injected dependencies for apps endpoint tests.
type TestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	VcsHelpers      *vcshelpers.Helpers
	AppsHelpers     *appshelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	AppsService     *service
}

// AppsTestSuite is the testify suite for apps endpoints.
type AppsTestSuite struct {
	suite.Suite

	app     *fxtest.App
	service TestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestAppsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AppsTestSuite))
}

func (s *AppsTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.app = fxtest.New(
		s.T(),
		fx.Provide(internal.NewConfig),

		// logging
		fx.Provide(log.New),
		fx.Provide(dblog.New),

		// external services
		fx.Provide(loops.New),
		fx.Provide(newMockGitHubClient), // Use mock GitHub client for tests
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(newMockEventLoopClient), // Use mock eventloop client for tests

		// databases
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		// validator
		fx.Provide(validator.New),

		// clients and dependencies for account client
		fx.Provide(authz.New),
		fx.Provide(analytics.New),
		fx.Provide(account.New),

		// helpers (order matters due to dependencies)
		fx.Provide(accountshelpers.New),
		fx.Provide(vcshelpers.New),
		fx.Provide(actionshelpers.New),
		fx.Provide(componentshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(runnershelpers.New),
		fx.Provide(installshelpers.New),

		// endpoint audit
		fx.Provide(api.NewEndpointAudit),

		// service under test
		fx.Provide(New),

		// invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),

		fx.Populate(&s.service),
	)

	s.app.RequireStart()

	// Create test org and account
	s.setupTestData()

	// Create test router and register routes
	s.router = gin.New()

	// Add pagination middleware to parse query parameters
	paginationMW := pagination.New(pagination.Params{
		L:  s.service.L,
		DB: s.service.DB,
	})
	s.router.Use(paginationMW.Handler())

	// Add test middleware to inject org and account context
	s.router.Use(func(c *gin.Context) {
		if s.testOrg != nil {
			cctx.SetOrgGinContext(c, s.testOrg)
		}
		if s.testAcc != nil {
			cctx.SetAccountGinContext(c, s.testAcc)
		}
		c.Next()
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AppsTestSuite) TearDownSuite() {
	s.cleanupTestData()
	s.app.RequireStop()
}

func (s *AppsTestSuite) setupTestData() {
	// Clean up any existing test data first
	s.service.DB.Unscoped().Where("email = ?", "test@example.com").Delete(&app.Account{})
	s.service.DB.Unscoped().Where("name LIKE ?", "test-org-%").Delete(&app.Org{})

	// Create test account
	testAcc := &app.Account{
		ID:      "acc" + domains.NewAccountID(), // Explicitly set ID
		Email:   "test@example.com",
		Subject: "test-subject",
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), testAcc.ID, "Account ID should be set after creation")
	s.testAcc = testAcc

	// Create test org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:   domains.NewOrgID(), // ID will get prefix from BeforeCreate
		Name: "test-org-" + domains.NewOrgID(),
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *AppsTestSuite) cleanupTestData() {
	if s.testOrg != nil {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", s.testOrg.ID)
	}
	if s.testAcc != nil {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", s.testAcc.ID)
	}
}

func (s *AppsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AppsTestSuite) TestGetAppsReturnsEmptyArrayWhenNoApps() {
	rr := s.makeRequest(http.MethodGet, "/v1/apps")

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []app.App
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), response)
	require.Len(s.T(), response, 0)
}

func (s *AppsTestSuite) TestGetAppsReturnsCreatedApps() {
	// Create test apps (BeforeCreate hook adds "app" prefix automatically)
	app1 := &app.App{
		ID:          domains.NewAppID(),
		Name:        "test-app-1",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	app2 := &app.App{
		ID:          domains.NewAppID(),
		Name:        "test-app-2",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}

	err := s.service.DB.Create(app1).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app1.ID)

	err = s.service.DB.Create(app2).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app2.ID)

	rr := s.makeRequest(http.MethodGet, "/v1/apps")

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []app.App
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	require.Len(s.T(), response, 2)

	// Verify apps are returned in alphabetical order by name
	require.Equal(s.T(), "test-app-1", response[0].Name)
	require.Equal(s.T(), "test-app-2", response[1].Name)
}

func (s *AppsTestSuite) TestGetAppsFiltersWithSearchQuery() {
	// Create test apps with different names (BeforeCreate hook adds "app" prefix automatically)
	app1 := &app.App{
		ID:          domains.NewAppID(),
		Name:        "frontend-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	app2 := &app.App{
		ID:          domains.NewAppID(),
		Name:        "backend-service",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}

	err := s.service.DB.Create(app1).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app1.ID)

	err = s.service.DB.Create(app2).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app2.ID)

	// Search for "frontend"
	rr := s.makeRequest(http.MethodGet, "/v1/apps?q=frontend")

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []app.App
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	require.Len(s.T(), response, 1)
	require.Equal(s.T(), "frontend-app", response[0].Name)
}

func (s *AppsTestSuite) TestGetAppsRespectsPagination() {
	// Create multiple test apps (BeforeCreate hook adds "app" prefix automatically)
	for i := 0; i < 15; i++ {
		testApp := &app.App{
			ID:          domains.NewAppID(),
			Name:        fmt.Sprintf("test-app-%02d", i),
			OrgID:       s.testOrg.ID,
			CreatedByID: s.testAcc.ID,
			Status:      app.AppStatusProvisioning,
		}
		err := s.service.DB.Create(testApp).Error
		require.NoError(s.T(), err)
		defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
	}

	// Request with limit
	rr := s.makeRequest(http.MethodGet, "/v1/apps?limit=5")

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []app.App
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	require.LessOrEqual(s.T(), len(response), 5)
}

func (s *AppsTestSuite) TestGetAppsOnlyReturnsAppsFromCurrentOrg() {
	// Create another org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	otherOrg := &app.Org{
		ID:   domains.NewOrgID(), // BeforeCreate adds "org" prefix
		Name: "other-org-" + domains.NewOrgID(),
	}
	err := s.service.DB.WithContext(ctx).Create(otherOrg).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrg.ID)

	// Create app in test org (BeforeCreate hook adds "app" prefix automatically)
	app1 := &app.App{
		ID:          domains.NewAppID(),
		Name:        "my-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	err = s.service.DB.Create(app1).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app1.ID)

	// Create app in other org (BeforeCreate hook adds "app" prefix automatically)
	app2 := &app.App{
		ID:          domains.NewAppID(),
		Name:        "other-app",
		OrgID:       otherOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	err = s.service.DB.Create(app2).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app2.ID)

	rr := s.makeRequest(http.MethodGet, "/v1/apps")

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []app.App
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	require.Len(s.T(), response, 1)
	require.Equal(s.T(), "my-app", response[0].Name)
	require.Equal(s.T(), s.testOrg.ID, response[0].OrgID)
}
