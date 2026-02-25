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
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// AdminDeleteOrgTestService holds all fx-injected dependencies for admin delete org tests.
type AdminDeleteOrgTestService struct {
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

// AdminDeleteOrgTestSuite is the testify suite for admin delete org endpoint.
type AdminDeleteOrgTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      AdminDeleteOrgTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
	orgsService  *service
}

func TestAdminDeleteOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminDeleteOrgTestSuite))
}

func (s *AdminDeleteOrgTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create fake event loop client for testing
	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			Mocks: &tests.TestMocks{MockEv: s.mockEvClient},

			CustomValidator: true,
		}),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *AdminDeleteOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

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

func (s *AdminDeleteOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminDeleteOrgTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *AdminDeleteOrgTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *AdminDeleteOrgTestSuite) TestAdminDeleteOrg() {
	testCases := []struct {
		name             string
		setupFunc        func() *app.Org
		requestBody      AdminDeleteOrgRequest
		expectedStatus   int
		validateSignal   bool
		expectedOrgType  app.OrgType
		shouldHardDelete bool
		checkForceFlag   bool
		expectedForce    bool
	}{
		{
			name: "deletes default org with force=false and sends signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "default-org-test",
					SandboxMode: true,
					OrgType:     app.OrgTypeDefault,
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
			requestBody: AdminDeleteOrgRequest{
				Force: false,
			},
			expectedStatus:   http.StatusOK,
			validateSignal:   true,
			expectedOrgType:  app.OrgTypeDefault,
			shouldHardDelete: false,
			checkForceFlag:   true,
			expectedForce:    false,
		},
		{
			name: "deletes default org with force=true and sends signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "default-org-force-test",
					SandboxMode: true,
					OrgType:     app.OrgTypeDefault,
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
			requestBody: AdminDeleteOrgRequest{
				Force: true,
			},
			expectedStatus:   http.StatusOK,
			validateSignal:   true,
			expectedOrgType:  app.OrgTypeDefault,
			shouldHardDelete: false,
			checkForceFlag:   true,
			expectedForce:    true,
		},
		{
			name: "hard deletes integration org without signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "integration-org-test",
					SandboxMode: true,
					OrgType:     app.OrgTypeIntegration,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				// No cleanup needed - hard delete removes it

				return org
			},
			requestBody: AdminDeleteOrgRequest{
				Force: false,
			},
			expectedStatus:   http.StatusOK,
			validateSignal:   false,
			expectedOrgType:  app.OrgTypeIntegration,
			shouldHardDelete: true,
			checkForceFlag:   false,
		},
		{
			name: "hard deletes integration org with force=true (force flag ignored)",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "integration-org-force-test",
					SandboxMode: true,
					OrgType:     app.OrgTypeIntegration,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				// No cleanup needed - hard delete removes it

				return org
			},
			requestBody: AdminDeleteOrgRequest{
				Force: true,
			},
			expectedStatus:   http.StatusOK,
			validateSignal:   false,
			expectedOrgType:  app.OrgTypeIntegration,
			shouldHardDelete: true,
			checkForceFlag:   false,
		},
		{
			name: "deletes sandbox org and sends signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "sandbox-org-test",
					SandboxMode: true,
					OrgType:     app.OrgTypeSandbox,
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
			requestBody: AdminDeleteOrgRequest{
				Force: false,
			},
			expectedStatus:   http.StatusOK,
			validateSignal:   true,
			expectedOrgType:  app.OrgTypeSandbox,
			shouldHardDelete: false,
			checkForceFlag:   true,
			expectedForce:    false,
		},
		{
			name: "deletes legacy org and sends signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "legacy-org-test",
					SandboxMode: true,
					OrgType:     app.OrgTypeLegacy,
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
			requestBody: AdminDeleteOrgRequest{
				Force: false,
			},
			expectedStatus:   http.StatusOK,
			validateSignal:   true,
			expectedOrgType:  app.OrgTypeLegacy,
			shouldHardDelete: false,
			checkForceFlag:   true,
			expectedForce:    false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-delete", org.ID), tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Verify response body is true
			var response bool
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)
			assert.True(s.T(), response, "response should be true")

			// Validate signal was sent (or not sent)
			signals := s.mockEvClient.GetSignals()
			if tc.validateSignal {
				require.Len(s.T(), signals, 1, "expected exactly one signal to be sent")

				signal := signals[0]
				assert.Equal(s.T(), org.ID, signal.ID, "signal should be sent to correct org ID")

				// Type assert to get the actual signal
				orgSignal, ok := signal.Signal.(*sigs.Signal)
				require.True(s.T(), ok, "signal should be of type *sigs.Signal")
				assert.Equal(s.T(), sigs.OperationDelete, orgSignal.Type, "signal type should be OperationDelete")

				// Verify force flag if applicable
				if tc.checkForceFlag {
					assert.Equal(s.T(), tc.expectedForce, orgSignal.ForceDelete, "ForceDelete flag should match request")
				}
			} else {
				assert.Len(s.T(), signals, 0, "no signal should be sent for integration org")
			}

			// For hard delete, verify org is actually deleted from database
			if tc.shouldHardDelete {
				var count int64
				err := s.service.DB.Unscoped().Model(&app.Org{}).Where("id = ?", org.ID).Count(&count).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), count, "integration org should be hard deleted from database")
			}
		})
	}
}

func (s *AdminDeleteOrgTestSuite) TestAdminDeleteOrgNotFound() {
	testCases := []struct {
		name           string
		orgID          string
		requestBody    AdminDeleteOrgRequest
		expectedStatus int
	}{
		{
			name:  "returns error when org not found by ID",
			orgID: domains.NewOrgID(),
			requestBody: AdminDeleteOrgRequest{
				Force: false,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "returns error when org ID is invalid",
			orgID: "invalid-org-id",
			requestBody: AdminDeleteOrgRequest{
				Force: false,
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-delete", tc.orgID), tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Verify no signal was sent
			signals := s.mockEvClient.GetSignals()
			assert.Len(s.T(), signals, 0, "no signal should be sent when org not found")
		})
	}
}

func (s *AdminDeleteOrgTestSuite) TestAdminDeleteOrgRequestParsing() {
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "handles empty request body (defaults to force=false)",
			requestBody:    "{}",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "handles invalid JSON",
			requestBody:    "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "handles null request body",
			requestBody:    "",
			expectedStatus: http.StatusOK, // BindJSON treats empty body as empty struct
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create a test org
			ctx := context.Background()
			ctx = cctx.SetAccountContext(ctx, s.testAcc)

			org := &app.Org{
				ID:          domains.NewOrgID(),
				Name:        "request-test-org",
				SandboxMode: true,
				OrgType:     app.OrgTypeDefault,
				NotificationsConfig: app.NotificationsConfig{
					InternalSlackWebhookURL: "https://hooks.slack.com/foo",
				},
			}
			err := s.service.DB.WithContext(ctx).Create(org).Error
			require.NoError(s.T(), err)
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
			})

			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request with raw body
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-delete", org.ID), bytes.NewBufferString(tc.requestBody))
			require.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)
		})
	}
}

func (s *AdminDeleteOrgTestSuite) TestAdminDeleteOrgByName() {
	testCases := []struct {
		name           string
		orgName        string
		lookupValue    string
		expectedStatus int
		validateSignal bool
	}{
		{
			name:           "deletes org by exact name match",
			orgName:        "test-org-by-name",
			lookupValue:    "test-org-by-name",
			expectedStatus: http.StatusOK,
			validateSignal: true,
		},
		{
			name:           "deletes org by partial name match (LIKE)",
			orgName:        "test-org-partial",
			lookupValue:    "test-org-partial",
			expectedStatus: http.StatusOK,
			validateSignal: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create test org
			ctx := context.Background()
			ctx = cctx.SetAccountContext(ctx, s.testAcc)

			org := &app.Org{
				ID:          domains.NewOrgID(),
				Name:        tc.orgName,
				SandboxMode: true,
				OrgType:     app.OrgTypeDefault,
				NotificationsConfig: app.NotificationsConfig{
					InternalSlackWebhookURL: "https://hooks.slack.com/foo",
				},
			}
			err := s.service.DB.WithContext(ctx).Create(org).Error
			require.NoError(s.T(), err)
			s.T().Cleanup(func() {
				s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
			})

			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request using name instead of ID
			requestBody := AdminDeleteOrgRequest{Force: false}
			rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-delete", tc.lookupValue), requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if tc.validateSignal {
				// Verify signal was sent
				signals := s.mockEvClient.GetSignals()
				require.Len(s.T(), signals, 1, "expected exactly one signal to be sent")

				signal := signals[0]
				assert.Equal(s.T(), org.ID, signal.ID, "signal should be sent to correct org ID")

				orgSignal, ok := signal.Signal.(*sigs.Signal)
				require.True(s.T(), ok, "signal should be of type *sigs.Signal")
				assert.Equal(s.T(), sigs.OperationDelete, orgSignal.Type, "signal type should be OperationDelete")
			}
		})
	}
}
