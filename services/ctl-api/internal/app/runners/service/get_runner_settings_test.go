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
)

type GetRunnerSettingsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	RunnersService *service
}

type GetRunnerSettingsTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetRunnerSettingsTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testSettings  *app.RunnerGroupSettings
}

func TestGetRunnerSettingsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetRunnerSettingsTestSuite))
}

func (s *GetRunnerSettingsTestSuite) SetupSuite() {
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

func (s *GetRunnerSettingsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner service routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetRunnerSettingsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetRunnerSettingsTestSuite) setupTestData() {
	ctx := context.Background()

	// Create test account with unique ID-based email to avoid conflicts
	accID := domains.NewAccountID()
	testAcc := &app.Account{
		ID:          accID,
		Email:       fmt.Sprintf("%s@test.nuon.co", accID),
		Subject:     accID,
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          testOrgID,
		Name:        fmt.Sprintf("test-org-%s", testOrgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg

	// Create runner group
	testRunnerGrp := &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     testOrg.ID,
		OwnerID:   testOrg.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err = s.service.DB.WithContext(ctx).Create(testRunnerGrp).Error
	require.NoError(s.T(), err)
	s.testRunnerGrp = testRunnerGrp

	// Create runner group settings
	testSettings := &app.RunnerGroupSettings{
		ID:                       domains.NewRunnerGroupSettingsID(),
		OrgID:                    testOrg.ID,
		RunnerGroupID:            testRunnerGrp.ID,
		ContainerImageURL:        "test-image-url",
		ContainerImageTag:        "v1.0.0",
		RunnerAPIURL:             "https://api.test.com",
		SandboxMode:              true,
		EnableSentry:             true,
		EnableMetrics:            true,
		EnableLogging:            true,
		LoggingLevel:             "info",
		OrgK8sServiceAccountName: "test-sa",
		OrgAWSIAMRoleARN:         "arn:aws:iam::123456789:role/test-role",
		AWSMaxInstanceLifetime:   604800, // Deprecated: no longer used by ASG
	}
	err = s.service.DB.WithContext(ctx).Create(testSettings).Error
	require.NoError(s.T(), err)
	s.testSettings = testSettings

	// Create runner
	testRunner := &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         testOrg.ID,
		Name:          "test-runner",
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
		RunnerGroupID: testRunnerGrp.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(testRunner).Error
	require.NoError(s.T(), err)
	s.testRunner = testRunner
}

func (s *GetRunnerSettingsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetRunnerSettingsTestSuite) TestGetRunnerSettings() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.RunnerGroupSettings)
		expectedNotFound bool
	}{
		{
			name: "successfully get runner settings",
			setupFunc: func() string {
				return s.testRunner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				assert.Equal(s.T(), s.testSettings.ID, settings.ID)
				assert.Equal(s.T(), s.testSettings.ContainerImageURL, settings.ContainerImageURL)
				assert.Equal(s.T(), s.testSettings.ContainerImageTag, settings.ContainerImageTag)
				assert.Equal(s.T(), s.testSettings.RunnerAPIURL, settings.RunnerAPIURL)
				assert.Equal(s.T(), s.testSettings.OrgK8sServiceAccountName, settings.OrgK8sServiceAccountName)
				assert.Equal(s.T(), s.testSettings.OrgAWSIAMRoleARN, settings.OrgAWSIAMRoleARN)
				assert.Equal(s.T(), s.testSettings.EnableSentry, settings.EnableSentry)
				assert.Equal(s.T(), s.testSettings.EnableMetrics, settings.EnableMetrics)
				assert.Equal(s.T(), s.testSettings.EnableLogging, settings.EnableLogging)
				assert.Equal(s.T(), s.testSettings.LoggingLevel, settings.LoggingLevel)
				assert.Equal(s.T(), s.testSettings.AWSMaxInstanceLifetime, settings.AWSMaxInstanceLifetime) // Deprecated: no longer used by ASG
			},
		},
		{
			name: "runner not found",
			setupFunc: func() string {
				return "rnrnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "runner with group but no settings returns empty settings",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create a second org to avoid unique index conflict on owner_id
				noSettingsOrgID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          noSettingsOrgID,
					Name:        fmt.Sprintf("no-settings-org-%s", noSettingsOrgID[:8]),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

				// Create runner group without settings
				runnerGrp := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OrgID:     org2.ID,
					OwnerID:   org2.ID,
					OwnerType: "org",
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx).Create(runnerGrp).Error
				require.NoError(s.T(), err)

				runner := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         org2.ID,
					Name:          "runner-no-settings",
					DisplayName:   "Runner Without Settings",
					Status:        app.RunnerStatusPending,
					RunnerGroupID: runnerGrp.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(runner).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner)
					s.service.DB.Unscoped().Delete(runnerGrp)
					s.service.DB.Unscoped().Delete(org2)
				})

				return runner.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				// Settings should be empty/zero since none were created
				assert.Empty(s.T(), settings.ContainerImageURL)
				assert.Empty(s.T(), settings.ContainerImageTag)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/runners/"+runnerID+"/settings")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				// Verify error response for not found cases
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var settings app.RunnerGroupSettings
				err := json.Unmarshal(rr.Body.Bytes(), &settings)
				require.NoError(s.T(), err)
				tc.validateFunc(&settings)
			}
		})
	}
}

func (s *GetRunnerSettingsTestSuite) TestGetRunnerSettingsMultipleRunners() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	// Create a second org to avoid unique index conflict on (deleted_at, owner_id)
	multiRunnerOrgID := domains.NewOrgID()
	org2 := &app.Org{
		ID:          multiRunnerOrgID,
		Name:        fmt.Sprintf("multi-runner-org-%s", multiRunnerOrgID[:8]),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/bar",
		},
	}
	err := s.service.DB.WithContext(ctx).Create(org2).Error
	require.NoError(s.T(), err)

	// Create second runner group with different settings under the second org
	runnerGrp2 := &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     org2.ID,
		OwnerID:   org2.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSECS,
	}
	err = s.service.DB.WithContext(ctx).Create(runnerGrp2).Error
	require.NoError(s.T(), err)

	settings2 := &app.RunnerGroupSettings{
		ID:                     domains.NewRunnerGroupSettingsID(),
		OrgID:                  org2.ID,
		RunnerGroupID:          runnerGrp2.ID,
		ContainerImageURL:      "different-image-url",
		ContainerImageTag:      "v2.5.0",
		RunnerAPIURL:           "https://api.different.com",
		SandboxMode:            false,
		AWSMaxInstanceLifetime: 86400, // 1 day. Deprecated: no longer used by ASG
		AWSInstanceType:        "t3.large",
	}
	err = s.service.DB.WithContext(ctx).Create(settings2).Error
	require.NoError(s.T(), err)

	runner2 := &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         org2.ID,
		Name:          "test-runner-2",
		DisplayName:   "Test Runner 2",
		Status:        app.RunnerStatusActive,
		RunnerGroupID: runnerGrp2.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(runner2).Error
	require.NoError(s.T(), err)

	// Test first runner settings
	rr1 := s.makeRequest("GET", "/v1/runners/"+s.testRunner.ID+"/settings")
	require.Equal(s.T(), http.StatusOK, rr1.Code)

	var settings1Result app.RunnerGroupSettings
	err = json.Unmarshal(rr1.Body.Bytes(), &settings1Result)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testSettings.ID, settings1Result.ID)
	assert.Equal(s.T(), s.testSettings.ContainerImageTag, settings1Result.ContainerImageTag)

	// Test second runner settings
	rr2 := s.makeRequest("GET", "/v1/runners/"+runner2.ID+"/settings")
	require.Equal(s.T(), http.StatusOK, rr2.Code)

	var settings2Result app.RunnerGroupSettings
	err = json.Unmarshal(rr2.Body.Bytes(), &settings2Result)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), settings2.ID, settings2Result.ID)
	assert.Equal(s.T(), settings2.ContainerImageTag, settings2Result.ContainerImageTag)
	assert.Equal(s.T(), settings2.AWSMaxInstanceLifetime, settings2Result.AWSMaxInstanceLifetime) // Deprecated: no longer used by ASG
	assert.Equal(s.T(), settings2.AWSInstanceType, settings2Result.AWSInstanceType)

	// Verify settings are different
	assert.NotEqual(s.T(), settings1Result.ID, settings2Result.ID)
	assert.NotEqual(s.T(), settings1Result.ContainerImageTag, settings2Result.ContainerImageTag)
}
