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

type CreateHealthCheckTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type CreateHealthCheckTestSuite struct {
	tests.BaseDBTestSuite
	app                   *fxtest.App
	service               CreateHealthCheckTestService
	router                *gin.Engine
	testOrg               *app.Org
	testAcc               *app.Account
	testRunner            *app.Runner
	testRunnerGrp         *app.RunnerGroup
	testRunnerGrpSettings *app.RunnerGroupSettings
}

func TestCreateHealthCheckSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(CreateHealthCheckTestSuite))
}

func (s *CreateHealthCheckTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
	s.SetCHDB(s.service.CHDB)
}

func (s *CreateHealthCheckTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner routes (no TestOrg/TestAcc needed for runner routes)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *CreateHealthCheckTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateHealthCheckTestSuite) setupTestData() {
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

	// Create runner group settings
	s.testRunnerGrpSettings = &app.RunnerGroupSettings{
		ID:                domains.NewRunnerGroupSettingsID(),
		OrgID:             s.testOrg.ID,
		RunnerGroupID:     s.testRunnerGrp.ID,
		ContainerImageURL: "test.ecr.aws/runner",
		ContainerImageTag: "v1.0.0",
		RunnerAPIURL:      "https://runner-api.test.com",
		SandboxMode:       true,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerGrpSettings).Error
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

func (s *CreateHealthCheckTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
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

func (s *CreateHealthCheckTestSuite) getHealthCheckFromClickHouse(healthCheckID string) *app.RunnerHealthCheck {
	ctx := context.Background()
	var healthCheck app.RunnerHealthCheck
	err := s.service.CHDB.WithContext(ctx).
		Where("id = ?", healthCheckID).
		First(&healthCheck).Error
	if err != nil {
		return nil
	}
	return &healthCheck
}

func (s *CreateHealthCheckTestSuite) countHealthChecksInClickHouse(runnerID string) int {
	ctx := context.Background()
	var count int64
	s.service.CHDB.WithContext(ctx).
		Model(&app.RunnerHealthCheck{}).
		Where("runner_id = ?", runnerID).
		Count(&count)
	return int(count)
}

func (s *CreateHealthCheckTestSuite) TestCreateRunnerHealthCheck() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, CreateRunnerHealthCheckRequest)
		expectedCode int
		validateFunc func(*app.RunnerHealthCheck, string)
	}{
		{
			name: "successfully creates health check with mng process",
			setupFunc: func() (string, CreateRunnerHealthCheckRequest) {
				return s.testRunner.ID, CreateRunnerHealthCheckRequest{
					Process: app.RunnerProcessTypeMng,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(healthCheck *app.RunnerHealthCheck, runnerID string) {
				assert.NotEmpty(s.T(), healthCheck.ID)
				assert.Equal(s.T(), runnerID, healthCheck.RunnerID)
				assert.Equal(s.T(), app.RunnerProcessTypeMng, healthCheck.Process)

				// Verify stored in ClickHouse
				stored := s.getHealthCheckFromClickHouse(healthCheck.ID)
				require.NotNil(s.T(), stored)
				assert.Equal(s.T(), healthCheck.ID, stored.ID)
				assert.Equal(s.T(), runnerID, stored.RunnerID)
				assert.Equal(s.T(), app.RunnerProcessTypeMng, stored.Process)
			},
		},
		{
			name: "successfully creates health check with install process",
			setupFunc: func() (string, CreateRunnerHealthCheckRequest) {
				return s.testRunner.ID, CreateRunnerHealthCheckRequest{
					Process: app.RunnerProcessTypeInstall,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(healthCheck *app.RunnerHealthCheck, runnerID string) {
				assert.NotEmpty(s.T(), healthCheck.ID)
				assert.Equal(s.T(), runnerID, healthCheck.RunnerID)
				assert.Equal(s.T(), app.RunnerProcessTypeInstall, healthCheck.Process)

				// Verify stored in ClickHouse
				stored := s.getHealthCheckFromClickHouse(healthCheck.ID)
				require.NotNil(s.T(), stored)
				assert.Equal(s.T(), app.RunnerProcessTypeInstall, stored.Process)
			},
		},
		{
			name: "successfully creates health check with org process",
			setupFunc: func() (string, CreateRunnerHealthCheckRequest) {
				return s.testRunner.ID, CreateRunnerHealthCheckRequest{
					Process: app.RunnerProcessTypeOrg,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(healthCheck *app.RunnerHealthCheck, runnerID string) {
				assert.NotEmpty(s.T(), healthCheck.ID)
				assert.Equal(s.T(), runnerID, healthCheck.RunnerID)
				assert.Equal(s.T(), app.RunnerProcessTypeOrg, healthCheck.Process)

				// Verify stored in ClickHouse
				stored := s.getHealthCheckFromClickHouse(healthCheck.ID)
				require.NotNil(s.T(), stored)
				assert.Equal(s.T(), app.RunnerProcessTypeOrg, stored.Process)
			},
		},
		{
			name: "empty process defaults to unknown",
			setupFunc: func() (string, CreateRunnerHealthCheckRequest) {
				return s.testRunner.ID, CreateRunnerHealthCheckRequest{
					Process: "",
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(healthCheck *app.RunnerHealthCheck, runnerID string) {
				assert.NotEmpty(s.T(), healthCheck.ID)
				assert.Equal(s.T(), runnerID, healthCheck.RunnerID)
				assert.Equal(s.T(), app.RunnerProcessTypeUnknown, healthCheck.Process)

				// Verify stored in ClickHouse
				stored := s.getHealthCheckFromClickHouse(healthCheck.ID)
				require.NotNil(s.T(), stored)
				assert.Equal(s.T(), app.RunnerProcessTypeUnknown, stored.Process)
			},
		},
		{
			name: "creates multiple health checks for same runner",
			setupFunc: func() (string, CreateRunnerHealthCheckRequest) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create a new runner for this test
				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "test-runner-multiple",
					DisplayName:   "Test Runner Multiple",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				// Create first health check
				firstReq := CreateRunnerHealthCheckRequest{
					Process: app.RunnerProcessTypeMng,
				}
				path := fmt.Sprintf("/v1/runners/%s/health-checks", runner.ID)
				rr := s.makeRequest("POST", path, firstReq)
				require.Equal(s.T(), http.StatusCreated, rr.Code)

				// Allow ClickHouse eventual consistency for the first health check write
				time.Sleep(200 * time.Millisecond)

				return runner.ID, CreateRunnerHealthCheckRequest{
					Process: app.RunnerProcessTypeInstall,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(healthCheck *app.RunnerHealthCheck, runnerID string) {
				// Allow ClickHouse eventual consistency for the second health check write
				time.Sleep(200 * time.Millisecond)

				// Should have 2 health checks for this runner
				count := s.countHealthChecksInClickHouse(runnerID)
				assert.Equal(s.T(), 2, count)
			},
		},
		{
			name: "creates health check for non-existent runner (ClickHouse doesn't validate foreign keys)",
			setupFunc: func() (string, CreateRunnerHealthCheckRequest) {
				nonExistentID := "rn1nonexistent123456789012"
				return nonExistentID, CreateRunnerHealthCheckRequest{
					Process: app.RunnerProcessTypeMng,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(healthCheck *app.RunnerHealthCheck, runnerID string) {
				// ClickHouse doesn't enforce foreign key constraints
				assert.NotEmpty(s.T(), healthCheck.ID)
				assert.Equal(s.T(), runnerID, healthCheck.RunnerID)

				// Verify stored in ClickHouse
				stored := s.getHealthCheckFromClickHouse(healthCheck.ID)
				require.NotNil(s.T(), stored)
			},
		},
		{
			name: "invalid JSON body returns 400",
			setupFunc: func() (string, CreateRunnerHealthCheckRequest) {
				// Will be handled specially in test execution
				return s.testRunner.ID, CreateRunnerHealthCheckRequest{}
			},
			expectedCode: http.StatusBadRequest,
			validateFunc: func(healthCheck *app.RunnerHealthCheck, runnerID string) {
				// No health check should be created
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID, req := tc.setupFunc()

			var rr *httptest.ResponseRecorder
			if tc.name == "invalid JSON body returns 400" {
				// Send malformed JSON
				path := fmt.Sprintf("/v1/runners/%s/health-checks", runnerID)
				reqBody := bytes.NewBuffer([]byte(`{"process": invalid}`))
				httpReq, err := http.NewRequest("POST", path, reqBody)
				require.NoError(s.T(), err)
				httpReq.Header.Set("Content-Type", "application/json")
				rr = httptest.NewRecorder()
				s.router.ServeHTTP(rr, httpReq)
			} else {
				path := fmt.Sprintf("/v1/runners/%s/health-checks", runnerID)
				rr = s.makeRequest("POST", path, req)
			}

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var healthCheck app.RunnerHealthCheck
				err := json.Unmarshal(rr.Body.Bytes(), &healthCheck)
				require.NoError(s.T(), err)

				if tc.validateFunc != nil {
					tc.validateFunc(&healthCheck, runnerID)
				}
			} else {
				// Error cases should have error in response
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}

func (s *CreateHealthCheckTestSuite) TestCreateRunnerHealthCheckProcessValidation() {
	testCases := []struct {
		name         string
		process      app.RunnerProcessType
		expectedCode int
	}{
		{
			name:         "valid mng process",
			process:      app.RunnerProcessTypeMng,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "valid install process",
			process:      app.RunnerProcessTypeInstall,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "valid org process",
			process:      app.RunnerProcessTypeOrg,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "empty process defaults to unknown",
			process:      "",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "unknown process value accepted (stored as-is)",
			process:      "custom-process",
			expectedCode: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := CreateRunnerHealthCheckRequest{
				Process: tc.process,
			}

			path := fmt.Sprintf("/v1/runners/%s/health-checks", s.testRunner.ID)
			rr := s.makeRequest("POST", path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var healthCheck app.RunnerHealthCheck
				err := json.Unmarshal(rr.Body.Bytes(), &healthCheck)
				require.NoError(s.T(), err)

				expectedProcess := tc.process
				if tc.process == "" {
					expectedProcess = app.RunnerProcessTypeUnknown
				}
				assert.Equal(s.T(), expectedProcess, healthCheck.Process)

				// Verify stored in ClickHouse
				stored := s.getHealthCheckFromClickHouse(healthCheck.ID)
				require.NotNil(s.T(), stored)
				assert.Equal(s.T(), expectedProcess, stored.Process)
			}
		})
	}
}
