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
		ID:   domains.NewOrgID(),
		Name: "test-org",
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/initial",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
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
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-success",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/old",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
					s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
				})

				return org
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: stringPtr("https://hooks.slack.com/services/NEW/WEBHOOK/URL"),
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
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-update",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/existing",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
					s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
				})

				return org
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: stringPtr("https://hooks.slack.com/services/UPDATED/URL"),
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
			name: "successfully sets empty URL",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-empty",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/to-be-cleared",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
					s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
				})

				return org
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: stringPtr(""),
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org, newURL string) {
				// Verify URL is cleared
				var notifConfig app.NotificationsConfig
				err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "", notifConfig.InternalSlackWebhookURL)
			},
		},
		{
			name: "fails with missing request body field",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-no-field",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/unchanged",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
					s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
				})

				return org
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
					ID:   domains.NewOrgID(), // This org doesn't exist in the database
					Name: "non-existent-org",
				}
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: stringPtr("https://hooks.slack.com/new"),
			},
			expectedStatus: http.StatusNotFound,
			validateFunc:   nil, // No validation needed
		},
		{
			name: "successfully handles URL with special characters",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-special-chars",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/old",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
					s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
				})

				return org
			},
			requestBody: SetSlackWebhookURLRequest{
				Name: stringPtr("https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"),
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
				if req.Name != nil {
					newURL = *req.Name
				}
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
		ctx := context.Background()
		ctx = cctx.SetAccountContext(ctx, s.testAcc)

		org := &app.Org{
			ID:   domains.NewOrgID(),
			Name: "test-org-concurrent",
			NotificationsConfig: app.NotificationsConfig{
				InternalSlackWebhookURL: "https://hooks.slack.com/initial",
			},
		}
		err := s.service.DB.WithContext(ctx).Create(org).Error
		require.NoError(s.T(), err)
		s.T().Cleanup(func() {
			s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
			s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
		})

		// First update
		path := fmt.Sprintf("/v1/orgs/%s/admin-internal-slack-webhook-url", org.ID)
		req1 := SetSlackWebhookURLRequest{Name: stringPtr("https://hooks.slack.com/first")}
		rr1 := s.makeRequest(http.MethodPost, path, req1)
		require.Equal(s.T(), http.StatusOK, rr1.Code)

		// Second update (should overwrite first)
		req2 := SetSlackWebhookURLRequest{Name: stringPtr("https://hooks.slack.com/second")}
		rr2 := s.makeRequest(http.MethodPost, path, req2)
		require.Equal(s.T(), http.StatusOK, rr2.Code)

		// Verify final state is the second update
		var notifConfig app.NotificationsConfig
		err = s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "https://hooks.slack.com/second", notifConfig.InternalSlackWebhookURL)
	})
}

func (s *AdminSetInternalSlackWebhookURLTestSuite) TestAdminSetInternalSlackWebhookURLDatabaseStateVerification() {
	// Test comprehensive database state verification
	s.Run("verifies complete database state changes", func() {
		ctx := context.Background()
		ctx = cctx.SetAccountContext(ctx, s.testAcc)

		org := &app.Org{
			ID:   domains.NewOrgID(),
			Name: "test-org-db-state",
			NotificationsConfig: app.NotificationsConfig{
				InternalSlackWebhookURL: "https://hooks.slack.com/before",
			},
		}
		err := s.service.DB.WithContext(ctx).Create(org).Error
		require.NoError(s.T(), err)
		s.T().Cleanup(func() {
			s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
			s.service.DB.Unscoped().Delete(&app.NotificationsConfig{}, "owner_id = ?", org.ID)
		})

		// Store original config ID for verification
		var originalConfig app.NotificationsConfig
		err = s.service.DB.Where("owner_id = ?", org.ID).First(&originalConfig).Error
		require.NoError(s.T(), err)

		// Update the webhook URL
		newURL := "https://hooks.slack.com/after"
		path := fmt.Sprintf("/v1/orgs/%s/admin-internal-slack-webhook-url", org.ID)
		req := SetSlackWebhookURLRequest{Name: stringPtr(newURL)}
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
