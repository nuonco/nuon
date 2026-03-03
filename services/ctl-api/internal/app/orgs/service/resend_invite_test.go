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
)

// ResendOrgInviteTestService holds all fx-injected dependencies for resend org invite tests.
type ResendOrgInviteTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
}

// ResendOrgInviteTestSuite is the testify suite for resend org invite endpoint.
type ResendOrgInviteTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      ResendOrgInviteTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
	orgsService  *service
}

func TestResendOrgInviteSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(ResendOrgInviteTestSuite))
}

func (s *ResendOrgInviteTestSuite) SetupSuite() {
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

func (s *ResendOrgInviteTestSuite) SetupTest() {
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

func (s *ResendOrgInviteTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *ResendOrgInviteTestSuite) setupTestData() {
	// Create test account
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

	// Create test org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("test-org-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *ResendOrgInviteTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(bodyBytes)
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

func (s *ResendOrgInviteTestSuite) TestResendOrgInvite() {
	testCases := []struct {
		name             string
		setupFunc        func() string // Returns invite ID to use in request
		expectedStatus   int
		validateResponse func(*httptest.ResponseRecorder)
		validateSignal   bool
	}{
		{
			name: "successfully resends pending invite",
			setupFunc: func() string {
				// Create a pending invite
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				testEmail := fmt.Sprintf("resend-success-%s@test.nuon.co", domains.NewAccountID()[:8])
				invite := &app.OrgInvite{
					Email:    testEmail,
					OrgID:    s.testOrg.ID,
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeOrgAdmin,
				}
				err := s.service.DB.WithContext(ctx).Create(invite).Error
				require.NoError(s.T(), err)

				// Cleanup
				inviteID := invite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
				})

				return invite.ID
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), invite.ID)
				assert.NotEmpty(s.T(), invite.Email)
				assert.Equal(s.T(), app.OrgInviteStatusPending, invite.Status)
				assert.Equal(s.T(), s.testOrg.ID, invite.OrgID)
			},
			validateSignal: true,
		},
		{
			name: "returns error for non-existent invite ID",
			setupFunc: func() string {
				// Return a fake invite ID that doesn't exist
				return domains.NewOrgID()
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "record not found")
			},
			validateSignal: false,
		},
		{
			name: "returns 400 for already accepted invite",
			setupFunc: func() string {
				// Create an accepted invite
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				testEmail := fmt.Sprintf("already-accepted-%s@test.nuon.co", domains.NewAccountID()[:8])
				invite := &app.OrgInvite{
					Email:    testEmail,
					OrgID:    s.testOrg.ID,
					Status:   app.OrgInviteStatusAccepted,
					RoleType: app.RoleTypeOrgAdmin,
				}
				err := s.service.DB.WithContext(ctx).Create(invite).Error
				require.NoError(s.T(), err)

				// Cleanup
				inviteID := invite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
				})

				return invite.ID
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invite has already been accepted")
			},
			validateSignal: false,
		},
		{
			name: "returns error for invite belonging to different org",
			setupFunc: func() string {
				// Create second org
				acc2ID := domains.NewAccountID()
				acc2 := &app.Account{
					ID:          acc2ID,
					Email:       fmt.Sprintf("%s@test.nuon.co", acc2ID),
					Subject:     acc2ID,
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, acc2)
				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err = s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)
				})

				// Create invite in different org
				testEmail := fmt.Sprintf("different-org-%s@test.nuon.co", domains.NewAccountID()[:8])
				invite := &app.OrgInvite{
					Email:    testEmail,
					OrgID:    org2.ID,
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeOrgAdmin,
				}
				err = s.service.DB.WithContext(ctx).Create(invite).Error
				require.NoError(s.T(), err)

				// Cleanup
				inviteID := invite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
				})

				return invite.ID
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				// Should not find the invite because it belongs to different org
				assert.Contains(s.T(), body, "record not found")
			},
			validateSignal: false,
		},
		{
			name: "returns 400 for empty invite_id",
			setupFunc: func() string {
				// Return empty string to test empty param handling
				return ""
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invite_id is required")
			},
			validateSignal: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

			// Setup test data
			inviteID := tc.setupFunc()

			// Make request
			path := fmt.Sprintf("/v1/orgs/current/invites/%s/resend", inviteID)
			rr := s.makeRequest(http.MethodPost, path, nil)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate response
			if tc.validateResponse != nil {
				tc.validateResponse(rr)
			}

			// Validate signal was sent (or not sent)
			signals := s.mockEvClient.GetSignals()
			if tc.validateSignal {
				require.Len(s.T(), signals, 1, "expected exactly one signal to be sent")

				signal := signals[0]
				assert.Equal(s.T(), s.testOrg.ID, signal.ID, "signal should be sent to correct org ID")

				// Type assert to get the actual signal
				orgSignal, ok := signal.Signal.(*sigs.Signal)
				require.True(s.T(), ok, "signal should be of type *sigs.Signal")
				assert.Equal(s.T(), sigs.OperationInviteCreated, orgSignal.Type, "signal type should be OperationInviteCreated")

				// Verify InviteID is set in signal
				assert.NotEmpty(s.T(), orgSignal.InviteID, "signal should contain invite ID")

				// Verify InviteID matches the invite we're resending
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), invite.ID, orgSignal.InviteID, "signal invite ID should match resent invite")
			} else {
				assert.Len(s.T(), signals, 0, "no signal should be sent for failed operations")
			}
		})
	}
}

func (s *ResendOrgInviteTestSuite) TestResendOrgInvite_DoesNotModifyInviteInDatabase() {
	// Create a pending invite
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	testEmail := fmt.Sprintf("no-modify-%s@test.nuon.co", domains.NewAccountID()[:8])
	invite := &app.OrgInvite{
		Email:    testEmail,
		OrgID:    s.testOrg.ID,
		Status:   app.OrgInviteStatusPending,
		RoleType: app.RoleTypeOrgAdmin,
	}
	err := s.service.DB.WithContext(ctx).Create(invite).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite.ID)
	})

	// Store original timestamp
	originalUpdatedAt := invite.UpdatedAt

	// Resend the invite
	path := fmt.Sprintf("/v1/orgs/current/invites/%s/resend", invite.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Verify database record was not modified
	var dbInvite app.OrgInvite
	err = s.service.DB.Where("id = ?", invite.ID).First(&dbInvite).Error
	require.NoError(s.T(), err)

	assert.Equal(s.T(), testEmail, dbInvite.Email)
	assert.Equal(s.T(), app.OrgInviteStatusPending, dbInvite.Status)
	assert.Equal(s.T(), s.testOrg.ID, dbInvite.OrgID)
	// UpdatedAt should not change since we only read the record
	assert.Equal(s.T(), originalUpdatedAt.Unix(), dbInvite.UpdatedAt.Unix())
}

func (s *ResendOrgInviteTestSuite) TestResendOrgInvite_CanResendMultipleTimes() {
	// Create a pending invite
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	testEmail := fmt.Sprintf("multi-resend-%s@test.nuon.co", domains.NewAccountID()[:8])
	invite := &app.OrgInvite{
		Email:    testEmail,
		OrgID:    s.testOrg.ID,
		Status:   app.OrgInviteStatusPending,
		RoleType: app.RoleTypeOrgAdmin,
	}
	err := s.service.DB.WithContext(ctx).Create(invite).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite.ID)
	})

	path := fmt.Sprintf("/v1/orgs/current/invites/%s/resend", invite.ID)

	// Resend first time
	s.mockEvClient.Reset()
	rr1 := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusOK, rr1.Code)
	signals1 := s.mockEvClient.GetSignals()
	require.Len(s.T(), signals1, 1, "first resend should send signal")

	// Resend second time
	s.mockEvClient.Reset()
	rr2 := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusOK, rr2.Code)
	signals2 := s.mockEvClient.GetSignals()
	require.Len(s.T(), signals2, 1, "second resend should send signal")

	// Resend third time
	s.mockEvClient.Reset()
	rr3 := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusOK, rr3.Code)
	signals3 := s.mockEvClient.GetSignals()
	require.Len(s.T(), signals3, 1, "third resend should send signal")

	// Verify all responses return the same invite
	var invite1, invite2, invite3 app.OrgInvite
	err = json.Unmarshal(rr1.Body.Bytes(), &invite1)
	require.NoError(s.T(), err)
	err = json.Unmarshal(rr2.Body.Bytes(), &invite2)
	require.NoError(s.T(), err)
	err = json.Unmarshal(rr3.Body.Bytes(), &invite3)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), invite.ID, invite1.ID)
	assert.Equal(s.T(), invite.ID, invite2.ID)
	assert.Equal(s.T(), invite.ID, invite3.ID)
}
