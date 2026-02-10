package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// GetOrgRunnerGroupTestSuite is the testify suite for get org runner group endpoint.
type GetOrgRunnerGroupTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     TestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestGetOrgRunnerGroupSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgRunnerGroupTestSuite))
}

func (s *GetOrgRunnerGroupTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GetOrgRunnerGroupTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.orgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgRunnerGroupTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgRunnerGroupTestSuite) setupTestData() {
	// Create test account
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "test-org",
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *GetOrgRunnerGroupTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetOrgRunnerGroupTestSuite) TestGetOrgRunnerGroup() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.RunnerGroup
		expectedStatus int
		validateFunc   func(*app.RunnerGroup)
	}{
		{
			name: "returns runner group for org with preloaded relations",
			setupFunc: func() *app.RunnerGroup {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create runner group for test org
				runnerGroup := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OwnerID:   s.testOrg.ID,
					OwnerType: "orgs",
					OrgID:     s.testOrg.ID,
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err := s.service.DB.WithContext(ctx).Create(runnerGroup).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.RunnerGroup{}, "id = ?", runnerGroup.ID)
				})

				// Create settings for the runner group
				settings := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					ContainerImageURL: "example.com/runner",
					ContainerImageTag: "latest",
				}
				err = s.service.DB.WithContext(ctx).Create(settings).Error
				require.NoError(s.T(), err)

				// Create runners in the runner group
				runner1 := &app.Runner{
					ID:                domains.NewRunnerID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					Name:              "runner-1",
					Status:            app.RunnerStatusActive,
					StatusDescription: "Running",
				}
				err = s.service.DB.WithContext(ctx).Create(runner1).Error
				require.NoError(s.T(), err)

				runner2 := &app.Runner{
					ID:                domains.NewRunnerID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					Name:              "runner-2",
					Status:            app.RunnerStatusActive,
					StatusDescription: "Running",
				}
				err = s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				return runnerGroup
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(rg *app.RunnerGroup) {
				require.NotNil(s.T(), rg)
				assert.Equal(s.T(), s.testOrg.ID, rg.OwnerID)
				assert.Equal(s.T(), "orgs", rg.OwnerType)

				// Verify Runners preloaded
				assert.NotNil(s.T(), rg.Runners)
				assert.Len(s.T(), rg.Runners, 2)
				for _, runner := range rg.Runners {
					assert.Equal(s.T(), rg.ID, runner.RunnerGroupID)
					assert.Equal(s.T(), app.RunnerStatusActive, runner.Status)
				}

				// Verify Settings preloaded
				assert.NotNil(s.T(), rg.Settings)
				assert.Equal(s.T(), rg.ID, rg.Settings.RunnerGroupID)
				assert.Equal(s.T(), "example.com/runner", rg.Settings.ContainerImageURL)
				assert.Equal(s.T(), "latest", rg.Settings.ContainerImageTag)
			},
		},
		{
			name: "returns runner group with no runners",
			setupFunc: func() *app.RunnerGroup {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create runner group without any runners
				runnerGroup := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OwnerID:   s.testOrg.ID,
					OwnerType: "orgs",
					OrgID:     s.testOrg.ID,
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err := s.service.DB.WithContext(ctx).Create(runnerGroup).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.RunnerGroup{}, "id = ?", runnerGroup.ID)
				})

				// Create settings for the runner group
				settings := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					ContainerImageURL: "example.com/runner",
					ContainerImageTag: "v1.0.0",
				}
				err = s.service.DB.WithContext(ctx).Create(settings).Error
				require.NoError(s.T(), err)

				return runnerGroup
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(rg *app.RunnerGroup) {
				require.NotNil(s.T(), rg)
				assert.Equal(s.T(), s.testOrg.ID, rg.OwnerID)

				// Verify Runners is empty array (not nil)
				assert.NotNil(s.T(), rg.Runners)
				assert.Len(s.T(), rg.Runners, 0)

				// Verify Settings still preloaded
				assert.NotNil(s.T(), rg.Settings)
				assert.Equal(s.T(), "v1.0.0", rg.Settings.ContainerImageTag)
			},
		},
		{
			name: "returns runner group with multiple runners in different states",
			setupFunc: func() *app.RunnerGroup {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create runner group
				runnerGroup := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OwnerID:   s.testOrg.ID,
					OwnerType: "orgs",
					OrgID:     s.testOrg.ID,
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err := s.service.DB.WithContext(ctx).Create(runnerGroup).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.RunnerGroup{}, "id = ?", runnerGroup.ID)
				})

				// Create settings
				settings := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					ContainerImageURL: "example.com/runner",
					ContainerImageTag: "latest",
				}
				err = s.service.DB.WithContext(ctx).Create(settings).Error
				require.NoError(s.T(), err)

				// Create runners with different statuses
				activeRunner := &app.Runner{
					ID:                domains.NewRunnerID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					Name:              "active-runner",
					Status:            app.RunnerStatusActive,
					StatusDescription: "Active and running",
				}
				err = s.service.DB.WithContext(ctx).Create(activeRunner).Error
				require.NoError(s.T(), err)

				offlineRunner := &app.Runner{
					ID:                domains.NewRunnerID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					Name:              "offline-runner",
					Status:            app.RunnerStatusOffline,
					StatusDescription: "Connection issues",
				}
				err = s.service.DB.WithContext(ctx).Create(offlineRunner).Error
				require.NoError(s.T(), err)

				unknownRunner := &app.Runner{
					ID:                domains.NewRunnerID(),
					RunnerGroupID:     runnerGroup.ID,
					OrgID:             s.testOrg.ID,
					Name:              "unknown-runner",
					Status:            app.RunnerStatusUnknown,
					StatusDescription: "Status unknown",
				}
				err = s.service.DB.WithContext(ctx).Create(unknownRunner).Error
				require.NoError(s.T(), err)

				return runnerGroup
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(rg *app.RunnerGroup) {
				require.NotNil(s.T(), rg)
				assert.Len(s.T(), rg.Runners, 3)

				// Verify different runner statuses are all included
				statuses := make(map[app.RunnerStatus]bool)
				for _, runner := range rg.Runners {
					statuses[runner.Status] = true
					assert.Equal(s.T(), rg.ID, runner.RunnerGroupID)
				}

				assert.True(s.T(), statuses[app.RunnerStatusActive])
				assert.True(s.T(), statuses[app.RunnerStatusOffline])
				assert.True(s.T(), statuses[app.RunnerStatusUnknown])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			runnerGroup := tc.setupFunc()

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/runner-group")

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response
			var response app.RunnerGroup
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Verify basic runner group data
			assert.Equal(s.T(), runnerGroup.ID, response.ID)
			assert.Equal(s.T(), s.testOrg.ID, response.OrgID)

			// Run validation
			if tc.validateFunc != nil {
				tc.validateFunc(&response)
			}
		})
	}
}

func (s *GetOrgRunnerGroupTestSuite) TestGetOrgRunnerGroupNotFound() {
	testCases := []struct {
		name           string
		setupFunc      func()
		expectedStatus int
	}{
		{
			name: "returns 404 when runner group does not exist",
			setupFunc: func() {
				// Explicitly ensure no runner group exists for test org
				s.service.DB.Unscoped().Where("owner_type = ? AND owner_id = ?", "orgs", s.testOrg.ID).
					Delete(&app.RunnerGroup{})
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			tc.setupFunc()

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/runner-group")

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)
		})
	}
}

func (s *GetOrgRunnerGroupTestSuite) TestGetOrgRunnerGroupOwnershipIsolation() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Org
		expectedStatus int
		validateFunc   func(*app.RunnerGroup)
	}{
		{
			name: "only returns runner group for current org",
			setupFunc: func() *app.Org {
				ctx := context.Background()

				// Create second account and org
				acc2 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "other@example.com",
					Subject:     "other-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

				ctx2 := cctx.SetAccountContext(ctx, acc2)
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err = s.service.DB.WithContext(ctx2).Create(org2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)
				})

				// Create runner groups for both orgs with same context
				ctx1 := cctx.SetAccountContext(ctx, s.testAcc)

				// Runner group for test org
				rg1 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OwnerID:   s.testOrg.ID,
					OwnerType: "orgs",
					OrgID:     s.testOrg.ID,
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx1).Create(rg1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.RunnerGroup{}, "id = ?", rg1.ID)
				})

				settings1 := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					RunnerGroupID:     rg1.ID,
					OrgID:             s.testOrg.ID,
					ContainerImageURL: "example.com/runner",
					ContainerImageTag: "test-org-version",
				}
				err = s.service.DB.WithContext(ctx1).Create(settings1).Error
				require.NoError(s.T(), err)

				// Runner group for other org
				rg2 := &app.RunnerGroup{
					ID:        domains.NewRunnerGroupID(),
					OwnerID:   org2.ID,
					OwnerType: "orgs",
					OrgID:     org2.ID,
					Type:      app.RunnerGroupTypeOrg,
					Platform:  app.AppRunnerTypeAWSEKS,
				}
				err = s.service.DB.WithContext(ctx2).Create(rg2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.RunnerGroup{}, "id = ?", rg2.ID)
				})

				settings2 := &app.RunnerGroupSettings{
					ID:                domains.NewRunnerGroupSettingsID(),
					RunnerGroupID:     rg2.ID,
					OrgID:             org2.ID,
					ContainerImageURL: "example.com/runner",
					ContainerImageTag: "other-org-version",
				}
				err = s.service.DB.WithContext(ctx2).Create(settings2).Error
				require.NoError(s.T(), err)

				return org2
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(rg *app.RunnerGroup) {
				require.NotNil(s.T(), rg)
				// Should only get test org's runner group
				assert.Equal(s.T(), s.testOrg.ID, rg.OwnerID)
				assert.Equal(s.T(), "orgs", rg.OwnerType)

				// Verify it's the correct runner group by checking settings tag
				assert.NotNil(s.T(), rg.Settings)
				assert.Equal(s.T(), "test-org-version", rg.Settings.ContainerImageTag)

				// Verify we didn't get the other org's runner group
				assert.NotEqual(s.T(), "other-org-version", rg.Settings.ContainerImageTag)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			tc.setupFunc()

			// Make request with original test org context (router already set up)
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/runner-group")

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response
			var response app.RunnerGroup
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Run validation
			if tc.validateFunc != nil {
				tc.validateFunc(&response)
			}
		})
	}
}

func (s *GetOrgRunnerGroupTestSuite) TestGetOrgRunnerGroupWithoutOrgContext() {
	// Create router without org context
	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
		// TestOrg intentionally omitted
	})

	err := s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "/v1/orgs/current/runner-group", nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should fail without org context
	require.NotEqual(s.T(), http.StatusOK, rr.Code)
}
