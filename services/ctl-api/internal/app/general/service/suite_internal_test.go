package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// GeneralInternalTestDeps holds all fx-injected dependencies for general internal routes tests.
type GeneralInternalTestDeps struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	MW             metrics.Writer
	Seeder         *testseed.Seeder
	GeneralService *service
}

// GeneralInternalTestSuite is the testify suite for general internal routes.
type GeneralInternalTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      GeneralInternalTestDeps
	router       *gin.Engine
	ctx          context.Context
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
}

func TestGeneralInternalTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GeneralInternalTestSuite))
}

func (s *GeneralInternalTestSuite) SetupSuite() {
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
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GeneralInternalTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.GeneralService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GeneralInternalTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GeneralInternalTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
}

// makeRequest creates an HTTP request and executes it through the test router.
// Returns the response recorder for assertions.
func (s *GeneralInternalTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}
