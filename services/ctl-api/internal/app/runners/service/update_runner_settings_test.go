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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type UpdateRunnerSettingsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	RunnersService *service
}

type UpdateRunnerSettingsTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       UpdateRunnerSettingsTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testRunner    *app.Runner
	testRunnerGrp *app.RunnerGroup
	testSettings  *app.RunnerGroupSettings
}

func TestUpdateRunnerSettingsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(UpdateRunnerSettingsTestSuite))
}

func (s *UpdateRunnerSettingsTestSuite) SetupSuite() {
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

func (s *UpdateRunnerSettingsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes (UpdateRunnerSettings is in RegisterPublicRoutes)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *UpdateRunnerSettingsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateRunnerSettingsTestSuite) setupTestData() {
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

func (s *UpdateRunnerSettingsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UpdateRunnerSettingsTestSuite) TestUpdateRunnerSettings() {
	testCases := []struct {
		name           string
		setupFunc      func() (string, UpdateRunnerSettingsRequest)
		expectedCode   int
		validateFunc   func(*app.RunnerGroupSettings)
		expectSignal   bool
		expectedError  bool
		errorSubstring string
	}{
		{
			name: "successfully update all fields",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				return s.testRunner.ID, UpdateRunnerSettingsRequest{
					ContainerImageURL:     "updated-image-url",
					ContainerImageTag:     "v2.0.0",
					RunnerAPIURL:          "https://api.updated.com",
					K8sServiceAccountName: "updated-sa",
					AWSIAMRoleARN:         "arn:aws:iam::987654321:role/updated-role",
				}
			},
			expectedCode: http.StatusOK,
			expectSignal: true,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				// Verify response
				assert.Equal(s.T(), s.testSettings.RunnerGroupID, settings.RunnerGroupID)

				// Verify database state
				var dbSettings app.RunnerGroupSettings
				err := s.service.DB.First(&dbSettings, "runner_group_id = ?", s.testRunnerGrp.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "updated-image-url", dbSettings.ContainerImageURL)
				assert.Equal(s.T(), "v2.0.0", dbSettings.ContainerImageTag)
				assert.Equal(s.T(), "https://api.updated.com", dbSettings.RunnerAPIURL)
				assert.Equal(s.T(), "updated-sa", dbSettings.OrgK8sServiceAccountName)
				assert.Equal(s.T(), "arn:aws:iam::987654321:role/updated-role", dbSettings.OrgAWSIAMRoleARN)
			},
		},
		{
			name: "update container tag triggers restart signal",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				return s.testRunner.ID, UpdateRunnerSettingsRequest{
					ContainerImageTag: "v3.0.0",
				}
			},
			expectedCode: http.StatusOK,
			expectSignal: true,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				signals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), signals, 1, "should send restart signal when image tag changes")
				assert.Equal(s.T(), s.testRunner.ID, signals[0].OwnerID)
			},
		},
		{
			name: "update without container tag does not trigger signal",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				return s.testRunner.ID, UpdateRunnerSettingsRequest{
					RunnerAPIURL: "https://api.nochange.com",
				}
			},
			expectedCode: http.StatusOK,
			expectSignal: false,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				signals := tests.GetQueueSignals(s.T(), s.service.DB)
				assert.Len(s.T(), signals, 0, "should not send signal when only API URL changes")
			},
		},
		{
			name: "update with AWS max instance lifetime (deprecated)",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				maxLifetime := 86400 // 1 day
				return s.testRunner.ID, UpdateRunnerSettingsRequest{
					AWSMaxInstanceLifetime: &maxLifetime, // Deprecated: no longer used by ASG
				}
			},
			expectedCode: http.StatusOK,
			expectSignal: false,
			validateFunc: func(settings *app.RunnerGroupSettings) {
				var dbSettings app.RunnerGroupSettings
				err := s.service.DB.First(&dbSettings, "runner_group_id = ?", s.testRunnerGrp.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), 86400, dbSettings.AWSMaxInstanceLifetime)
			},
		},
		{
			name: "runner not found",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				return "rnrnonexistent123456789012", UpdateRunnerSettingsRequest{
					ContainerImageTag: "v2.0.0",
				}
			},
			expectedCode:   http.StatusNotFound,
			expectedError:  true,
			errorSubstring: "unable to get runner",
		},
		{
			name: "runner with group but no settings",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create a second org to avoid unique index conflict
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "no-settings-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

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

				return runner.ID, UpdateRunnerSettingsRequest{
					ContainerImageTag: "v2.0.0",
				}
			},
			expectedCode:   http.StatusNotFound,
			expectedError:  true,
			errorSubstring: "unable to get runner",
		},
		{
			name: "invalid AWS max instance lifetime - too low (deprecated)",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				maxLifetime := 60 // Less than minimum (86400)
				return s.testRunner.ID, UpdateRunnerSettingsRequest{
					AWSMaxInstanceLifetime: &maxLifetime, // Deprecated: no longer used by ASG
				}
			},
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "min",
		},
		{
			name: "invalid AWS max instance lifetime - too high (deprecated)",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				maxLifetime := 40000000 // More than maximum (31536000)
				return s.testRunner.ID, UpdateRunnerSettingsRequest{
					AWSMaxInstanceLifetime: &maxLifetime, // Deprecated: no longer used by ASG
				}
			},
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "max",
		},
		{
			name: "update runner in different org fails",
			setupFunc: func() (string, UpdateRunnerSettingsRequest) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
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

				// Create settings for org2
				settings2 := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					OrgID:             org2.ID,
					RunnerGroupID:     runnerGrp2.ID,
					ContainerImageURL: "other-image-url",
					ContainerImageTag: "v2.0.0",
					RunnerAPIURL:      "https://api.other.com",
					SandboxMode:       true,
				}
				err = s.service.DB.WithContext(ctx).Create(settings2).Error
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
					s.service.DB.Unscoped().Delete(settings2)
					s.service.DB.Unscoped().Delete(runnerGrp2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return runner2.ID, UpdateRunnerSettingsRequest{
					ContainerImageTag: "v3.0.0",
				}
			},
			expectedCode:   http.StatusNotFound,
			expectedError:  true,
			errorSubstring: "unable to get runner",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID, req := tc.setupFunc()
			rr := s.makeRequest("PATCH", "/v1/runners/"+runnerID+"/settings", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedError {
				body := rr.Body.String()
				if tc.errorSubstring != "" {
					assert.Contains(s.T(), body, tc.errorSubstring)
				}
			} else if tc.validateFunc != nil {
				var settings app.RunnerGroupSettings
				err := json.Unmarshal(rr.Body.Bytes(), &settings)
				require.NoError(s.T(), err)
				tc.validateFunc(&settings)
			}

			// Verify signal expectations
			signals := tests.GetQueueSignals(s.T(), s.service.DB)
			if tc.expectSignal {
				assert.NotEmpty(s.T(), signals, "expected signal to be sent")
			} else if !tc.expectedError {
				assert.Empty(s.T(), signals, "expected no signal to be sent")
			}
		})
	}
}

func (s *UpdateRunnerSettingsTestSuite) TestUpdateRunnerSettingsFullUpdate() {
	// Verify that sending all fields in a single request updates them all correctly
	req := UpdateRunnerSettingsRequest{
		ContainerImageURL:     "full-update-url",
		ContainerImageTag:     "v3.0.0",
		RunnerAPIURL:          "https://api.full.com",
		K8sServiceAccountName: "full-sa",
		AWSIAMRoleARN:         "arn:aws:iam::111111111:role/full-role",
	}
	rr := s.makeRequest("PATCH", "/v1/runners/"+s.testRunner.ID+"/settings", req)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Verify all fields were updated in the database
	var settings app.RunnerGroupSettings
	err := s.service.DB.First(&settings, "runner_group_id = ?", s.testRunnerGrp.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "full-update-url", settings.ContainerImageURL)
	assert.Equal(s.T(), "v3.0.0", settings.ContainerImageTag)
	assert.Equal(s.T(), "https://api.full.com", settings.RunnerAPIURL)
	assert.Equal(s.T(), "full-sa", settings.OrgK8sServiceAccountName)
	assert.Equal(s.T(), "arn:aws:iam::111111111:role/full-role", settings.OrgAWSIAMRoleARN)
}

// Deprecated: AWSMaxInstanceLifetime is no longer used by ASG. Instance refresh is handled by a backend cron.
func (s *UpdateRunnerSettingsTestSuite) TestUpdateRunnerSettingsValidAWSMaxInstanceLifetime() {
	testCases := []struct {
		name     string
		lifetime int
		valid    bool
	}{
		{"minimum valid - 1 day", 86400, true},
		{"mid range - 7 days", 604800, true},
		{"maximum valid - 365 days", 31536000, true},
		{"below minimum", 86399, false},
		{"above maximum", 31536001, false},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := UpdateRunnerSettingsRequest{
				AWSMaxInstanceLifetime: &tc.lifetime,
			}
			rr := s.makeRequest("PATCH", "/v1/runners/"+s.testRunner.ID+"/settings", req)

			if tc.valid {
				if rr.Code != http.StatusOK {
					s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
				}
				require.Equal(s.T(), http.StatusOK, rr.Code)

				var dbSettings app.RunnerGroupSettings
				err := s.service.DB.First(&dbSettings, "runner_group_id = ?", s.testRunnerGrp.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tc.lifetime, dbSettings.AWSMaxInstanceLifetime)
			} else {
				require.Equal(s.T(), http.StatusBadRequest, rr.Code)
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}
