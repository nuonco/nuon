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
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// AdminDeprovisionOrgTestService holds all fx-injected dependencies for admin deprovision org tests.
type AdminDeprovisionOrgTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
	Seeder          *testseed.Seeder
}

// AdminDeprovisionOrgTestSuite is the testify suite for AdminDeprovisionOrg endpoint.
type AdminDeprovisionOrgTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      AdminDeprovisionOrgTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
}

func TestAdminDeprovisionOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminDeprovisionOrgTestSuite))
}

func (s *AdminDeprovisionOrgTestSuite) SetupSuite() {
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
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *AdminDeprovisionOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

	// Create test router with standard middlewares
	// AdminDeprovisionOrg is an admin endpoint, no org context needed
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminDeprovisionOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminDeprovisionOrgTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *AdminDeprovisionOrgTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminDeprovisionOrgTestSuite) TestAdminDeprovisionOrg() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Org
		requestBody    interface{}
		expectedCode   int
		validateSignal bool
		expectedType   eventloop.SignalType
	}{
		{
			name: "successfully sends deprovision signal with force=false",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "deprovision-org-normal",
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
			requestBody: AdminDeprovisionOrgRequest{
				Force: false,
			},
			expectedCode:   http.StatusOK,
			validateSignal: true,
			expectedType:   sigs.OperationDeprovision,
		},
		{
			name: "successfully sends force deprovision signal with force=true",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "deprovision-org-force",
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
			requestBody: AdminDeprovisionOrgRequest{
				Force: true,
			},
			expectedCode:   http.StatusOK,
			validateSignal: true,
			expectedType:   sigs.OperationForceDeprovision,
		},
		{
			name: "defaults to force=false when field missing",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "deprovision-org-default",
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
			requestBody:    map[string]interface{}{}, // Empty JSON object
			expectedCode:   http.StatusOK,
			validateSignal: true,
			expectedType:   sigs.OperationDeprovision,
		},
		{
			name: "returns error when org_id not found",
			setupFunc: func() *app.Org {
				// Return non-existent org
				return &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "non-existent-org",
					SandboxMode: true,
				}
			},
			requestBody: AdminDeprovisionOrgRequest{
				Force: false,
			},
			expectedCode:   http.StatusNotFound,
			validateSignal: false,
		},
		{
			name: "handles invalid JSON request body",
			setupFunc: func() *app.Org {
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

				return org
			},
			requestBody:    "invalid json",
			expectedCode:   http.StatusBadRequest,
			validateSignal: false,
		},
		{
			name: "returns true on success",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "returns-true-org",
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
			requestBody: AdminDeprovisionOrgRequest{
				Force: false,
			},
			expectedCode:   http.StatusOK,
			validateSignal: true,
			expectedType:   sigs.OperationDeprovision,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			path := fmt.Sprintf("/v1/orgs/%s/admin-deprovision", org.ID)
			rr := s.makeRequest(http.MethodPost, path, tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// Validate signal was sent (or not sent)
			signals := s.mockEvClient.GetSignals()
			if tc.validateSignal {
				require.Len(s.T(), signals, 1, "expected exactly one signal to be sent")

				signal := signals[0]
				assert.Equal(s.T(), org.ID, signal.ID, "signal should be sent to correct org ID")

				// Type assert to get the actual signal
				orgSignal, ok := signal.Signal.(*sigs.Signal)
				require.True(s.T(), ok, "signal should be of type *sigs.Signal")
				assert.Equal(s.T(), tc.expectedType, orgSignal.Type, "signal type should match expected")

				// Parse response body to verify it returns true
				if rr.Code == http.StatusOK {
					var result bool
					err := json.Unmarshal(rr.Body.Bytes(), &result)
					if err != nil {
						s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
					}
					require.NoError(s.T(), err)
					assert.True(s.T(), result, "endpoint should return true on success")
				}
			} else {
				assert.Len(s.T(), signals, 0, "no signal should be sent for error cases")
			}
		})
	}
}

func (s *AdminDeprovisionOrgTestSuite) TestAdminDeprovisionOrgSignalTypes() {
	testCases := []struct {
		name         string
		force        bool
		expectedType eventloop.SignalType
	}{
		{
			name:         "force=false sends OperationDeprovision",
			force:        false,
			expectedType: sigs.OperationDeprovision,
		},
		{
			name:         "force=true sends OperationForceDeprovision",
			force:        true,
			expectedType: sigs.OperationForceDeprovision,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create org for this test
			ctx := context.Background()
			ctx = cctx.SetAccountContext(ctx, s.testAcc)

			org := &app.Org{
				ID:          domains.NewOrgID(),
				Name:        fmt.Sprintf("signal-test-org-%v", tc.force),
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

			// Reset mock
			s.mockEvClient.Reset()

			// Make request
			path := fmt.Sprintf("/v1/orgs/%s/admin-deprovision", org.ID)
			req := AdminDeprovisionOrgRequest{Force: tc.force}
			rr := s.makeRequest(http.MethodPost, path, req)

			require.Equal(s.T(), http.StatusOK, rr.Code)

			// Validate signal type
			signals := s.mockEvClient.GetSignals()
			require.Len(s.T(), signals, 1)

			orgSignal, ok := signals[0].Signal.(*sigs.Signal)
			require.True(s.T(), ok)
			assert.Equal(s.T(), tc.expectedType, orgSignal.Type)
		})
	}
}

func (s *AdminDeprovisionOrgTestSuite) TestAdminDeprovisionOrgDoesNotModifyDatabase() {
	// Create org
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	org := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "db-unchanged-org",
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

	// Store original state
	originalCreatedAt := org.CreatedAt
	originalUpdatedAt := org.UpdatedAt

	// Reset mock
	s.mockEvClient.Reset()

	// Make deprovision request
	path := fmt.Sprintf("/v1/orgs/%s/admin-deprovision", org.ID)
	req := AdminDeprovisionOrgRequest{Force: false}
	rr := s.makeRequest(http.MethodPost, path, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Verify org still exists and unchanged in database
	var orgAfter app.Org
	err = s.service.DB.First(&orgAfter, "id = ?", org.ID).Error
	require.NoError(s.T(), err)

	assert.Equal(s.T(), org.ID, orgAfter.ID)
	assert.Equal(s.T(), org.Name, orgAfter.Name)
	assert.Equal(s.T(), org.OrgType, orgAfter.OrgType)
	assert.Equal(s.T(), originalCreatedAt.Unix(), orgAfter.CreatedAt.Unix())
	assert.Equal(s.T(), originalUpdatedAt.Unix(), orgAfter.UpdatedAt.Unix())
	assert.Equal(s.T(), soft_delete.DeletedAt(0), orgAfter.DeletedAt, "org should not be soft-deleted")
}
