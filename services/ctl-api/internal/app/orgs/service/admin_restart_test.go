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
	orgrestart "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/restart"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// RestartOrgTestService holds all fx-injected dependencies for restart org tests.
type RestartOrgTestService struct {
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

// RestartOrgTestSuite is the testify suite for restart org endpoint.
type RestartOrgTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     RestartOrgTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestRestartOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(RestartOrgTestSuite))
}

func (s *RestartOrgTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

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

func (s *RestartOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test

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

func (s *RestartOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *RestartOrgTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *RestartOrgTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *RestartOrgTestSuite) TestRestartOrg() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Org
		requestBody    interface{}
		expectedStatus int
		validateSignal bool
	}{
		{
			name: "successfully sends restart signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "restart-org",
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
			requestBody:    RestartOrgRequest{},
			expectedStatus: http.StatusOK,
			validateSignal: true,
		},
		{
			name: "handles empty request body",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "empty-request-org",
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
			requestBody:    RestartOrgRequest{},
			expectedStatus: http.StatusOK,
			validateSignal: true,
		},
		{
			name: "returns true on success",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "success-org",
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
			requestBody:    RestartOrgRequest{},
			expectedStatus: http.StatusOK,
			validateSignal: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Reset mock before test

			// Make request
			rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-restart", org.ID), tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// For successful requests, validate response is true
			if tc.expectedStatus == http.StatusOK {
				var response bool
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)
				assert.True(s.T(), response, "response should be true")
			}

			// Validate signal was sent
			if tc.validateSignal {
				signals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), signals, 1, "expected exactly one signal to be sent")

				signals = tests.GetQueueSignals(s.T(), s.service.DB)
				signal := signals[0]
				assert.Equal(s.T(), org.ID, signal.OwnerID, "signal should be sent to correct org ID")

				// Type assert to get the actual signal
				_ = signal // type check

				assert.Equal(s.T(), orgrestart.SignalType, signal.Type, "signal type should be OperationRestart")
			}
		})
	}
}

func (s *RestartOrgTestSuite) TestRestartOrgErrors() {
	testCases := []struct {
		name             string
		setupFunc        func() string // Returns org ID to use
		requestBody      interface{}
		expectedStatus   int
		shouldSendSignal bool
	}{
		{
			name: "returns error when org_id not found",
			setupFunc: func() string {
				// Return non-existent org ID
				return domains.NewOrgID()
			},
			requestBody:      RestartOrgRequest{},
			expectedStatus:   http.StatusNotFound,
			shouldSendSignal: false,
		},
		{
			name: "handles invalid JSON",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "invalid-json-org",
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

				return org.ID
			},
			requestBody:      "invalid json",
			expectedStatus:   http.StatusBadRequest,
			shouldSendSignal: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			orgID := tc.setupFunc()

			// Reset mock before test

			// Make request with invalid JSON if needed
			var rr *httptest.ResponseRecorder
			if tc.name == "handles invalid JSON" {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-restart", orgID), bytes.NewBufferString("invalid json"))
				require.NoError(s.T(), err)
				req.Header.Set("Content-Type", "application/json")
				rr = httptest.NewRecorder()
				s.router.ServeHTTP(rr, req)
			} else {
				rr = s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-restart", orgID), tc.requestBody)
			}

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate no signal was sent for error cases
			signals := tests.GetQueueSignals(s.T(), s.service.DB)
			if tc.shouldSendSignal {
				assert.Greater(s.T(), len(signals), 0, "expected signal to be sent")
			} else {
				assert.Len(s.T(), signals, 0, "no signal should be sent on error")
			}
		})
	}
}

func (s *RestartOrgTestSuite) TestRestartOrgSignalDetails() {
	s.Run("verifies signal type is OperationRestart", func() {
		ctx := context.Background()
		ctx = cctx.SetAccountContext(ctx, s.testAcc)

		org := &app.Org{
			ID:          domains.NewOrgID(),
			Name:        "signal-details-org",
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

		// Make request
		rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-restart", org.ID), RestartOrgRequest{})

		require.Equal(s.T(), http.StatusOK, rr.Code)

		// Validate signal details
		signals := tests.GetQueueSignals(s.T(), s.service.DB)
		require.Len(s.T(), signals, 1, "expected exactly one signal")

		signal := signals[0]
		assert.Equal(s.T(), org.ID, signal.OwnerID, "signal ID should match org ID")
		assert.Equal(s.T(), orgrestart.SignalType, signal.Type, "signal type must be restart")
	})
}
