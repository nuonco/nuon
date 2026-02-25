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
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// TestService holds all fx-injected dependencies for apps endpoint tests.
type TestService struct {
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
	Seeder          *testseed.Seeder
}

// AppsTestSuite is the testify suite for apps endpoints.
type AppsTestSuite struct {
	tests.BaseDBTestSuite

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
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)

	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *AppsTestSuite) SetupTest() {
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

func (s *AppsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AppsTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
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

	// Use OpenAPI-generated response type
	var response []*models.AppApp
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

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Use OpenAPI-generated response type
	var response []*models.AppApp
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
	}
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

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Use OpenAPI-generated response type
	var response []*models.AppApp
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
	}
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

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Use OpenAPI-generated response type
	var response []*models.AppApp
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
	}
	require.NoError(s.T(), err)
	require.LessOrEqual(s.T(), len(response), 5)
}

func (s *AppsTestSuite) TestGetAppsOnlyReturnsAppsFromCurrentOrg() {
	// Create another org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	otherOrg := &app.Org{
		ID:          domains.NewOrgID(), // BeforeCreate adds "org" prefix
		Name:        "other-org-" + domains.NewOrgID(),
		SandboxMode: true,
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

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Use OpenAPI-generated response type
	var response []*models.AppApp
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
	}
	require.NoError(s.T(), err)
	require.Len(s.T(), response, 1)
	require.Equal(s.T(), "my-app", response[0].Name)
	require.Equal(s.T(), s.testOrg.ID, response[0].OrgID)
}
