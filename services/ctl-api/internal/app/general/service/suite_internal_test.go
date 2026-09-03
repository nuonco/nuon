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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	generalhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/general/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type generalWorkflowRun struct{}

func (*generalWorkflowRun) GetID() string                          { return "workflow-id" }
func (*generalWorkflowRun) GetRunID() string                       { return "run-id" }
func (*generalWorkflowRun) Get(context.Context, interface{}) error { return nil }
func (*generalWorkflowRun) GetWithOptions(context.Context, interface{}, tclient.WorkflowRunGetOptions) error {
	return nil
}

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

	app     *fxtest.App
	service GeneralInternalTestDeps
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	mockTC  *temporal.MockClient
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
	s.mockTC = temporal.NewMockClient(gomock.NewController(s.T()))

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			Mocks:           &tests.TestMocks{MockTC: s.mockTC},
			CustomValidator: true,
		}),
		fx.Provide(generalhelpers.New),
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
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), "general", gomock.Any(), "Queue", gomock.Any(),
	).Return(&generalWorkflowRun{}, nil).AnyTimes()
	s.setupTestData()

	// Reset mock before each test

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
