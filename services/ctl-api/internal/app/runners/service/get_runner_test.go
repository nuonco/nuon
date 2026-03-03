package service

import (
	"context"
	"encoding/json"
	"fmt"
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

type GetRunnerTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetRunnerTestSuite struct {
	tests.BaseDBTestSuite
	app                   *fxtest.App
	service               GetRunnerTestService
	router                *gin.Engine
	testOrg               *app.Org
	testAcc               *app.Account
	testRunner            *app.Runner
	testRunnerGrp         *app.RunnerGroup
	testRunnerGrpSettings *app.RunnerGroupSettings
}

func TestGetRunnerSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerTestSuite))
}

func (s *GetRunnerTestSuite) SetupSuite() {
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
}

func (s *GetRunnerTestSuite) SetupTest() {
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

func (s *GetRunnerTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerTestSuite) setupTestData() {
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

func (s *GetRunnerTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerTestSuite) TestGetRunner() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.Runner)
		expectedNotFound bool
	}{
		{
			name: "successfully get runner with preloaded associations",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(runner *app.Runner) {
				assert.Equal(s.T(), s.testOrg.ID, runner.OrgID)
				assert.Equal(s.T(), s.testRunnerGrp.ID, runner.RunnerGroupID)
				assert.Equal(s.T(), "test-runner", runner.Name)
				assert.Equal(s.T(), "Test Runner", runner.DisplayName)
				assert.Equal(s.T(), app.RunnerStatusActive, runner.Status)

				// Note: Org field has json:"-" tag so it won't be in HTTP response
				// Only OrgID is returned

				// Verify RunnerGroup is preloaded
				assert.Equal(s.T(), s.testRunnerGrp.ID, runner.RunnerGroup.ID)

				// Verify RunnerGroup.Settings is preloaded
				assert.Equal(s.T(), s.testRunnerGrpSettings.ID, runner.RunnerGroup.Settings.ID)
				assert.Equal(s.T(), "test.ecr.aws/runner", runner.RunnerGroup.Settings.ContainerImageURL)
				assert.Equal(s.T(), "v1.0.0", runner.RunnerGroup.Settings.ContainerImageTag)
			},
		},
		{
			name: "runner not found returns error",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "get runner from different org",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/other",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

				// Create runner group for org2
				runnerGrp2 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     org2.ID,
					OwnerID:   org2.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx).Create(runnerGrp2).Error
				require.NoError(s.T(), err)

				// Create runner in org2
				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         org2.ID,
					Name:          "runner-org2",
					DisplayName:   "Runner in Org 2",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: runnerGrp2.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(runnerGrp2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return runner2.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(runner *app.Runner) {
				// Runner routes don't enforce org scoping - can access any runner by ID
				assert.Equal(s.T(), "runner-org2", runner.Name)
				assert.Equal(s.T(), "Runner in Org 2", runner.DisplayName)
			},
		},
		{
			name: "multiple runners exist, correct one returned by ID",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "test-runner-2",
					DisplayName:   "Test Runner 2",
					Status:        app.RunnerStatusActive,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				runner3 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "test-runner-3",
					DisplayName:   "Test Runner 3",
					Status:        app.RunnerStatusPending,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner3).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
					s.service.DB.Unscoped().Delete(runner3)
				})

				return runner2.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(runner *app.Runner) {
				assert.Equal(s.T(), "test-runner-2", runner.Name)
				assert.Equal(s.T(), "Test Runner 2", runner.DisplayName)
				assert.Equal(s.T(), app.RunnerStatusActive, runner.Status)
			},
		},
		{
			name: "runner with different status values",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					Name:          "pending-runner",
					DisplayName:   "Pending Runner",
					Status:        app.RunnerStatusPending,
					RunnerGroupID: s.testRunnerGrp.ID,
				}
				err := s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(runner *app.Runner) {
				assert.Equal(s.T(), app.RunnerStatusPending, runner.Status)
				assert.Equal(s.T(), "pending-runner", runner.Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runners/"+runnerID)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var runner app.Runner
				err := json.Unmarshal(rr.Body.Bytes(), &runner)
				require.NoError(s.T(), err)
				tc.validateFunc(&runner)
			}
		})
	}
}
