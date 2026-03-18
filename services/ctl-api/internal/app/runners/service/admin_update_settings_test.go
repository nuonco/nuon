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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminUpdateSettingsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminUpdateSettingsTestSuite struct {
	tests.BaseDBTestSuite
	app               *fxtest.App
	service           AdminUpdateSettingsTestService
	router            *gin.Engine
	testOrg           *app.Org
	testAcc           *app.Account
	testRunner        *app.Runner
	testRunnerGrp     *app.RunnerGroup
	testRunnerGrpSett *app.RunnerGroupSettings
	mockEvClient      *tests.MockEventLoopClient
}

func TestAdminUpdateSettingsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminUpdateSettingsTestSuite))
}

func (s *AdminUpdateSettingsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewMockEventLoopClient()
	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminUpdateSettingsTestSuite) SetupTest() {
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

func (s *AdminUpdateSettingsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminUpdateSettingsTestSuite) setupTestData() {
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
	s.testRunnerGrpSett = &app.RunnerGroupSettings{
		ID:                domains.NewRunnerGroupSettingsID(),
		OrgID:             s.testOrg.ID,
		RunnerGroupID:     s.testRunnerGrp.ID,
		ContainerImageURL: "original.ecr.aws/runner",
		ContainerImageTag: "v1.0.0",
		RunnerAPIURL:      "https://runner-api.test.com",
		SandboxMode:       true,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerGrpSett).Error
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

func (s *AdminUpdateSettingsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminUpdateSettingsTestSuite) TestAdminUpdateRunnerSettings() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		requestBody      map[string]interface{}
		expectedCode     int
		expectedSignal   bool
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully update settings without tag change",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			requestBody: map[string]interface{}{
				"container_image_url": "updated.ecr.aws/runner",
				"runner_api_url":      "https://updated-api.test.com",
			},
			expectedCode:   http.StatusOK,
			expectedSignal: false,
			validateFunc: func(runnerID string) {
				var settings app.RunnerGroupSettings
				err := s.service.DB.Where("runner_group_id = ?", s.testRunnerGrp.ID).First(&settings).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "updated.ecr.aws/runner", settings.ContainerImageURL)
				assert.Equal(s.T(), "https://updated-api.test.com", settings.RunnerAPIURL)
				// Original tag should be preserved
				assert.Equal(s.T(), "v1.0.0", settings.ContainerImageTag)
			},
		},
		{
			name: "update settings with tag change triggers restart signal",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			requestBody: map[string]interface{}{
				"container_image_tag": "v2.0.0",
			},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(runnerID string) {
				var settings app.RunnerGroupSettings
				err := s.service.DB.Where("runner_group_id = ?", s.testRunnerGrp.ID).First(&settings).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "v2.0.0", settings.ContainerImageTag)

				// Verify restart signal sent
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1)
				assert.Equal(s.T(), runnerID, capturedSignals[0].ID)

				sig, ok := capturedSignals[0].Signal.(*signals.Signal)
				require.True(s.T(), ok)
				assert.Equal(s.T(), signals.OperationRestart, sig.Type)
			},
		},
		{
			name: "update with valid aws_max_instance_lifetime",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			requestBody: map[string]interface{}{
				"aws_max_instance_lifetime": 86400, // 1 day (min value)
			},
			expectedCode:   http.StatusOK,
			expectedSignal: false,
			validateFunc: func(runnerID string) {
				var settings app.RunnerGroupSettings
				err := s.service.DB.Where("runner_group_id = ?", s.testRunnerGrp.ID).First(&settings).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), 86400, settings.AWSMaxInstanceLifetime) // Deprecated: no longer used by ASG
			},
		},
		{
			name: "validation error for aws_max_instance_lifetime too low",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			requestBody: map[string]interface{}{
				"aws_max_instance_lifetime": 1000, // Below min
			},
			expectedCode:   http.StatusBadRequest,
			expectedSignal: false,
		},
		{
			name: "validation error for aws_max_instance_lifetime too high",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			requestBody: map[string]interface{}{
				"aws_max_instance_lifetime": 999999999, // Above max
			},
			expectedCode:   http.StatusBadRequest,
			expectedSignal: false,
		},
		{
			name: "nonexistent runner returns error",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			requestBody: map[string]interface{}{
				"container_image_url": "test.ecr.aws/runner",
			},
			expectedCode:     http.StatusNotFound,
			expectedSignal:   false,
			expectedNotFound: true,
		},
		{
			name: "update runner in different org",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

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

				settings2 := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					OrgID:             org2.ID,
					RunnerGroupID:     runnerGrp2.ID,
					ContainerImageURL: "org2.ecr.aws/runner",
					ContainerImageTag: "v1.0.0",
					RunnerAPIURL:      "https://org2-api.test.com",
					SandboxMode:       true,
				}
				err = s.service.DB.WithContext(ctx).Create(settings2).Error
				require.NoError(s.T(), err)

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
					s.service.DB.Unscoped().Delete(settings2)
					s.service.DB.Unscoped().Delete(runnerGrp2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return runner2.ID
			},
			requestBody: map[string]interface{}{
				"container_image_tag": "v3.0.0",
			},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockEvClient.Reset()
			runnerID := tc.setupFunc()
			rr := s.makeRequest("PATCH", "/v1/runners/"+runnerID+"/settings", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(runnerID)
			}

			// Verify signal presence matches expectation
			capturedSignals := s.mockEvClient.GetSignals()
			if tc.expectedSignal {
				assert.Len(s.T(), capturedSignals, 1, "expected signal to be sent")
			} else {
				assert.Len(s.T(), capturedSignals, 0, "expected no signal to be sent")
			}
		})
	}
}
