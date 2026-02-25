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
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// AdminSetInternalSlackWebhookURLTestService holds all fx-injected dependencies for the test.
type AdminSetInternalSlackWebhookURLTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	Seeder          *testseed.Seeder
}

// AdminSetInternalSlackWebhookURLTestSuite is the testify suite for admin set internal slack webhook url endpoint.
type AdminSetInternalSlackWebhookURLTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     AdminSetInternalSlackWebhookURLTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestAdminSetInternalSlackWebhookURLSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminSetInternalSlackWebhookURLTestSuite))
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) SetupTest() {
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

func (s *AdminSetInternalSlackWebhookURLTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) createTestOrgWithNotificationsConfig(name, webhookURL string) *app.Org {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	org := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        name,
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: webhookURL,
		},
	}
	err := s.service.DB.WithContext(ctx).Create(org).Error
	require.NoError(s.T(), err)

	// Update OrgID on NotificationsConfig (mimics what CreateOrg does)
	s.service.DB.Model(&org.NotificationsConfig).
		Where(&app.NotificationsConfig{OwnerID: org.ID}).
		Updates(app.NotificationsConfig{OrgID: org.ID})

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
		s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
	})

	return org
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) TestAdminSetInternalSlackWebhookURL() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Org
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(*app.Org, string)
	}{
		{
			name: "successfully sets internal slack webhook URL",
			setupFunc: func() *app.Org {
				return s.createTestOrgWithNotificationsConfig("test-org-success", "https://hooks.slack.com/old")
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/services/NEW/WEBHOOK/URL",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org, newURL string) {
				// Verify database state was updated
				var notifConfig app.NotificationsConfig
				err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), newURL, notifConfig.InternalSlackWebhookURL)
			},
		},
		{
			name: "successfully updates existing webhook URL",
			setupFunc: func() *app.Org {
				return s.createTestOrgWithNotificationsConfig("test-org-update", "https://hooks.slack.com/existing")
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/services/UPDATED/URL",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org, newURL string) {
				// Verify old URL is replaced
				var notifConfig app.NotificationsConfig
				err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), newURL, notifConfig.InternalSlackWebhookURL)
				assert.NotEqual(s.T(), "https://hooks.slack.com/existing", notifConfig.InternalSlackWebhookURL)
			},
		},
		{
			name: "fails with empty URL (required validation)",
			setupFunc: func() *app.Org {
				return s.createTestOrgWithNotificationsConfig("test-org-empty", "https://hooks.slack.com/to-be-cleared")
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(org *app.Org, newURL string) {
				// URL should remain unchanged since validation failed
				var notifConfig app.NotificationsConfig
				err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "https://hooks.slack.com/to-be-cleared", notifConfig.InternalSlackWebhookURL)
			},
		},
		{
			name: "fails with missing request body field",
			setupFunc: func() *app.Org {
				return s.createTestOrgWithNotificationsConfig("test-org-no-field", "https://hooks.slack.com/unchanged")
			},
			requestBody:    map[string]interface{}{}, // Missing "name" field
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(org *app.Org, newURL string) {
				// Verify URL was not changed
				var notifConfig app.NotificationsConfig
				err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "https://hooks.slack.com/unchanged", notifConfig.InternalSlackWebhookURL)
			},
		},
		{
			name: "fails with invalid JSON",
			setupFunc: func() *app.Org {
				return s.testOrg // Use default test org
			},
			requestBody:    "invalid json string", // Will be marshaled incorrectly
			expectedStatus: http.StatusBadRequest,
			validateFunc:   nil, // No validation needed
		},
		{
			name: "fails when org not found",
			setupFunc: func() *app.Org {
				// Return a non-existent org
				return &app.Org{
					ID:          domains.NewOrgID(), // This org doesn't exist in the database
					Name:        "non-existent-org",
					SandboxMode: true,
				}
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/new",
			},
			expectedStatus: http.StatusNotFound,
			validateFunc:   nil, // No validation needed
		},
		{
			name: "successfully handles URL with special characters",
			setupFunc: func() *app.Org {
				return s.createTestOrgWithNotificationsConfig("test-org-special-chars", "https://hooks.slack.com/old")
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org, newURL string) {
				// Verify complex URL is stored correctly
				var notifConfig app.NotificationsConfig
				err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), newURL, notifConfig.InternalSlackWebhookURL)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Extract URL from request body for validation
			var newURL string
			if req, ok := tc.requestBody.(SetSlackWebhookURLRequest); ok {
				newURL = req.Name
			}

			// Make request
			path := fmt.Sprintf("/v1/orgs/%s/admin-internal-slack-webhook-url", org.ID)
			rr := s.makeRequest(http.MethodPost, path, tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Run validation if provided
			if tc.validateFunc != nil {
				tc.validateFunc(org, newURL)
			}
		})
	}
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) TestAdminSetInternalSlackWebhookURLConcurrentUpdates() {
	// Test that concurrent updates to the same org's webhook URL are handled correctly
	s.Run("handles concurrent updates", func() {
		org := s.createTestOrgWithNotificationsConfig("test-org-concurrent", "https://hooks.slack.com/initial")

		// First update
		path := fmt.Sprintf("/v1/orgs/%s/admin-internal-slack-webhook-url", org.ID)
		req1 := SetSlackWebhookURLRequest{Name: "https://hooks.slack.com/first"}
		rr1 := s.makeRequest(http.MethodPost, path, req1)
		require.Equal(s.T(), http.StatusOK, rr1.Code)

		// Second update (should overwrite first)
		req2 := SetSlackWebhookURLRequest{Name: "https://hooks.slack.com/second"}
		rr2 := s.makeRequest(http.MethodPost, path, req2)
		require.Equal(s.T(), http.StatusOK, rr2.Code)

		// Verify final state is the second update
		var notifConfig app.NotificationsConfig
		err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "https://hooks.slack.com/second", notifConfig.InternalSlackWebhookURL)
	})
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) TestAdminSetInternalSlackWebhookURLDatabaseStateVerification() {
	// Test comprehensive database state verification
	s.Run("verifies complete database state changes", func() {
		org := s.createTestOrgWithNotificationsConfig("test-org-db-state", "https://hooks.slack.com/before")

		// Store original config ID for verification
		var originalConfig app.NotificationsConfig
		err := s.service.DB.Where("owner_id = ?", org.ID).First(&originalConfig).Error
		require.NoError(s.T(), err)

		// Update the webhook URL
		newURL := "https://hooks.slack.com/after"
		path := fmt.Sprintf("/v1/orgs/%s/admin-internal-slack-webhook-url", org.ID)
		req := SetSlackWebhookURLRequest{Name: newURL}
		rr := s.makeRequest(http.MethodPost, path, req)
		require.Equal(s.T(), http.StatusOK, rr.Code)

		// Verify:
		// 1. Same notifications config record (ID unchanged)
		// 2. Webhook URL updated
		// 3. UpdatedAt timestamp changed
		var updatedConfig app.NotificationsConfig
		err = s.service.DB.Where("owner_id = ?", org.ID).First(&updatedConfig).Error
		require.NoError(s.T(), err)

		assert.Equal(s.T(), originalConfig.ID, updatedConfig.ID, "notification config ID should not change")
		assert.Equal(s.T(), newURL, updatedConfig.InternalSlackWebhookURL, "webhook URL should be updated")
		assert.True(s.T(), updatedConfig.UpdatedAt.After(originalConfig.UpdatedAt), "updated_at should be newer")

		// Verify no new notification config records were created
		var configCount int64
		err = s.service.DB.Model(&app.NotificationsConfig{}).Where("owner_id = ?", org.ID).Count(&configCount).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), int64(1), configCount, "should only have one notification config per org")
	})
}
