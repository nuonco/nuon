package health

import (
	"encoding/json"
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

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/psql"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/loops"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/metrics"
	signaldb "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
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
	suite.Suite

	app     *fxtest.App
	service TestService
	router  *gin.Engine
}

func TestHealthSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(HealthTestSuite))
}

func (s *HealthTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.app = fxtest.New(
		s.T(),
		fx.Provide(internal.NewTestConfig),

		// logging
		fx.Provide(log.New),
		fx.Provide(dblog.New),

		// external services
		fx.Provide(loops.New),
		fx.Provide(github.New),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),

		// databases
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		// temporal (required for health checks)
		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(signaldb.NewPayloadConverter),
		fx.Provide(dataconverter.New),
		fx.Provide(temporal.New),

		// validator
		fx.Provide(validator.New),

		// service under test
		fx.Provide(New),

		// invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),

		fx.Populate(&s.service),
	)

	s.app.RequireStart()

	// Create test router and register routes
	s.router = gin.New()
	err := s.service.Health.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *HealthTestSuite) TearDownSuite() {
	s.app.RequireStop()
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
