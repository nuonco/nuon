package service

import (
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
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
)

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
		fx.Provide(github.New),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(eventloop.NewMockClient),

		// databases
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		// validator
		fx.Provide(validator.New),

		// helpers
		fx.Provide(vcshelpers.New),
		fx.Provide(appshelpers.New),
		fx.Provide(installshelpers.New),
		fx.Provide(accountshelpers.New),

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

	// Add test middleware to inject org and account context
	s.router.Use(func(c *gin.Context) {
		ctx := c.Request.Context()
		if s.testOrg != nil {
			ctx = cctx.SetOrgContext(ctx, s.testOrg)
		}
		if s.testAcc != nil {
			ctx = cctx.SetAccountContext(ctx, s.testAcc)
		}
		c.Request = c.Request.WithContext(ctx)
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
	// Create test account
	testAcc := &app.Account{
		Email:   "test@example.com",
		Subject: "test-subject",
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org
	testOrg := &app.Org{
		Name:        "test-org-" + domains.NewOrgID(),
		CreatedByID: testAcc.ID,
	}
	err = s.service.DB.Create(testOrg).Error
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

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []app.App
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), response)
	require.Len(s.T(), response, 0)
}

func (s *AppsTestSuite) TestGetAppsReturnsCreatedApps() {
	// Create test apps
	app1 := &app.App{
		ID:          "app" + domains.NewAppID(),
		Name:        "test-app-1",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	app2 := &app.App{
		ID:          "app" + domains.NewAppID(),
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
	// Create test apps with different names
	app1 := &app.App{
		ID:          "app" + domains.NewAppID(),
		Name:        "frontend-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	app2 := &app.App{
		ID:          "app" + domains.NewAppID(),
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
	// Create multiple test apps
	for i := 0; i < 15; i++ {
		testApp := &app.App{
			ID:          "app" + domains.NewAppID(),
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
	// Create another org
	otherOrg := &app.Org{
		ID:          "org" + domains.NewAppID(),
		Name:        "other-org-" + domains.NewAppID(),
		CreatedByID: s.testAcc.ID,
	}
	err := s.service.DB.Create(otherOrg).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrg.ID)

	// Create app in test org
	app1 := &app.App{
		ID:          "app" + domains.NewAppID(),
		Name:        "my-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	err = s.service.DB.Create(app1).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app1.ID)

	// Create app in other org
	app2 := &app.App{
		ID:          "app" + domains.NewAppID(),
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
