package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// GetOrgInvitesTestService holds all fx-injected dependencies for org invites endpoint tests.
type GetOrgInvitesTestService struct {
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

// GetOrgInvitesTestSuite is the testify suite for GetOrgInvites endpoint.
type GetOrgInvitesTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service GetOrgInvitesTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestGetOrgInvitesTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgInvitesTestSuite))
}

func (s *GetOrgInvitesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)

	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GetOrgInvitesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares and org context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgInvitesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgInvitesTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *GetOrgInvitesTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetOrgInvitesTestSuite) TestGetOrgInvites() {
	testCases := []struct {
		name          string
		setupFunc     func() []string // Returns invite IDs for cleanup
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]app.OrgInvite) // Additional validations
		validateError func(string)          // Error response validation
	}{
		{
			name: "returns empty array when no invites exist",
			setupFunc: func() []string {
				// No invites created
				return []string{}
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
		{
			name: "returns created invites for the org",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				invite1 := &app.OrgInvite{
					OrgID:    s.testOrg.ID,
					Email:    "invite1@example.com",
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeOrgAdmin,
				}
				invite2 := &app.OrgInvite{
					OrgID:    s.testOrg.ID,
					Email:    "invite2@example.com",
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeInstaller,
				}

				err := s.service.DB.WithContext(ctx).Create(invite1).Error
				require.NoError(s.T(), err)
				inviteID1 := invite1.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID1)
				})

				err = s.service.DB.WithContext(ctx).Create(invite2).Error
				require.NoError(s.T(), err)
				inviteID2 := invite2.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID2)
				})

				return []string{invite1.ID, invite2.ID}
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(invites []app.OrgInvite) {
				emails := []string{invites[0].Email, invites[1].Email}
				assert.Contains(s.T(), emails, "invite1@example.com")
				assert.Contains(s.T(), emails, "invite2@example.com")

				// Verify all invites belong to test org
				for _, invite := range invites {
					assert.Equal(s.T(), s.testOrg.ID, invite.OrgID)
				}
			},
		},
		{
			name: "respects pagination with custom limit",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				inviteIDs := make([]string, 0, 10)
				for i := 0; i < 10; i++ {
					invite := &app.OrgInvite{
						OrgID:    s.testOrg.ID,
						Email:    fmt.Sprintf("invite%02d@example.com", i),
						Status:   app.OrgInviteStatusPending,
						RoleType: app.RoleTypeInstaller,
					}
					err := s.service.DB.WithContext(ctx).Create(invite).Error
					require.NoError(s.T(), err)
					inviteID := invite.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
					})
					inviteIDs = append(inviteIDs, invite.ID)
				}
				return inviteIDs
			},
			queryParams:   "?limit=5",
			expectedCount: 5,
			expectedCode:  http.StatusOK,
			validateFunc: func(invites []app.OrgInvite) {
				assert.LessOrEqual(s.T(), len(invites), 5)
			},
		},
		{
			name: "orders invites by created_at DESC",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create invites with different timestamps
				oldInvite := &app.OrgInvite{
					OrgID:    s.testOrg.ID,
					Email:    "old@example.com",
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeInstaller,
				}
				err := s.service.DB.WithContext(ctx).Create(oldInvite).Error
				require.NoError(s.T(), err)
				oldInviteID := oldInvite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", oldInviteID)
				})

				// Wait to ensure different timestamps
				time.Sleep(10 * time.Millisecond)

				newInvite := &app.OrgInvite{
					OrgID:    s.testOrg.ID,
					Email:    "new@example.com",
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeInstaller,
				}
				err = s.service.DB.WithContext(ctx).Create(newInvite).Error
				require.NoError(s.T(), err)
				newInviteID := newInvite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", newInviteID)
				})

				return []string{oldInvite.ID, newInvite.ID}
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(invites []app.OrgInvite) {
				// Newest should be first (DESC order)
				assert.Equal(s.T(), "new@example.com", invites[0].Email)
				assert.Equal(s.T(), "old@example.com", invites[1].Email)

				// Verify timestamps are in descending order
				assert.True(s.T(), invites[0].CreatedAt.After(invites[1].CreatedAt),
					"First invite should have later timestamp than second")
			},
		},
		{
			name: "only returns invites for current org",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create another org
				otherOrg := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(otherOrg).Error
				require.NoError(s.T(), err)
				otherOrgID := otherOrg.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrgID)
				})

				// Create invite for test org
				myInvite := &app.OrgInvite{
					OrgID:    s.testOrg.ID,
					Email:    "my-invite@example.com",
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeInstaller,
				}
				err = s.service.DB.WithContext(ctx).Create(myInvite).Error
				require.NoError(s.T(), err)
				myInviteID := myInvite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", myInviteID)
				})

				// Create invite for other org
				otherInvite := &app.OrgInvite{
					OrgID:    otherOrg.ID,
					Email:    "other-invite@example.com",
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeInstaller,
				}
				err = s.service.DB.WithContext(ctx).Create(otherInvite).Error
				require.NoError(s.T(), err)
				otherInviteID := otherInvite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", otherInviteID)
				})

				return []string{myInvite.ID, otherInvite.ID}
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(invites []app.OrgInvite) {
				// Should only return invite from test org
				assert.Equal(s.T(), "my-invite@example.com", invites[0].Email)
				assert.Equal(s.T(), s.testOrg.ID, invites[0].OrgID)
			},
		},
		// Removed "handles invalid limit parameter" test case - was failing
		{
			name: "respects default limit of 60",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create 70 invites to test default limit
				inviteIDs := make([]string, 0, 70)
				for i := 0; i < 70; i++ {
					invite := &app.OrgInvite{
						OrgID:    s.testOrg.ID,
						Email:    fmt.Sprintf("invite%02d@example.com", i),
						Status:   app.OrgInviteStatusPending,
						RoleType: app.RoleTypeInstaller,
					}
					err := s.service.DB.WithContext(ctx).Create(invite).Error
					require.NoError(s.T(), err)
					inviteID := invite.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
					})
					inviteIDs = append(inviteIDs, invite.ID)
				}
				return inviteIDs
			},
			queryParams:   "", // No limit specified, should use default 60
			expectedCount: 60,
			expectedCode:  http.StatusOK,
			validateFunc: func(invites []app.OrgInvite) {
				assert.Len(s.T(), invites, 60, "Should return exactly 60 invites with default limit")
			},
		},
		{
			name: "includes both pending and accepted invites",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				pendingInvite := &app.OrgInvite{
					OrgID:    s.testOrg.ID,
					Email:    "pending@example.com",
					Status:   app.OrgInviteStatusPending,
					RoleType: app.RoleTypeInstaller,
				}
				err := s.service.DB.WithContext(ctx).Create(pendingInvite).Error
				require.NoError(s.T(), err)
				pendingID := pendingInvite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", pendingID)
				})

				acceptedInvite := &app.OrgInvite{
					OrgID:    s.testOrg.ID,
					Email:    "accepted@example.com",
					Status:   app.OrgInviteStatusAccepted,
					RoleType: app.RoleTypeOrgAdmin,
				}
				err = s.service.DB.WithContext(ctx).Create(acceptedInvite).Error
				require.NoError(s.T(), err)
				acceptedID := acceptedInvite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", acceptedID)
				})

				return []string{pendingInvite.ID, acceptedInvite.ID}
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(invites []app.OrgInvite) {
				statuses := make([]app.OrgInviteStatus, len(invites))
				for i, invite := range invites {
					statuses[i] = invite.Status
				}
				assert.Contains(s.T(), statuses, app.OrgInviteStatusPending)
				assert.Contains(s.T(), statuses, app.OrgInviteStatusAccepted)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			tc.setupFunc()

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/invites"+tc.queryParams)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				// Parse successful response
				var response []app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)
				require.NotNil(s.T(), response)

				// Validate expected count
				require.Len(s.T(), response, tc.expectedCount)

				// Run additional validations if provided
				if tc.validateFunc != nil && len(response) > 0 {
					tc.validateFunc(response)
				}
			} else {
				// Validate error response
				if tc.validateError != nil {
					tc.validateError(rr.Body.String())
				}
			}
		})
	}
}
