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

type AdminDeprovisionTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminDeprovisionTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       AdminDeprovisionTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
}

func TestAdminDeprovisionSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminDeprovisionTestSuite))
}

func (s *AdminDeprovisionTestSuite) SetupSuite() {
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

func (s *AdminDeprovisionTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with internal routes (no org context for admin routes)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminDeprovisionTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminDeprovisionTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create runner group
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

	// Create runner
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

func (s *AdminDeprovisionTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminDeprovisionTestSuite) TestAdminDeprovisionRunner() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		shouldSendSignal bool
		expectedNotFound bool
	}{
		{
			name: "successfully deprovision runner",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode:     http.StatusOK,
			shouldSendSignal: true,
		},
		{
			name: "runner not found returns error - no signal sent",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			shouldSendSignal: false,
			expectedNotFound: true,
		},
		{
			name: "deprovision already deprovisioned runner",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "deprovisioned-runner",
					DisplayName:   "Deprovisioned Runner",
					Status:        app.RunnerStatusDeprovisioned,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode:     http.StatusOK,
			shouldSendSignal: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/runners/"+runnerID+"/deprovision", AdminDeprovisionRunnerRequest{})

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// Verify signal was sent (or not sent)
			if tc.shouldSendSignal {
				capturedSignals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), runnerID, capturedSignals[0].OwnerID)

				// Type assert to verify signal type
				_ = capturedSignals[0] // using .Type directly

				assert.NotEmpty(s.T(), string(capturedSignals[0].Type))
			} else {
				capturedSignals := tests.GetQueueSignals(s.T(), s.service.DB)
				assert.Len(s.T(), capturedSignals, 0)
			}

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else {
				var result bool
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				require.NoError(s.T(), err)
				assert.True(s.T(), result)
			}
		})
	}
}
