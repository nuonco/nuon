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
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// AdminEnableDebugForOrgTestService holds all fx-injected dependencies.
type AdminEnableDebugForOrgTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
}

// AdminEnableDebugForOrgTestSuite is the testify suite for admin debug mode endpoint.
type AdminEnableDebugForOrgTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     AdminEnableDebugForOrgTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestAdminEnableDebugForOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminEnableDebugForOrgTestSuite))
}

func (s *AdminEnableDebugForOrgTestSuite) SetupSuite() {
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

func (s *AdminEnableDebugForOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.orgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminEnableDebugForOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminEnableDebugForOrgTestSuite) setupTestData() {
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

func (s *AdminEnableDebugForOrgTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(s.T(), err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(s.T(), err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminEnableDebugForOrgTestSuite) TestAdminDebugModeOrg() {
	testCases := []struct {
		name               string
		setupFunc          func() *app.Org
		requestBody        DebugModeRequest
		expectedStatus     int
		expectedDebugValue bool
		validateFunc       func(*app.Org)
	}{
		{
			name: "successfully enables debug mode",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-enable-debug",
					SandboxMode: true,
					DebugMode:   false,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: DebugModeRequest{
				DebugMode: true,
			},
			expectedStatus:     http.StatusOK,
			expectedDebugValue: true,
			validateFunc: func(org *app.Org) {
				// Verify database state
				var dbOrg app.Org
				err := s.service.DB.Where("id = ?", org.ID).First(&dbOrg).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.DebugMode, "debug mode should be enabled in database")
			},
		},
		{
			name: "successfully disables debug mode",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-disable-debug",
					SandboxMode: true,
					DebugMode:   true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: DebugModeRequest{
				DebugMode: false,
			},
			expectedStatus:     http.StatusOK,
			expectedDebugValue: false,
			validateFunc: func(org *app.Org) {
				// Verify database state
				var dbOrg app.Org
				err := s.service.DB.Where("id = ?", org.ID).First(&dbOrg).Error
				require.NoError(s.T(), err)
				assert.False(s.T(), dbOrg.DebugMode, "debug mode should be disabled in database")
			},
		},
		{
			name: "enables debug mode when already enabled (idempotent)",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-idempotent-enable",
					SandboxMode: true,
					DebugMode:   true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: DebugModeRequest{
				DebugMode: true,
			},
			expectedStatus:     http.StatusOK,
			expectedDebugValue: true,
			validateFunc: func(org *app.Org) {
				// Verify database state remains unchanged
				var dbOrg app.Org
				err := s.service.DB.Where("id = ?", org.ID).First(&dbOrg).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbOrg.DebugMode, "debug mode should still be enabled")
			},
		},
		{
			name: "disables debug mode when already disabled (idempotent)",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-idempotent-disable",
					SandboxMode: true,
					DebugMode:   false,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: DebugModeRequest{
				DebugMode: false,
			},
			expectedStatus:     http.StatusOK,
			expectedDebugValue: false,
			validateFunc: func(org *app.Org) {
				// Verify database state remains unchanged
				var dbOrg app.Org
				err := s.service.DB.Where("id = ?", org.ID).First(&dbOrg).Error
				require.NoError(s.T(), err)
				assert.False(s.T(), dbOrg.DebugMode, "debug mode should still be disabled")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Make request
			path := "/v1/orgs/" + org.ID + "/admin-debug-mode"
			rr := s.makeRequest(http.MethodPost, path, tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response
			var response bool
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			assert.True(s.T(), response, "response should be true for successful operation")

			// Run validation
			if tc.validateFunc != nil {
				tc.validateFunc(org)
			}
		})
	}
}

func (s *AdminEnableDebugForOrgTestSuite) TestAdminDebugModeOrgErrors() {
	testCases := []struct {
		name           string
		setupFunc      func() string
		requestBody    interface{}
		expectedStatus int
		validateError  func(string)
	}{
		{
			name: "returns error for non-existent org",
			setupFunc: func() string {
				// Return a valid-format but non-existent org ID
				return domains.NewOrgID()
			},
			requestBody: DebugModeRequest{
				DebugMode: true,
			},
			expectedStatus: http.StatusNotFound,
			validateError: func(body string) {
				assert.Contains(s.T(), body, "not found")
			},
		},
		{
			name: "returns error for invalid request body",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-invalid-body",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org.ID
			},
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
			validateError: func(body string) {
				assert.Contains(s.T(), body, "unable to parse request")
			},
		},
		{
			name: "returns error for empty org_id",
			setupFunc: func() string {
				// Return empty org ID
				return ""
			},
			requestBody: DebugModeRequest{
				DebugMode: true,
			},
			expectedStatus: http.StatusNotFound,
			validateError: func(body string) {
				assert.Contains(s.T(), body, "not found")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			orgID := tc.setupFunc()

			// Make request
			path := "/v1/orgs/" + orgID + "/admin-debug-mode"
			rr := s.makeRequest(http.MethodPost, path, tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate error message
			if tc.validateError != nil {
				tc.validateError(rr.Body.String())
			}
		})
	}
}

func (s *AdminEnableDebugForOrgTestSuite) TestAdminDebugModeToggle() {
	// This test verifies that debug mode can be toggled multiple times
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	org := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "test-toggle",
		SandboxMode: true,
		DebugMode:   false,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err := s.service.DB.WithContext(ctx).Create(org).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
	})

	path := "/v1/orgs/" + org.ID + "/admin-debug-mode"

	// Enable debug mode
	rr := s.makeRequest(http.MethodPost, path, DebugModeRequest{DebugMode: true})
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var dbOrg app.Org
	err = s.service.DB.Where("id = ?", org.ID).First(&dbOrg).Error
	require.NoError(s.T(), err)
	assert.True(s.T(), dbOrg.DebugMode, "debug mode should be enabled after first request")

	// Disable debug mode
	rr = s.makeRequest(http.MethodPost, path, DebugModeRequest{DebugMode: false})
	require.Equal(s.T(), http.StatusOK, rr.Code)

	err = s.service.DB.Where("id = ?", org.ID).First(&dbOrg).Error
	require.NoError(s.T(), err)
	assert.False(s.T(), dbOrg.DebugMode, "debug mode should be disabled after second request")

	// Enable again
	rr = s.makeRequest(http.MethodPost, path, DebugModeRequest{DebugMode: true})
	require.Equal(s.T(), http.StatusOK, rr.Code)

	err = s.service.DB.Where("id = ?", org.ID).First(&dbOrg).Error
	require.NoError(s.T(), err)
	assert.True(s.T(), dbOrg.DebugMode, "debug mode should be enabled after third request")
}
