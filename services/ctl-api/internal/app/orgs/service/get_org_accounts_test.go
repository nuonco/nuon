package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// GetOrgAccountsTestService holds all fx-injected dependencies for get org accounts tests.
type GetOrgAccountsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	AuthzClient     *authz.Client
	Seeder          *testseed.Seeder
}

// GetOrgAccountsTestSuite is the testify suite for GET /v1/orgs/current/accounts endpoint.
type GetOrgAccountsTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     GetOrgAccountsTestService
	router      *gin.Engine
	testAcc     *app.Account
	testOrg     *app.Org
	orgsService *service
}

func TestGetOrgAccountsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgAccountsTestSuite))
}

func (s *GetOrgAccountsTestSuite) SetupSuite() {
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

func (s *GetOrgAccountsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

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

func (s *GetOrgAccountsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgAccountsTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *GetOrgAccountsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// cleanupOrgRoles removes roles, policies, and account_roles for testOrg only.
// Scoped to testOrg to avoid FK violations from broad deletes.
// Account cleanup is handled by per-account s.T().Cleanup() functions.
func (s *GetOrgAccountsTestSuite) cleanupOrgRoles() {
	// 1. Delete account_roles for this org's roles only
	err := s.service.DB.Exec(
		"DELETE FROM account_roles WHERE role_id IN (SELECT id FROM roles WHERE org_id = ?)",
		s.testOrg.ID,
	).Error
	require.NoError(s.T(), err)

	// 2. Delete policies for testOrg
	err = s.service.DB.Unscoped().Where("org_id = ?", s.testOrg.ID).Delete(&app.Policy{}).Error
	require.NoError(s.T(), err)

	// 3. Delete roles for testOrg
	err = s.service.DB.Unscoped().Where("org_id = ?", s.testOrg.ID).Delete(&app.Role{}).Error
	require.NoError(s.T(), err)
}

// cleanupAccount registers cleanup to delete an account and its account_roles.
func (s *GetOrgAccountsTestSuite) cleanupAccount(acc *app.Account) {
	s.T().Cleanup(func() {
		s.service.DB.Exec("DELETE FROM account_roles WHERE account_id = ?", acc.ID)
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc.ID)
	})
}

func (s *GetOrgAccountsTestSuite) TestGetOrgAccounts() {
	testCases := []struct {
		name          string
		setupFunc     func() []string // Returns account IDs that should be returned
		queryParams   string
		expectedCount int
		validateFunc  func([]app.Account) // Additional validations
	}{
		{
			name: "returns empty array when no accounts have org admin role",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org roles (OrgAdmin role must exist even if no accounts assigned)
				err := s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				// Org exists and role exists, but no accounts have been assigned admin role yet
				return []string{}
			},
			queryParams:   "",
			expectedCount: 0,
		},
		{
			name: "returns accounts with org admin role",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org roles (this creates OrgAdmin, Installer, Runner roles)
				err := s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				// Get the OrgAdmin role
				var adminRole app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", s.testOrg.ID, app.RoleTypeOrgAdmin).First(&adminRole).Error
				require.NoError(s.T(), err)

				// Create two accounts and assign them as org admins
				acc1ID := domains.NewAccountID()
				acc1 := &app.Account{
					ID:          acc1ID,
					Email:       fmt.Sprintf("%s@test.nuon.co", acc1ID),
					Subject:     "admin1-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc1)

				acc2ID := domains.NewAccountID()
				acc2 := &app.Account{
					ID:          acc2ID,
					Email:       fmt.Sprintf("%s@test.nuon.co", acc2ID),
					Subject:     "admin2-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc2)

				// Assign accounts to org admin role
				accountRole1 := &app.AccountRole{
					AccountID: acc1.ID,
					RoleID:    adminRole.ID,
				}
				err = s.service.DB.Create(accountRole1).Error
				require.NoError(s.T(), err)

				accountRole2 := &app.AccountRole{
					AccountID: acc2.ID,
					RoleID:    adminRole.ID,
				}
				err = s.service.DB.Create(accountRole2).Error
				require.NoError(s.T(), err)

				return []string{acc1.ID, acc2.ID}
			},
			queryParams:   "",
			expectedCount: 2,
			validateFunc: func(accounts []app.Account) {
				for _, acc := range accounts {
					assert.Contains(s.T(), acc.Email, "@test.nuon.co")
				}
			},
		},
		{
			name: "respects pagination - limit parameter",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org roles
				err := s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				// Get the OrgAdmin role
				var adminRole app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", s.testOrg.ID, app.RoleTypeOrgAdmin).First(&adminRole).Error
				require.NoError(s.T(), err)

				accountIDs := make([]string, 0, 10)
				for i := 0; i < 10; i++ {
					accID := domains.NewAccountID()
					acc := &app.Account{
						ID:          accID,
						Email:       fmt.Sprintf("limit-test-user%d-%s@test.nuon.co", i, accID[:8]),
						Subject:     fmt.Sprintf("limit-test-user-subject-%d", i),
						AccountType: app.AccountTypeAuth0,
					}
					err = s.service.DB.Create(acc).Error
					require.NoError(s.T(), err)
					s.cleanupAccount(acc)

					// Assign to org admin role
					accountRole := &app.AccountRole{
						AccountID: acc.ID,
						RoleID:    adminRole.ID,
					}
					err = s.service.DB.Create(accountRole).Error
					require.NoError(s.T(), err)

					accountIDs = append(accountIDs, acc.ID)
				}

				// Don't return accountIDs since handler has no ORDER BY
				// We can only verify the count, not which specific accounts are returned
				return []string{}
			},
			queryParams:   "?limit=5",
			expectedCount: 5,
			validateFunc:  nil, // Handler has no ORDER BY, so we can only verify count
		},
		{
			name: "respects pagination - offset parameter",
			setupFunc: func() []string {
				// Clean up roles from previous subtests
				s.cleanupOrgRoles()

				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org roles
				err := s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				// Get the OrgAdmin role
				var adminRole app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", s.testOrg.ID, app.RoleTypeOrgAdmin).First(&adminRole).Error
				require.NoError(s.T(), err)

				accountIDs := make([]string, 0, 5)
				for i := 0; i < 5; i++ {
					accID := domains.NewAccountID()
					acc := &app.Account{
						ID:          accID,
						Email:       fmt.Sprintf("offset-test-user%d-%s@test.nuon.co", i, accID[:8]),
						Subject:     fmt.Sprintf("offset-test-user-subject-%d", i),
						AccountType: app.AccountTypeAuth0,
					}
					err = s.service.DB.Create(acc).Error
					require.NoError(s.T(), err)
					s.cleanupAccount(acc)

					// Assign to org admin role
					accountRole := &app.AccountRole{
						AccountID: acc.ID,
						RoleID:    adminRole.ID,
					}
					err = s.service.DB.Create(accountRole).Error
					require.NoError(s.T(), err)

					accountIDs = append(accountIDs, acc.ID)
				}

				// Don't return accountIDs since handler has no ORDER BY
				// We can only verify the count, not which specific accounts are returned
				return []string{}
			},
			queryParams:   "?offset=2",
			expectedCount: 3,
			validateFunc:  nil, // Handler has no ORDER BY, so we can only verify count
		},
		{
			name: "filters nuon.co emails for non-nuon users",
			setupFunc: func() []string {
				// Clean up roles from previous subtests
				s.cleanupOrgRoles()

				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create org roles
				err := s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				// Get the OrgAdmin role
				var adminRole app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", s.testOrg.ID, app.RoleTypeOrgAdmin).First(&adminRole).Error
				require.NoError(s.T(), err)

				// Create a non-nuon requesting user account for this test
				nonNuonAccID := domains.NewAccountID()
				nonNuonAcc := &app.Account{
					ID:          nonNuonAccID,
					Email:       fmt.Sprintf("%s@example.com", nonNuonAccID),
					Subject:     "requester-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(nonNuonAcc).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(nonNuonAcc)

				// Recreate router with non-nuon account context
				testRouter := tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: s.testOrg,
					TestAcc: nonNuonAcc,
				})
				err = s.orgsService.RegisterPublicRoutes(testRouter)
				require.NoError(s.T(), err)
				s.router = testRouter

				// Create regular user account (non-nuon email)
				acc1ID := domains.NewAccountID()
				acc1 := &app.Account{
					ID:          acc1ID,
					Email:       fmt.Sprintf("%s@example.com", acc1ID),
					Subject:     "user-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc1)

				// Create Nuon employee account
				acc2 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "employee@nuon.co",
					Subject:     "employee-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc2)

				// Assign both to org admin role
				accountRole1 := &app.AccountRole{
					AccountID: acc1.ID,
					RoleID:    adminRole.ID,
				}
				err = s.service.DB.Create(accountRole1).Error
				require.NoError(s.T(), err)

				accountRole2 := &app.AccountRole{
					AccountID: acc2.ID,
					RoleID:    adminRole.ID,
				}
				err = s.service.DB.Create(accountRole2).Error
				require.NoError(s.T(), err)

				// For non-nuon.co user, nuon.co email should be filtered
				// So we should only see acc1
				return []string{acc1.ID}
			},
			queryParams:   "",
			expectedCount: 1,
			validateFunc: func(accounts []app.Account) {
				assert.Len(s.T(), accounts, 1)
				// Verify nuon.co account is NOT in the response
				for _, acc := range accounts {
					assert.NotContains(s.T(), acc.Email, "@nuon.co")
				}
			},
		},
		{
			name: "shows all accounts including nuon.co for nuon.co users",
			setupFunc: func() []string {
				// Clean up roles from previous subtests
				s.cleanupOrgRoles()

				ctx := context.Background()

				// CRITICAL: Create new nuon.co account instead of modifying s.testAcc
				// Modifying s.testAcc causes database constraint violations in later tests
				nuonAdminAccID := domains.NewAccountID()
				nuonAdminAcc := &app.Account{
					ID:          nuonAdminAccID,
					Email:       fmt.Sprintf("%s@nuon.co", nuonAdminAccID),
					Subject:     fmt.Sprintf("admin-nuon-%s", nuonAdminAccID),
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(nuonAdminAcc).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(nuonAdminAcc)

				ctx = cctx.SetAccountContext(ctx, nuonAdminAcc)

				// Create org roles with nuon.co account context
				err = s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				// Get the OrgAdmin role
				var adminRole app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", s.testOrg.ID, app.RoleTypeOrgAdmin).First(&adminRole).Error
				require.NoError(s.T(), err)

				// Create regular user account
				acc1ID := domains.NewAccountID()
				acc1 := &app.Account{
					ID:          acc1ID,
					Email:       fmt.Sprintf("nuon-test-regular-user-%s@example.com", acc1ID[:8]),
					Subject:     "nuon-test-regular-user-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc1)

				// Create another Nuon employee account
				acc2ID := domains.NewAccountID()
				acc2 := &app.Account{
					ID:          acc2ID,
					Email:       fmt.Sprintf("nuon-test-employee-%s@nuon.co", acc2ID[:8]),
					Subject:     "nuon-test-employee-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc2)

				// Assign both to org admin role
				accountRole1 := &app.AccountRole{
					AccountID: acc1.ID,
					RoleID:    adminRole.ID,
				}
				err = s.service.DB.Create(accountRole1).Error
				require.NoError(s.T(), err)

				accountRole2 := &app.AccountRole{
					AccountID: acc2.ID,
					RoleID:    adminRole.ID,
				}
				err = s.service.DB.Create(accountRole2).Error
				require.NoError(s.T(), err)

				// Recreate router with nuon.co account context
				testRouter := tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: s.testOrg,
					TestAcc: nuonAdminAcc,
				})
				err = s.orgsService.RegisterPublicRoutes(testRouter)
				require.NoError(s.T(), err)
				s.router = testRouter

				// For nuon.co user, all accounts should be visible
				return []string{acc1.ID, acc2.ID}
			},
			queryParams:   "",
			expectedCount: 2,
			validateFunc: func(accounts []app.Account) {
				assert.Len(s.T(), accounts, 2)
				// Both regular and nuon.co accounts should be visible
				hasNuonEmail := false
				for _, acc := range accounts {
					if strings.HasSuffix(acc.Email, "@nuon.co") {
						hasNuonEmail = true
					}
				}
				assert.True(s.T(), hasNuonEmail, "nuon.co account should be visible to nuon.co users")
			},
		},
		{
			name: "only returns accounts for current org",
			setupFunc: func() []string {
				// Clean up roles from previous subtests
				s.cleanupOrgRoles()

				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					// Clean up org2's roles/policies/account_roles before deleting org
					s.service.DB.Exec("DELETE FROM account_roles WHERE role_id IN (SELECT id FROM roles WHERE org_id = ?)", org2.ID)
					s.service.DB.Unscoped().Where("org_id = ?", org2.ID).Delete(&app.Policy{})
					s.service.DB.Unscoped().Where("org_id = ?", org2.ID).Delete(&app.Role{})
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)
				})

				// Create org roles for both orgs
				err = s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				err = s.service.AuthzClient.CreateOrgRoles(ctx, org2.ID)
				require.NoError(s.T(), err)

				// Get admin roles for both orgs
				var adminRole1 app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", s.testOrg.ID, app.RoleTypeOrgAdmin).First(&adminRole1).Error
				require.NoError(s.T(), err)

				var adminRole2 app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", org2.ID, app.RoleTypeOrgAdmin).First(&adminRole2).Error
				require.NoError(s.T(), err)

				// Create account for current org
				acc1ID := domains.NewAccountID()
				acc1 := &app.Account{
					ID:          acc1ID,
					Email:       fmt.Sprintf("%s@test.nuon.co", acc1ID),
					Subject:     "org1-admin-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc1)

				// Create account for other org
				acc2ID := domains.NewAccountID()
				acc2 := &app.Account{
					ID:          acc2ID,
					Email:       fmt.Sprintf("%s@test.nuon.co", acc2ID),
					Subject:     "org2-admin-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(acc2)

				// Assign to respective org admin roles
				accountRole1 := &app.AccountRole{
					AccountID: acc1.ID,
					RoleID:    adminRole1.ID,
				}
				err = s.service.DB.Create(accountRole1).Error
				require.NoError(s.T(), err)

				accountRole2 := &app.AccountRole{
					AccountID: acc2.ID,
					RoleID:    adminRole2.ID,
				}
				err = s.service.DB.Create(accountRole2).Error
				require.NoError(s.T(), err)

				// Should only see acc1 for current org
				return []string{acc1.ID}
			},
			queryParams:   "",
			expectedCount: 1,
			validateFunc: func(accounts []app.Account) {
				assert.Len(s.T(), accounts, 1)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data (setupFunc may reassign s.router for tests that need different context)
			expectedAccountIDs := tc.setupFunc()

			// Make request using s.router (may have been reassigned by setupFunc)
			req, err := http.NewRequest(http.MethodGet, "/v1/orgs/current/accounts"+tc.queryParams, nil)
			require.NoError(s.T(), err)

			rr := httptest.NewRecorder()
			s.router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			// Parse response
			var response []app.Account
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.NotNil(s.T(), response)

			// Validate expected count
			require.Len(s.T(), response, tc.expectedCount)

			// Verify returned accounts match expected IDs (if any)
			if tc.expectedCount > 0 && len(expectedAccountIDs) > 0 {
				responseIDs := make([]string, len(response))
				for i, acc := range response {
					responseIDs[i] = acc.ID
				}
				for _, expectedID := range expectedAccountIDs {
					assert.Contains(s.T(), responseIDs, expectedID)
				}
			}

			// Run additional validations if provided
			if tc.validateFunc != nil && len(response) > 0 {
				tc.validateFunc(response)
			}
		})
	}
}
