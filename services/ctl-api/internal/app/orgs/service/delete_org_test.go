package service

import (
	"context"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// DeleteOrgTestService holds all fx-injected dependencies for delete org tests.
type DeleteOrgTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
}

// DeleteOrgTestSuite is the testify suite for delete org endpoint.
type DeleteOrgTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      DeleteOrgTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.FakeEventLoopClient
	orgsService  *service
}

func TestDeleteOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(DeleteOrgTestSuite))
}

func (s *DeleteOrgTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	// Create fake event loop client for testing
	s.mockEvClient = tests.NewFakeEventLoopClient()

	options := append(
		tests.CtlApiFXOptions(),
		// Override eventloop.Client with mock
		fx.Decorate(func() eventloop.Client {
			return s.mockEvClient
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

func (s *DeleteOrgTestSuite) SetupTest() {
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

	err := s.orgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeleteOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeleteOrgTestSuite) setupTestData() {
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

func (s *DeleteOrgTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeleteOrgTestSuite) TestDeleteOrg() {
	testCases := []struct {
		name             string
		setupFunc        func() *app.Org
		expectedStatus   int
		validateSignal   bool
		expectedOrgType  app.OrgType
		shouldHardDelete bool
	}{
		{
			name: "deletes default org and sends signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:      domains.NewOrgID(),
					Name:    "default-org",
					OrgType: app.OrgTypeDefault,
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
			expectedStatus:   http.StatusOK,
			validateSignal:   true,
			expectedOrgType:  app.OrgTypeDefault,
			shouldHardDelete: false,
		},
		{
			name: "hard deletes integration org without signal",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:      domains.NewOrgID(),
					Name:    "integration-org",
					OrgType: app.OrgTypeIntegration,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				// No cleanup needed - hard delete removes it

				return org
			},
			expectedStatus:   http.StatusOK,
			validateSignal:   false,
			expectedOrgType:  app.OrgTypeIntegration,
			shouldHardDelete: true,
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
			err := s.orgsService.RegisterPublicRoutes(s.router)
			require.NoError(s.T(), err)

			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			rr := s.makeRequest(http.MethodDelete, "/v1/orgs/current")

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

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
				assert.False(s.T(), orgSignal.ForceDelete, "ForceDelete should be false")
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
