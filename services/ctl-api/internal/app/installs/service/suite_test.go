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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// InstallsTestDeps holds all fx-injected dependencies for installs service tests.
type InstallsTestDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	CHDB   *gorm.DB `name:"ch"`
	V      *validator.Validate
	L      *zap.Logger
	MW     metrics.Writer
	Seeder *testseed.Seeder
}

// InstallsServiceTestSuite is the shared testify suite for all installs service endpoint tests.
type InstallsServiceTestSuite struct {
	tests.BaseDBTestSuite

	fxApp           *fxtest.App
	deps            InstallsTestDeps
	installsService *service
	router          *gin.Engine
	ctx             context.Context
	testOrg         *app.Org
	testAcc         *app.Account
	testApp         *app.App
	testAppConfig   *app.AppConfig
	mockEvClient    *tests.MockEventLoopClient
}

func TestInstallsServiceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(InstallsServiceTestSuite))
}

func (s *InstallsServiceTestSuite) SetupSuite() {
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
		fx.Populate(&s.deps, &s.installsService),
	)

	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.deps.DB)
}

func (s *InstallsServiceTestSuite) SetupTest() {
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

	err := s.installsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *InstallsServiceTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *InstallsServiceTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.testAppConfig = s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
}

// createTestInstall seeds an install via the database for read-only endpoint tests.
func (s *InstallsServiceTestSuite) createTestInstall() *app.Install {
	return s.deps.Seeder.CreateInstall(s.ctx, s.T(), s.testApp)
}

// makeRequest sends an HTTP request through the test router and returns the recorder.
// Pass nil for body on requests that have no body (GET, no-body POST).
func (s *InstallsServiceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// makeRawRequest sends a raw string body through the test router, bypassing json.Marshal.
func (s *InstallsServiceTestSuite) makeRawRequest(method, path string, rawBody string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(rawBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

type testInstallWithWorkflow struct {
	Install    *app.Install
	WorkflowID string
}

func (s *InstallsServiceTestSuite) createTestInstallViaAPI() testInstallWithWorkflow {
	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name: fmt.Sprintf("api-install-%d", time.Now().UnixNano()),
			AWSAccount: &struct {
				Region string `json:"region"`
			}{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))

	return testInstallWithWorkflow{
		Install:    &install,
		WorkflowID: rr.Header().Get(app.HeaderInstallWorkflowID),
	}
}

func (s *InstallsServiceTestSuite) getSeededComponent(componentType app.ComponentType) *app.Component {
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
