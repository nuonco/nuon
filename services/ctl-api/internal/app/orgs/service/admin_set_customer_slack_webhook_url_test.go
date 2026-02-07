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

// AdminSetCustomerSlackWebhookURLTestService holds all fx-injected dependencies for admin set customer slack webhook url tests.
type AdminSetCustomerSlackWebhookURLTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
}

// AdminSetCustomerSlackWebhookURLTestSuite is the testify suite for admin set customer slack webhook url endpoint.
type AdminSetCustomerSlackWebhookURLTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     AdminSetCustomerSlackWebhookURLTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestAdminSetCustomerSlackWebhookURLSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminSetCustomerSlackWebhookURLTestSuite))
}

func (s *AdminSetCustomerSlackWebhookURLTestSuite) SetupSuite() {
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

func (s *AdminSetCustomerSlackWebhookURLTestSuite) SetupTest() {
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

func (s *AdminSetCustomerSlackWebhookURLTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminSetCustomerSlackWebhookURLTestSuite) setupTestData() {
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
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *AdminSetCustomerSlackWebhookURLTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminSetCustomerSlackWebhookURLTestSuite) TestAdminSetCustomerSlackWebhookURL() {
	testCases := []struct {
		name                 string
		setupFunc            func() *app.Org
		requestBody          interface{}
		expectedStatus       int
		validateFunc         func(*app.Org)
		expectDatabaseUpdate bool
		expectedWebhookURL   string
	}{
		{
			name: "successfully sets customer slack webhook URL",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-set-webhook",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/internal",
						SlackWebhookURL:         "",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: SetCustomerSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
			},
			expectedStatus:       http.StatusOK,
			expectDatabaseUpdate: true,
			expectedWebhookURL:   "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX", org.NotificationsConfig.SlackWebhookURL)
			},
		},
		{
			name: "updates existing customer slack webhook URL",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-update-webhook",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/internal",
						SlackWebhookURL:         "https://hooks.slack.com/old-webhook",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: SetCustomerSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/new-webhook",
			},
			expectedStatus:       http.StatusOK,
			expectDatabaseUpdate: true,
			expectedWebhookURL:   "https://hooks.slack.com/new-webhook",
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "https://hooks.slack.com/new-webhook", org.NotificationsConfig.SlackWebhookURL)
			},
		},
		{
			name: "fails with empty webhook URL (required validation)",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-empty-webhook",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/internal",
						SlackWebhookURL:         "https://hooks.slack.com/existing",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: SetCustomerSlackWebhookURLRequest{
				Name: "",
			},
			expectedStatus:       http.StatusBadRequest,
			expectDatabaseUpdate: false,
		},
		{
			name: "fails with missing name field",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-missing-field",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/internal",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody:          map[string]interface{}{},
			expectedStatus:       http.StatusBadRequest,
			expectDatabaseUpdate: false,
		},
		{
			name: "handles invalid JSON",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-invalid-json",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/internal",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody:          nil,
			expectedStatus:       http.StatusBadRequest,
			expectDatabaseUpdate: false,
		},
		{
			name: "handles org not found error",
			setupFunc: func() *app.Org {
				// Create org but use non-existent ID for request
				return &app.Org{
					ID:   domains.NewOrgID(), // Non-existent ID
					Name: "nonexistent-org",
				}
			},
			requestBody: SetCustomerSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/webhook",
			},
			expectedStatus:       http.StatusNotFound,
			expectDatabaseUpdate: false,
		},
		{
			name: "preserves internal slack webhook URL",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "test-org-preserve-internal",
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/internal-preserved",
						SlackWebhookURL:         "",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: SetCustomerSlackWebhookURLRequest{
				Name: "https://hooks.slack.com/customer",
			},
			expectedStatus:       http.StatusOK,
			expectDatabaseUpdate: true,
			expectedWebhookURL:   "https://hooks.slack.com/customer",
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), "https://hooks.slack.com/customer", org.NotificationsConfig.SlackWebhookURL)
				require.Equal(s.T(), "https://hooks.slack.com/internal-preserved", org.NotificationsConfig.InternalSlackWebhookURL)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Update router context to use the test org
			s.router = tests.NewTestRouter(tests.RouterOptions{
				L:       s.service.L,
				DB:      s.service.DB,
				TestOrg: org,
				TestAcc: s.testAcc,
			})
			err := s.orgsService.RegisterInternalRoutes(s.router)
			require.NoError(s.T(), err)

			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/orgs/"+org.ID+"/admin-customer-slack-webhook-url", tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Verify database state if update was expected
			if tc.expectDatabaseUpdate {
				// Parse response
				var response bool
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)
				require.True(s.T(), response)

				// Fetch org from database to verify update
				var updatedOrg app.Org
				err = s.service.DB.
					Preload("NotificationsConfig").
					Where("id = ?", org.ID).
					First(&updatedOrg).Error
				require.NoError(s.T(), err)

				// Verify webhook URL was updated
				require.Equal(s.T(), tc.expectedWebhookURL, updatedOrg.NotificationsConfig.SlackWebhookURL)

				// Run additional validations if provided
				if tc.validateFunc != nil {
					tc.validateFunc(&updatedOrg)
				}
			}
		})
	}
}

func (s *AdminSetCustomerSlackWebhookURLTestSuite) TestAdminSetCustomerSlackWebhookURLVerifyNotificationsConfigUpdate() {
	// This test specifically verifies that only the SlackWebhookURL field is updated
	// and that the update happens on the NotificationsConfig table
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	org := &app.Org{
		ID:   domains.NewOrgID(),
		Name: "test-org-verify-update",
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL:  "https://hooks.slack.com/internal",
			SlackWebhookURL:          "https://hooks.slack.com/old",
			EnableSlackNotifications: true,
			EnableEmailNotifications: false,
		},
	}
	err := s.service.DB.WithContext(ctx).Create(org).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
	})

	// Store original values
	originalInternalURL := org.NotificationsConfig.InternalSlackWebhookURL
	originalEnableSlack := org.NotificationsConfig.EnableSlackNotifications
	originalEnableEmail := org.NotificationsConfig.EnableEmailNotifications

	// Update router context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: org,
		TestAcc: s.testAcc,
	})
	err = s.orgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)

	// Make request
	req := SetCustomerSlackWebhookURLRequest{
		Name: "https://hooks.slack.com/new-customer-webhook",
	}
	rr := s.makeRequest(http.MethodPost, "/v1/orgs/"+org.ID+"/admin-customer-slack-webhook-url", req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Fetch updated org
	var updatedOrg app.Org
	err = s.service.DB.
		Preload("NotificationsConfig").
		Where("id = ?", org.ID).
		First(&updatedOrg).Error
	require.NoError(s.T(), err)

	// Verify ONLY SlackWebhookURL was updated
	require.Equal(s.T(), "https://hooks.slack.com/new-customer-webhook", updatedOrg.NotificationsConfig.SlackWebhookURL)

	// Verify other fields remain unchanged
	require.Equal(s.T(), originalInternalURL, updatedOrg.NotificationsConfig.InternalSlackWebhookURL)
	require.Equal(s.T(), originalEnableSlack, updatedOrg.NotificationsConfig.EnableSlackNotifications)
	require.Equal(s.T(), originalEnableEmail, updatedOrg.NotificationsConfig.EnableEmailNotifications)
}
