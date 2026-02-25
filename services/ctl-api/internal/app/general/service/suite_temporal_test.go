package service

import (
	"bytes"
	"context"
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
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// GeneralTemporalTestDeps holds all fx-injected dependencies for general temporal tests.
type GeneralTemporalTestDeps struct {
	fx.In

	DB                *gorm.DB `name:"psql"`
	CHDB              *gorm.DB `name:"ch"`
	V                 *validator.Validate
	L                 *zap.Logger
	MW                metrics.Writer
	Cfg               *internal.Config
	AuthzClient       *authz.Client
	AcctClient        *account.Client
	EvClient          eventloop.Client
	GzipCodec         converter.PayloadCodec `name:"gzip"`
	LargePayloadCodec converter.PayloadCodec `name:"largepayload"`
	Seeder            *testseed.Seeder
}

// GeneralTemporalTestSuite is the test suite for temporal-dependent general service endpoints.
type GeneralTemporalTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	deps    GeneralTemporalTestDeps
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	ctrl    *gomock.Controller
	mockTC  *temporal.MockClient
}

func TestGeneralTemporalTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GeneralTemporalTestSuite))
}

func (s *GeneralTemporalTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create gomock controller and mock temporal client
	s.ctrl = gomock.NewController(s.T())
	s.mockTC = temporal.NewMockClient(s.ctrl)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),
			Mocks: &tests.TestMocks{
				MockTC: s.mockTC,
				MockEv: tests.NewMockEventLoopClient(),
			},
			CustomValidator: true,
		}),
		fx.Populate(&s.deps),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.deps.DB)
}

func (s *GeneralTemporalTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Manually create the service with the mock temporal client
	svc := &service{
		l:              s.deps.L,
		v:              s.deps.V,
		db:             s.deps.DB,
		mw:             s.deps.MW,
		cfg:            s.deps.Cfg,
		temporalClient: s.mockTC,
		authzClient:    s.deps.AuthzClient,
		acctClient:     s.deps.AcctClient,
		evClient:       s.deps.EvClient,
		codecs: []converter.PayloadCodec{
			s.deps.GzipCodec,
			s.deps.LargePayloadCodec,
		},
	}

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := svc.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GeneralTemporalTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GeneralTemporalTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GeneralTemporalTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())
}

// makeRequest sends an HTTP request through the test router and returns the recorder.
// Pass nil for body on requests that have no body (GET, no-body POST).
func (s *GeneralTemporalTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// parseResponse unmarshals the response body into the provided interface.
func parseResponse(rr *httptest.ResponseRecorder, v interface{}) error {
	return json.Unmarshal(rr.Body.Bytes(), v)
}
