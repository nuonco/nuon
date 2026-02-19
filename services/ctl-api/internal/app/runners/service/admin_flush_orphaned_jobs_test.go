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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminFlushOrphanedJobsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminFlushOrphanedJobsTestSuite struct {
	tests.BaseDBTestSuite
	app          *fxtest.App
	service      AdminFlushOrphanedJobsTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.FakeEventLoopClient
}

func TestAdminFlushOrphanedJobsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminFlushOrphanedJobsTestSuite))
}

func (s *AdminFlushOrphanedJobsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewFakeEventLoopClient()
	options := append(
		tests.CtlApiFXOptions(),
		fx.Decorate(func() eventloop.Client {
			return s.mockEvClient
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminFlushOrphanedJobsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.mockEvClient.Reset()
	s.setupTestData()

	// Admin routes do NOT use TestOrg/TestAcc context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminFlushOrphanedJobsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminFlushOrphanedJobsTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *AdminFlushOrphanedJobsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(s.T(), err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminFlushOrphanedJobsTestSuite) TestAdminFlushOrphanedJobs() {
	testCases := []struct {
		name           string
		runnerID       string
		requestBody    interface{}
		expectedCode   int
		expectedSignal bool
		validateFunc   func(string)
	}{
		{
			name:           "flush orphaned jobs with valid runner ID",
			runnerID:       "rnr123456789012345678901",
			requestBody:    AdminFlushOrphanedJobsRequest{},
			expectedCode:   http.StatusCreated,
			expectedSignal: true,
			validateFunc: func(runnerID string) {
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), runnerID, capturedSignals[0].ID)

				sig, ok := capturedSignals[0].Signal.(*signals.Signal)
				require.True(s.T(), ok)
				assert.Equal(s.T(), signals.OperationFlushOrphanedJobs, sig.Type)
			},
		},
		{
			name:           "flush orphaned jobs with empty body",
			runnerID:       "rnr987654321098765432109",
			requestBody:    nil,
			expectedCode:   http.StatusCreated,
			expectedSignal: true,
			validateFunc: func(runnerID string) {
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), runnerID, capturedSignals[0].ID)
			},
		},
		{
			name:           "flush orphaned jobs with arbitrary runner ID",
			runnerID:       "nonexistent-runner-id-abc",
			requestBody:    AdminFlushOrphanedJobsRequest{},
			expectedCode:   http.StatusCreated,
			expectedSignal: true,
			validateFunc: func(runnerID string) {
				// Handler sends signal directly without validation
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), runnerID, capturedSignals[0].ID)
			},
		},
		{
			name:           "flush orphaned jobs with empty runner ID",
			runnerID:       "",
			requestBody:    AdminFlushOrphanedJobsRequest{},
			expectedCode:   http.StatusCreated,
			expectedSignal: true,
			validateFunc: func(runnerID string) {
				// Even with empty ID, signal is sent
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), "", capturedSignals[0].ID)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()
			rr := s.makeRequest("POST", "/v1/runners/"+tc.runnerID+"/flush-orphaned-jobs", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(tc.runnerID)
			}

			// Verify signal presence
			capturedSignals := s.mockEvClient.GetSignals()
			if tc.expectedSignal {
				assert.Len(s.T(), capturedSignals, 1, "expected signal to be sent")
			} else {
				assert.Len(s.T(), capturedSignals, 0, "expected no signal to be sent")
			}
		})
	}
}
