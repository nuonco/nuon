package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// TestService holds all fx-injected dependencies for health endpoint tests.
type TestService struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	CHDB   *gorm.DB `name:"ch"`
	V      *validator.Validate
	L      *zap.Logger
	Health *Service
}

// HealthTestSuite is the testify suite for health endpoints.
type HealthTestSuite struct {
	tests.BaseDBTestSuite

	app                *fxtest.App
	service            TestService
	router             *gin.Engine
	mockCtrl           *gomock.Controller
	mockTemporalClient *temporalclient.MockClient
}

func TestHealthSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(HealthTestSuite))
}

func (s *HealthTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create mock controller and temporal client
	s.mockCtrl = gomock.NewController(s.T())
	s.mockTemporalClient = temporalclient.NewMockClient(s.mockCtrl)

	// Set up expectation for CheckHealth - can be called multiple times in tests
	s.mockTemporalClient.EXPECT().
		CheckHealth(gomock.Any(), gomock.Any()).
		Return(&client.CheckHealthResponse{}, nil).
		AnyTimes()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),
			Mocks: &tests.TestMocks{
				MockTC: s.mockTemporalClient,
			},
			CustomValidator: false,
		}),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)

	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)

	// Create test router and register routes
	s.router = gin.New()
	err := s.service.Health.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *HealthTestSuite) TearDownSuite() {
	s.app.RequireStop()
	s.mockCtrl.Finish()
}

func (s *HealthTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *HealthTestSuite) TestLivezReturnsValidResponse() {
	rr := s.makeRequest(http.MethodGet, "/livez")

	// Should return 200 OK or 207 Multi-Status (degraded)
	require.Contains(s.T(), []int{http.StatusOK, http.StatusMultiStatus}, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Response should have status field
	status, ok := response["status"].(string)
	require.True(s.T(), ok, "response should have status field")
	require.Contains(s.T(), []string{"ok", "degraded"}, status)

	// Response should have degraded array
	_, ok = response["degraded"].([]any)
	require.True(s.T(), ok, "response should have degraded field")
}

func (s *HealthTestSuite) TestReadyzReturnsOK() {
	rr := s.makeRequest(http.MethodGet, "/readyz")

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	status, ok := response["status"].(string)
	require.True(s.T(), ok, "response should have status field")
	require.Equal(s.T(), "ok", status)
}

func (s *HealthTestSuite) TestVersionReturnsVersionInfo() {
	rr := s.makeRequest(http.MethodGet, "/version")

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Should have version field
	_, ok := response["version"]
	require.True(s.T(), ok, "response should have version field")

	// Should have git_ref field
	_, ok = response["git_ref"]
	require.True(s.T(), ok, "response should have git_ref field")
}
