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

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminNoopJobTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminNoopJobTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       AdminNoopJobTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
}

func TestAdminNoopJobSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminNoopJobTestSuite))
}

func (s *AdminNoopJobTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			CustomValidator: true,
		}),

		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminNoopJobTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminNoopJobTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminNoopJobTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	s.testRunnerGrp = &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testOrg.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testRunnerGrp).Error
	require.NoError(s.T(), err)

	s.testRunner = &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         s.testOrg.ID,
		Name:          "test-runner",
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
		RunnerGroupID: s.testRunnerGrp.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunner).Error
	require.NoError(s.T(), err)
}

func (s *AdminNoopJobTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminNoopJobTestSuite) TestAdminCreateNoopJob() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully create noop job",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				var job app.RunnerJob
				err := s.service.DB.WithContext(ctx).
					Where("runner_id = ? AND type = ?", runnerID, app.RunnerJobTypeNOOP).
					Order("created_at DESC").
					First(&job).Error
				require.NoError(s.T(), err)

				assert.Equal(s.T(), app.RunnerJobStatusQueued, job.Status)
				assert.Equal(s.T(), app.RunnerJobTypeNOOP, job.Type)
				assert.Equal(s.T(), app.RunnerJobGroupOperations, job.Group)
				assert.NotNil(s.T(), job.LogStreamID)

				var logStream app.LogStream
				err = s.service.DB.WithContext(ctx).First(&logStream, "id = ?", *job.LogStreamID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), runnerID, logStream.OwnerID)
				assert.Equal(s.T(), "runners", logStream.OwnerType)
				assert.True(s.T(), logStream.Open)

				signals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), signals, 1)
				assert.Equal(s.T(), runnerID, signals[0].OwnerID)
			},
		},
		{
			name: "nonexistent runner returns error",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "empty body accepted",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				var job app.RunnerJob
				err := s.service.DB.
					Where("runner_id = ? AND type = ?", runnerID, app.RunnerJobTypeNOOP).
					Order("created_at DESC").
					First(&job).Error
				require.NoError(s.T(), err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/runners/"+runnerID+"/noop-job", AdminCreateNoopJobRequest{})

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				tc.validateFunc(runnerID)
			}
		})
	}
}
