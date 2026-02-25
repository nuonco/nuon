package service

import (
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

// cleanupAccountsAndRoles removes all accounts except testAcc and their associated roles
// Used to prevent data pollution between subtests
func (s *GetOrgAccountsTestSuite) cleanupAccountsAndRoles() {
	// Delete in correct FK dependency order to avoid constraint violations:

	// 1. Delete account_roles first (has FK to both accounts and roles)
	err := s.service.DB.Exec("DELETE FROM account_roles").Error
	require.NoError(s.T(), err)

	// 2. Delete policies (has FK to roles)
	err = s.service.DB.Unscoped().Where("org_id = ?", s.testOrg.ID).Delete(&app.Policy{}).Error
	require.NoError(s.T(), err)

	// 3. Delete roles (has FK to orgs, but policies are now deleted)
	err = s.service.DB.Unscoped().Where("org_id = ?", s.testOrg.ID).Delete(&app.Role{}).Error
	require.NoError(s.T(), err)

	// 4. Delete all accounts except testAcc
	err = s.service.DB.Unscoped().Where("id != ?", s.testAcc.ID).Delete(&app.Account{}).Error
	require.NoError(s.T(), err)
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
				acc1 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "admin1@example.com",
					Subject:     "admin1-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc1.ID)
				})

				acc2 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "admin2@example.com",
					Subject:     "admin2-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

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
				emails := []string{accounts[0].Email, accounts[1].Email}
				assert.Contains(s.T(), emails, "admin1@example.com")
				assert.Contains(s.T(), emails, "admin2@example.com")
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
					acc := &app.Account{
						ID:          domains.NewAccountID(),
						Email:       fmt.Sprintf("limit-test-user%d@example.com", i),
						Subject:     fmt.Sprintf("limit-test-user-subject-%d", i),
						AccountType: app.AccountTypeAuth0,
					}
					err = s.service.DB.Create(acc).Error
					require.NoError(s.T(), err)
					accID := acc.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", accID)
					})

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
				// Clean up accounts from previous subtests
				s.cleanupAccountsAndRoles()

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
					acc := &app.Account{
						ID:          domains.NewAccountID(),
						Email:       fmt.Sprintf("offset-test-user%d@example.com", i),
						Subject:     fmt.Sprintf("offset-test-user-subject-%d", i),
						AccountType: app.AccountTypeAuth0,
					}
					err = s.service.DB.Create(acc).Error
					require.NoError(s.T(), err)
					accID := acc.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", accID)
					})

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
				// Clean up accounts from previous subtests
				s.cleanupAccountsAndRoles()

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
				nonNuonAcc := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "requester@example.com",
					Subject:     "requester-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(nonNuonAcc).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", nonNuonAcc.ID)
				})

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

				// Create regular user account
				acc1 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "user@example.com",
					Subject:     "user-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc1.ID)
				})

				// Create Nuon employee account
				acc2 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "employee@nuon.co",
					Subject:     "employee-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

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
				assert.Equal(s.T(), "user@example.com", accounts[0].Email)
				// Verify nuon.co account is NOT in the response
				for _, acc := range accounts {
					assert.NotContains(s.T(), acc.Email, "nuon.co")
				}
			},
		},
		{
			name: "shows all accounts including nuon.co for nuon.co users",
			setupFunc: func() []string {
				// Clean up accounts from previous subtests
				s.cleanupAccountsAndRoles()

				ctx := context.Background()

				// CRITICAL: Create new nuon.co account instead of modifying s.testAcc
				// Modifying s.testAcc causes database constraint violations in later tests
				nuonAdminAcc := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "admin-nuon-test@nuon.co",
					Subject:     "admin-nuon-test-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(nuonAdminAcc).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", nuonAdminAcc.ID)
				})

				ctx = cctx.SetAccountContext(ctx, nuonAdminAcc)

				// Create org roles with nuon.co account context
				err = s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				// Get the OrgAdmin role
				var adminRole app.Role
				err = s.service.DB.Where("org_id = ? AND role_type = ?", s.testOrg.ID, app.RoleTypeOrgAdmin).First(&adminRole).Error
				require.NoError(s.T(), err)

				// Create regular user account
				acc1 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "nuon-test-regular-user@example.com",
					Subject:     "nuon-test-regular-user-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc1.ID)
				})

				// Create another Nuon employee account
				acc2 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "nuon-test-employee@nuon.co",
					Subject:     "nuon-test-employee-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

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

				// For nuon.co user, all accounts should be visible
				return []string{acc1.ID, acc2.ID}
			},
			queryParams:   "",
			expectedCount: 2,
			validateFunc: func(accounts []app.Account) {
				assert.Len(s.T(), accounts, 2)
				emails := []string{accounts[0].Email, accounts[1].Email}
				// Both regular and nuon.co accounts should be visible
				assert.Contains(s.T(), emails, "nuon-test-regular-user@example.com")
				assert.Contains(s.T(), emails, "nuon-test-employee@nuon.co")
			},
		},
		{
			name: "only returns accounts for current org",
			setupFunc: func() []string {
				// Clean up accounts from previous subtests
				s.cleanupAccountsAndRoles()

				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
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
				acc1 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "org1-admin@example.com",
					Subject:     "org1-admin-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc1.ID)
				})

				// Create account for other org
				acc2 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "org2-admin@example.com",
					Subject:     "org2-admin-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err = s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

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
				assert.Equal(s.T(), "org1-admin@example.com", accounts[0].Email)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			expectedAccountIDs := tc.setupFunc()

			// For "shows all accounts including nuon.co" test, we need to recreate the router
			// with the nuon.co account context since router middleware captures context at creation
			var testRouter *gin.Engine
			if tc.name == "shows all accounts including nuon.co for nuon.co users" {
				// Get the nuon.co admin account we created in setupFunc
				var nuonAdminAcc app.Account
				err := s.service.DB.Where("email = ?", "admin-nuon-test@nuon.co").First(&nuonAdminAcc).Error
				require.NoError(s.T(), err)

				// Create new router with nuon.co account context
				testRouter = tests.NewTestRouter(tests.RouterOptions{
					L:       s.service.L,
					DB:      s.service.DB,
					TestOrg: s.testOrg,
					TestAcc: &nuonAdminAcc,
				})
				err = s.orgsService.RegisterPublicRoutes(testRouter)
				require.NoError(s.T(), err)
			} else {
				// Use the default router for other tests
				testRouter = s.router
			}

			// Make request with appropriate router
			req, err := http.NewRequest(http.MethodGet, "/v1/orgs/current/accounts"+tc.queryParams, nil)
			require.NoError(s.T(), err)

			rr := httptest.NewRecorder()
			testRouter.ServeHTTP(rr, req)

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
