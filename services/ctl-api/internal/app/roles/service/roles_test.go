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

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// TestService holds all fx-injected dependencies for roles endpoint tests.
type TestService struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	L           *zap.Logger
	AuthzClient *authz.Client
	RolesSvc    *service
	Seeder      *testseed.Seeder
}

type RolesTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service TestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
	testApp *app.App
}

func TestRolesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(RolesTestSuite))
}

func (s *RolesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *RolesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.RolesSvc.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *RolesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *RolesTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())

	require.NoError(s.T(), s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID))
	require.NoError(s.T(), s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, s.testAcc.ID))
	s.testAcc = s.reloadAccount(s.testAcc.ID)
}

func (s *RolesTestSuite) reloadAccount(accountID string) *app.Account {
	var acct app.Account
	res := s.service.DB.
		Preload("Roles").
		Preload("Roles.Policies").
		Preload("Roles.Org").
		First(&acct, "id = ?", accountID)
	require.NoError(s.T(), res.Error)
	return &acct
}

func (s *RolesTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
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

func (s *RolesTestSuite) createCustomRole() *app.Role {
	rr := s.makeRequest(http.MethodPost, "/v1/roles", CreateRoleRequest{
		Title:    "Release manager",
		Contexts: []string{app.RoleContextTeam},
		Permissions: []PermissionEntryRequest{
			{ResourceType: "app", ResourceID: s.testApp.Name, Permissions: []string{"read"}},
			{ResourceType: "app_branch", ResourceID: "*", ScopeType: "app", ScopeID: s.testApp.ID, Permissions: []string{"all"}},
		},
	})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var role app.Role
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &role))
	return &role
}

func (s *RolesTestSuite) TestCreateRole() {
	role := s.createCustomRole()

	require.Equal(s.T(), app.RoleTypeCustom, role.RoleType)
	require.False(s.T(), role.Managed)
	require.Equal(s.T(), "Release manager", role.Title)
	require.Len(s.T(), role.Policies, 1)

	entries := role.Policies[0].ScopedPermissions
	require.Len(s.T(), entries, 2)
	require.Equal(s.T(), s.testApp.ID, entries[0].ResourceID, "app name resolves to its id")
	require.Equal(s.T(), permissions.Verbs{permissions.PermissionRead}, entries[0].Permissions)
	require.Equal(s.T(), "*", entries[1].ResourceID)
	require.Equal(s.T(), s.testApp.ID, entries[1].ScopeID)
}

func (s *RolesTestSuite) TestCreateRoleValidation() {
	cases := []struct {
		name    string
		request CreateRoleRequest
	}{
		{
			name:    "no entries",
			request: CreateRoleRequest{Title: "t", Permissions: []PermissionEntryRequest{}},
		},
		{
			name: "invalid verb",
			request: CreateRoleRequest{Title: "t", Permissions: []PermissionEntryRequest{
				{ResourceType: "app", ResourceID: s.testApp.ID, Permissions: []string{"deploy"}},
			}},
		},
		{
			name: "invalid resource type",
			request: CreateRoleRequest{Title: "t", Permissions: []PermissionEntryRequest{
				{ResourceType: "component", ResourceID: "x", Permissions: []string{"read"}},
			}},
		},
		{
			name: "unknown resource id",
			request: CreateRoleRequest{Title: "t", Permissions: []PermissionEntryRequest{
				{ResourceType: "app", ResourceID: "appdoesnotexistxxxxxxxxxxx", Permissions: []string{"read"}},
			}},
		},
		{
			name: "wildcard on org",
			request: CreateRoleRequest{Title: "t", Permissions: []PermissionEntryRequest{
				{ResourceType: "org", ResourceID: "*", Permissions: []string{"read"}},
			}},
		},
		{
			name: "scope on concrete entry",
			request: CreateRoleRequest{Title: "t", Permissions: []PermissionEntryRequest{
				{ResourceType: "install", ResourceID: "instx", ScopeType: "app", ScopeID: s.testApp.ID, Permissions: []string{"read"}},
			}},
		},
		{
			name: "illegal wildcard scope tier",
			request: CreateRoleRequest{Title: "t", Permissions: []PermissionEntryRequest{
				{ResourceType: "app", ResourceID: "*", ScopeType: "install", ScopeID: "instx", Permissions: []string{"read"}},
			}},
		},
		{
			name: "invalid context",
			request: CreateRoleRequest{Title: "t", Contexts: []string{"nope"}, Permissions: []PermissionEntryRequest{
				{ResourceType: "app", ResourceID: s.testApp.ID, Permissions: []string{"read"}},
			}},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodPost, "/v1/roles", tc.request)
			require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
		})
	}
}

func (s *RolesTestSuite) TestRoleTitlesAreUniquePerOrg() {
	role := s.createCustomRole()

	entry := []PermissionEntryRequest{
		{ResourceType: "app", ResourceID: s.testApp.ID, Permissions: []string{"read"}},
	}

	s.Run("exact duplicate is rejected", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/roles", CreateRoleRequest{
			Title:       role.Title,
			Permissions: entry,
		})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
		require.Contains(s.T(), rr.Body.String(), "already exists")
	})

	s.Run("case and whitespace variants are rejected", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/roles", CreateRoleRequest{
			Title:       "  release MANAGER  ",
			Permissions: entry,
		})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	})

	s.Run("a managed role's title is reserved too", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/roles", CreateRoleRequest{
			Title:       "Admin",
			Permissions: entry,
		})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	})

	s.Run("blank title is rejected", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/roles", CreateRoleRequest{
			Title:       "   ",
			Permissions: entry,
		})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	})

	s.Run("renaming onto another role's title is rejected", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/roles", CreateRoleRequest{
			Title:       "Branch reviewer",
			Permissions: entry,
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

		var other app.Role
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &other))

		rr = s.makeRequest(http.MethodPatch, "/v1/roles/"+other.ID, UpdateRoleRequest{Title: role.Title})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	})

	s.Run("a role keeps its own title on update", func() {
		rr := s.makeRequest(http.MethodPatch, "/v1/roles/"+role.ID, UpdateRoleRequest{Title: role.Title})
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
	})

	s.Run("a deleted role frees its title", func() {
		rr := s.makeRequest(http.MethodDelete, "/v1/roles/"+role.ID, nil)
		require.Equal(s.T(), http.StatusNoContent, rr.Code, rr.Body.String())

		rr = s.makeRequest(http.MethodPost, "/v1/roles", CreateRoleRequest{
			Title:       role.Title,
			Permissions: entry,
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
	})
}

func (s *RolesTestSuite) TestManagedRolesAreImmutable() {
	var managed app.Role
	res := s.service.DB.
		Where(app.Role{OrgID: generics.NewNullString(s.testOrg.ID), RoleType: app.RoleTypeOrgAdmin}).
		First(&managed)
	require.NoError(s.T(), res.Error)

	rr := s.makeRequest(http.MethodPatch, "/v1/roles/"+managed.ID, UpdateRoleRequest{Title: "hijacked"})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())

	rr = s.makeRequest(http.MethodDelete, "/v1/roles/"+managed.ID, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
}

func (s *RolesTestSuite) TestUpdateRole() {
	role := s.createCustomRole()

	contexts := []string{app.RoleContextAPIToken, app.RoleContextServiceAccount}
	rr := s.makeRequest(http.MethodPatch, "/v1/roles/"+role.ID, UpdateRoleRequest{
		Title:    "Renamed",
		Contexts: &contexts,
		Permissions: []PermissionEntryRequest{
			{ResourceType: "app", ResourceID: s.testApp.ID, Permissions: []string{"read", "update"}},
		},
	})
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var updated app.Role
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &updated))
	require.Equal(s.T(), "Renamed", updated.Title)
	require.ElementsMatch(s.T(), contexts, updated.Contexts)
	require.Len(s.T(), updated.Policies, 1)
	require.Len(s.T(), updated.Policies[0].ScopedPermissions, 1)
	require.ElementsMatch(s.T(),
		permissions.Verbs{permissions.PermissionRead, permissions.PermissionUpdate},
		updated.Policies[0].ScopedPermissions[0].Permissions,
	)
}

// An empty context list withdraws a role from every picker, which only persists
// because the update selects the field explicitly rather than skipping its zero
// value.
func (s *RolesTestSuite) TestUpdateRoleWithdrawsEveryContext() {
	role := s.createCustomRole()

	rr := s.makeRequest(http.MethodPatch, "/v1/roles/"+role.ID, UpdateRoleRequest{
		Contexts: &[]string{},
	})
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var updated app.Role
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &updated))
	require.Empty(s.T(), updated.Contexts)
	require.Equal(s.T(), "Release manager", updated.Title)
}

func (s *RolesTestSuite) TestDeleteRole() {
	role := s.createCustomRole()

	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	member := s.service.Seeder.CreateAccount(ctx, s.T())
	require.NoError(s.T(), s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleType(role.ID), s.testOrg.ID, member.ID))

	rr := s.makeRequest(http.MethodDelete, "/v1/roles/"+role.ID, nil)
	require.Equal(s.T(), http.StatusNoContent, rr.Code, rr.Body.String())

	rr = s.makeRequest(http.MethodGet, "/v1/roles/"+role.ID, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)

	var count int64
	require.NoError(s.T(), s.service.DB.Model(&app.AccountRole{}).Where(app.AccountRole{RoleID: role.ID}).Count(&count).Error)
	require.Zero(s.T(), count, "assignments are revoked with the role")
}

func (s *RolesTestSuite) TestNonAdminCannotManageRoles() {
	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	member := s.service.Seeder.CreateAccount(ctx, s.T())
	require.NoError(s.T(), s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgReadOnly, s.testOrg.ID, member.ID))

	memberRouter := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.reloadAccount(member.ID),
	})
	require.NoError(s.T(), s.service.RolesSvc.RegisterPublicRoutes(memberRouter))

	body, err := json.Marshal(CreateRoleRequest{
		Title: "nope",
		Permissions: []PermissionEntryRequest{
			{ResourceType: "app", ResourceID: s.testApp.ID, Permissions: []string{"read"}},
		},
	})
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, "/v1/roles", bytes.NewBuffer(body))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	memberRouter.ServeHTTP(rr, req)
	require.Equal(s.T(), http.StatusForbidden, rr.Code, rr.Body.String())
}

func (s *RolesTestSuite) TestCustomRoleResolvesIntoAccountPermissions() {
	role := s.createCustomRole()

	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	member := s.service.Seeder.CreateAccount(ctx, s.T())
	require.NoError(s.T(), s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleType(role.ID), s.testOrg.ID, member.ID))

	acct := s.reloadAccount(member.ID)

	require.NoError(s.T(), acct.AllPermissions.CanPerform(s.testApp.ID, permissions.PermissionRead))
	require.Error(s.T(), acct.AllPermissions.CanPerform(s.testApp.ID, permissions.PermissionDelete))

	grants := acct.TypeGrants[s.testOrg.ID][app.LevelAppBranch]
	require.Len(s.T(), grants, 1)
	require.Equal(s.T(), s.testApp.ID, grants[0].ScopeID)
	require.Contains(s.T(), acct.OrgIDs, s.testOrg.ID, "custom role assignment grants org membership")
}
