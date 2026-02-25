package service

import (
	"bytes"
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

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// ReleasesTestDeps holds all fx-injected dependencies for releases service tests.
type ReleasesTestDeps struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	MW          metrics.Writer
	Cfg         *internal.Config
	CompHelpers *componenthelpers.Helpers
	Seeder      *testseed.Seeder
}

// ReleasesServiceTestSuite is the shared testify suite for all releases service endpoint tests.
type ReleasesServiceTestSuite struct {
	tests.BaseDBTestSuite

	fxApp           *fxtest.App
	deps            ReleasesTestDeps
	releasesService *service
	router          *gin.Engine
	ctx             context.Context
	testOrg         *app.Org
	testAcc         *app.Account
	testApp         *app.App
	testAppConfig   *app.AppConfig
	mockEvClient    *tests.MockEventLoopClient
}

func TestReleasesServiceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(ReleasesServiceTestSuite))
}

func (s *ReleasesServiceTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create fake event loop client for testing
	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			Mocks:           &tests.TestMocks{MockEv: s.mockEvClient},
			CustomValidator: true,
		}),
		// Service under test
		fx.Provide(New),
		fx.Populate(&s.deps, &s.releasesService),
	)

	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.deps.DB)
}

func (s *ReleasesServiceTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	// Reset mock before each test
	s.mockEvClient.Reset()

	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	// Register both public and internal routes so all tests can run
	err := s.releasesService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)

	err = s.releasesService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *ReleasesServiceTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *ReleasesServiceTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.testAppConfig = s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
}

// makeRequest sends an HTTP request through the test router and returns the recorder.
// Pass nil for body on requests that have no body (GET, no-body POST).
func (s *ReleasesServiceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// getSeededComponent returns the first seeded component of the given type from testAppConfig.
func (s *ReleasesServiceTestSuite) getSeededComponent(componentType app.ComponentType) *app.Component {
	for _, ccc := range s.testAppConfig.ComponentConfigConnections {
		var cmp app.Component
		res := s.deps.DB.First(&cmp, "id = ?", ccc.ComponentID)
		if res.Error == nil && cmp.Type == componentType {
			return &cmp
		}
	}
	s.T().Fatalf("no seeded component of type %s", componentType)
	return nil
}

// getSeededConfigConnection returns the ComponentConfigConnection for the given component ID.
func (s *ReleasesServiceTestSuite) getSeededConfigConnection(componentID string) *app.ComponentConfigConnection {
	for _, ccc := range s.testAppConfig.ComponentConfigConnections {
		if ccc.ComponentID == componentID {
			return &ccc
		}
	}
	s.T().Fatalf("no seeded config connection for component %s", componentID)
	return nil
}

// createInstallForApp creates an install scoped to the test app (needed for release creation).
func (s *ReleasesServiceTestSuite) createInstallForApp() *app.Install {
	installID := domains.NewInstallID()
	name := fmt.Sprintf("test-install-%s", generics.GetFakeObj[string]())
	res := s.deps.DB.WithContext(s.ctx).Exec(
		`INSERT INTO installs (id, name, org_id, created_by_id, app_id, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, NOW(), NOW(), 0)`,
		installID, name, s.testOrg.ID, s.testAcc.ID, s.testApp.ID,
	)
	require.NoError(s.T(), res.Error, "failed to create install: %v", res.Error)
	return &app.Install{
		ID:          installID,
		Name:        name,
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		AppID:       s.testApp.ID,
	}
}
