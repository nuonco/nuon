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

// GeneralRunnerTestDeps holds all fx-injected dependencies for general runner routes tests.
type GeneralRunnerTestDeps struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	MW             metrics.Writer
	Seeder         *testseed.Seeder
	GeneralService *service
}

// GeneralRunnerTestSuite is the testify suite for general runner routes.
type GeneralRunnerTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service GeneralRunnerTestDeps
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
}

func TestGeneralRunnerTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GeneralRunnerTestSuite))
}

func (s *GeneralRunnerTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GeneralRunnerTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.GeneralService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GeneralRunnerTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GeneralRunnerTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
}

// makeRequest creates an HTTP request and executes it through the test router.
// Returns the response recorder for assertions.
func (s *GeneralRunnerTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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
