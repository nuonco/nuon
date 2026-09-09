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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type GetOrgMembersTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	AuthzClient     *authz.Client
	OrgsService     *service
	Seeder          *testseed.Seeder
}

type GetOrgMembersTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service GetOrgMembersTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestGetOrgMembers(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgMembersTestSuite))
}

func (s *GetOrgMembersTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *GetOrgMembersTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgMembersTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgMembersTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *GetOrgMembersTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetOrgMembersTestSuite) cleanupOrgRoles() {
	err := s.service.DB.Exec(
		"DELETE FROM account_roles WHERE role_id IN (SELECT id FROM roles WHERE org_id = ?)",
		s.testOrg.ID,
	).Error
	require.NoError(s.T(), err)

	err = s.service.DB.Unscoped().Where("org_id = ?", s.testOrg.ID).Delete(&app.Policy{}).Error
	require.NoError(s.T(), err)

	err = s.service.DB.Unscoped().Where("org_id = ?", s.testOrg.ID).Delete(&app.Role{}).Error
	require.NoError(s.T(), err)
}

func (s *GetOrgMembersTestSuite) cleanupAccount(acc *app.Account) {
	s.T().Cleanup(func() {
		s.service.DB.Exec("DELETE FROM account_roles WHERE account_id = ?", acc.ID)
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc.ID)
	})
}

func (s *GetOrgMembersTestSuite) cleanupInvite(invite *app.OrgInvite) {
	inviteID := invite.ID
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
	})
}

func (s *GetOrgMembersTestSuite) accountCtx() context.Context {
	return cctx.SetAccountContext(context.Background(), s.testAcc)
}

func (s *GetOrgMembersTestSuite) ensureOrgRoles() {
	err := s.service.DB.Unscoped().Where(app.OrgInvite{OrgID: s.testOrg.ID}).Delete(&app.OrgInvite{}).Error
	require.NoError(s.T(), err)
	s.cleanupOrgRoles()
	err = s.service.AuthzClient.CreateOrgRoles(s.accountCtx(), s.testOrg.ID)
	require.NoError(s.T(), err)
}

func (s *GetOrgMembersTestSuite) createMember(email, name string, roleType app.RoleType) *app.Account {
	accID := domains.NewAccountID()
	acc := &app.Account{
		ID:          accID,
		Email:       email,
		Name:        name,
		Subject:     accID,
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(acc).Error
	require.NoError(s.T(), err)
	s.cleanupAccount(acc)

	err = s.service.AuthzClient.AddAccountOrgRole(s.accountCtx(), roleType, s.testOrg.ID, acc.ID)
	require.NoError(s.T(), err)
	return acc
}

func (s *GetOrgMembersTestSuite) createInvite(email string, status app.OrgInviteStatus, roleType app.RoleType) *app.OrgInvite {
	invite := &app.OrgInvite{
		OrgID:    s.testOrg.ID,
		Email:    email,
		Status:   status,
		RoleType: roleType,
	}
	err := s.service.DB.WithContext(s.accountCtx()).Create(invite).Error
	require.NoError(s.T(), err)
	s.cleanupInvite(invite)
	return invite
}

func (s *GetOrgMembersTestSuite) setCaller(acc *app.Account) {
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: acc,
	})
	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgMembersTestSuite) firstAccountRole(accountID string) app.AccountRole {
	var role app.AccountRole
	err := s.service.DB.
		Where(app.AccountRole{AccountID: accountID}).
		Order("created_at ASC").
		First(&role).Error
	require.NoError(s.T(), err)
	return role
}

func (s *GetOrgMembersTestSuite) TestGetOrgMembers() {
	testCases := []struct {
		name          string
		setupFunc     func()
		queryParams   string
		expectedCount int
		expectedCode  int
		validateFunc  func([]app.OrgMember, *httptest.ResponseRecorder)
	}{
		{
			name: "returns empty array when org has no members",
			setupFunc: func() {
				s.ensureOrgRoles()
			},
			queryParams:   "",
			expectedCount: 0,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, rr *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "[]", strings.TrimSpace(rr.Body.String()))
				assert.NotNil(s.T(), members)
			},
		},
		{
			name: "returns active members and pending invites in one email-ordered list",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("zeta@example.com", "Zeta", app.RoleTypeOrgAdmin)
				s.createMember("alpha@example.com", "Alpha", app.RoleTypeOrgReadOnly)
				s.createInvite("middle@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "",
			expectedCount: 3,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "alpha@example.com", members[0].Email)
				assert.Equal(s.T(), app.OrgMemberStatusActive, members[0].Status)
				assert.Equal(s.T(), "middle@example.com", members[1].Email)
				assert.Equal(s.T(), app.OrgMemberStatusInvited, members[1].Status)
				assert.Equal(s.T(), "zeta@example.com", members[2].Email)
				assert.Equal(s.T(), app.OrgMemberStatusActive, members[2].Status)
			},
		},
		{
			name: "excludes accepted and revoked invites",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createInvite("pending@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
				s.createInvite("accepted@example.com", app.OrgInviteStatusAccepted, app.RoleTypeOrgAdmin)
				revoked := s.createInvite("revoked@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
				revoked.Status = app.OrgInviteStatusRevoked
				err := s.service.DB.WithContext(s.accountCtx()).Save(revoked).Error
				require.NoError(s.T(), err)
				err = s.service.DB.WithContext(s.accountCtx()).Delete(revoked).Error
				require.NoError(s.T(), err)
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "pending@example.com", members[0].Email)
			},
		},
		{
			name: "excludes service accounts",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("human@example.com", "Human", app.RoleTypeOrgAdmin)
				svc := s.service.Seeder.CreateServiceAccount(s.accountCtx(), s.T(), domains.NewAccountID())
				s.cleanupAccount(svc)
				err := s.service.AuthzClient.AddAccountOrgRole(s.accountCtx(), app.RoleTypeRunner, s.testOrg.ID, svc.ID)
				require.NoError(s.T(), err)
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "human@example.com", members[0].Email)
			},
		},
		{
			name: "hides nuon.co members and invites from a non-nuon.co caller",
			setupFunc: func() {
				s.ensureOrgRoles()
				callerID := domains.NewAccountID()
				caller := &app.Account{
					ID:          callerID,
					Email:       fmt.Sprintf("%s@example.com", callerID),
					Subject:     callerID,
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(caller).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(caller)
				s.setCaller(caller)

				s.createMember("customer@example.com", "Customer", app.RoleTypeOrgAdmin)
				s.createMember(fmt.Sprintf("staff-%s@nuon.co", domains.NewAccountID()[:8]), "Staff", app.RoleTypeOrgAdmin)
				s.createInvite("invite-customer@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
				s.createInvite(fmt.Sprintf("invite-staff-%s@nuon.co", domains.NewAccountID()[:8]), app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				for _, m := range members {
					assert.NotContains(s.T(), m.Email, "nuon.co")
				}
			},
		},
		{
			name: "shows nuon.co members and invites to a nuon.co caller",
			setupFunc: func() {
				s.ensureOrgRoles()
				callerID := domains.NewAccountID()
				caller := &app.Account{
					ID:          callerID,
					Email:       fmt.Sprintf("%s@nuon.co", callerID),
					Subject:     callerID,
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(caller).Error
				require.NoError(s.T(), err)
				s.cleanupAccount(caller)
				s.setCaller(caller)

				s.createMember("visible-customer@example.com", "Customer", app.RoleTypeOrgAdmin)
				s.createMember(fmt.Sprintf("visible-staff-%s@nuon.co", domains.NewAccountID()[:8]), "Staff", app.RoleTypeOrgAdmin)
				s.createInvite("visible-invite-customer@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
				s.createInvite(fmt.Sprintf("visible-invite-staff-%s@nuon.co", domains.NewAccountID()[:8]), app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "",
			expectedCount: 4,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				var hasNuonMember, hasNuonInvite bool
				for _, m := range members {
					if strings.HasSuffix(m.Email, "@nuon.co") && m.Status == app.OrgMemberStatusActive {
						hasNuonMember = true
					}
					if strings.HasSuffix(m.Email, "@nuon.co") && m.Status == app.OrgMemberStatusInvited {
						hasNuonInvite = true
					}
				}
				assert.True(s.T(), hasNuonMember)
				assert.True(s.T(), hasNuonInvite)
			},
		},
		{
			name: "pending invite matching an active member email appears once as the member",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("dup@example.com", "Dup", app.RoleTypeOrgAdmin)
				s.createInvite("dup@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgReadOnly)
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), app.OrgMemberStatusActive, members[0].Status)
				assert.Equal(s.T(), "dup@example.com", members[0].Email)
				assert.NotEmpty(s.T(), members[0].AccountID)
				assert.Empty(s.T(), members[0].InviteID)
			},
		},
		{
			name: "joined_at is account_roles.created_at for members and null for invites",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("joined@example.com", "Joined", app.RoleTypeOrgAdmin)
				s.createInvite("not-joined@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				var memberRow, inviteRow app.OrgMember
				for _, m := range members {
					switch m.Status {
					case app.OrgMemberStatusActive:
						memberRow = m
					case app.OrgMemberStatusInvited:
						inviteRow = m
					}
				}
				role := s.firstAccountRole(memberRow.AccountID)
				require.NotNil(s.T(), memberRow.JoinedAt)
				assert.WithinDuration(s.T(), role.CreatedAt, *memberRow.JoinedAt, time.Second)
				assert.Nil(s.T(), inviteRow.JoinedAt)
			},
		},
		{
			name: "name is set for members and empty for invites",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("named@example.com", "Named Person", app.RoleTypeOrgAdmin)
				s.createInvite("unnamed@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				byEmail := map[string]app.OrgMember{}
				for _, m := range members {
					byEmail[m.Email] = m
				}
				assert.Equal(s.T(), "Named Person", byEmail["named@example.com"].Name)
				assert.Empty(s.T(), byEmail["unnamed@example.com"].Name)
			},
		},
		{
			name: "exactly one of account_id and invite_id is set",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("ids-member@example.com", "IDs", app.RoleTypeOrgAdmin)
				s.createInvite("ids-invite@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				for _, m := range members {
					hasAccount := m.AccountID != ""
					hasInvite := m.InviteID != ""
					assert.True(s.T(), hasAccount != hasInvite)
					if hasAccount {
						assert.Equal(s.T(), m.ID, m.AccountID)
					}
					if hasInvite {
						assert.Equal(s.T(), m.ID, m.InviteID)
					}
				}
			},
		},
		{
			name: "account with two org roles yields one row with the earliest grant",
			setupFunc: func() {
				s.ensureOrgRoles()
				acc := s.createMember("tworoles@example.com", "Two Roles", app.RoleTypeOrgAdmin)
				first := s.firstAccountRole(acc.ID)
				older := first.CreatedAt.Add(-2 * time.Hour)
				err := s.service.DB.Model(&app.AccountRole{}).Where(app.AccountRole{ID: first.ID}).Update("created_at", older).Error
				require.NoError(s.T(), err)
				err = s.service.AuthzClient.AddAccountOrgRole(s.accountCtx(), app.RoleTypeOrgReadOnly, s.testOrg.ID, acc.ID)
				require.NoError(s.T(), err)
			},
			queryParams:   "",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), app.RoleTypeOrgAdmin, members[0].RoleType)
				require.NotNil(s.T(), members[0].JoinedAt)
				role := s.firstAccountRole(members[0].AccountID)
				assert.WithinDuration(s.T(), role.CreatedAt, *members[0].JoinedAt, time.Second)
			},
		},
		{
			name: "q matches email substring case-insensitively across members and invites",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("keep-alpha@example.com", "Alpha", app.RoleTypeOrgAdmin)
				s.createMember("findme-member@example.com", "Other", app.RoleTypeOrgAdmin)
				s.createInvite("FINDME-invite@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "?q=FindMe",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				emails := []string{members[0].Email, members[1].Email}
				assert.Contains(s.T(), emails, "findme-member@example.com")
				assert.Contains(s.T(), emails, "FINDME-invite@example.com")
			},
		},
		{
			name: "q matches member name",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("name-search@example.com", "UniqueDisplayName", app.RoleTypeOrgAdmin)
				s.createMember("other-search@example.com", "Nope", app.RoleTypeOrgAdmin)
			},
			queryParams:   "?q=uniquedisplay",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "name-search@example.com", members[0].Email)
			},
		},
		{
			name: "status=active returns only members",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("active-only@example.com", "Active", app.RoleTypeOrgAdmin)
				s.createInvite("invited-only@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "?status=active",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), app.OrgMemberStatusActive, members[0].Status)
			},
		},
		{
			name: "status=invited returns only invites",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("active-hidden@example.com", "Active", app.RoleTypeOrgAdmin)
				s.createInvite("invited-shown@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "?status=invited",
			expectedCount: 1,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				assert.Equal(s.T(), app.OrgMemberStatusInvited, members[0].Status)
			},
		},
		{
			name: "status=active,invited returns both",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("both-active@example.com", "Active", app.RoleTypeOrgAdmin)
				s.createInvite("both-invited@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
			},
			queryParams:   "?status=active,invited",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
		},
		{
			name: "unrecognised status returns 400",
			setupFunc: func() {
				s.ensureOrgRoles()
			},
			queryParams:   "?status=nope",
			expectedCount: 0,
			expectedCode:  http.StatusBadRequest,
		},
		{
			name: "role_type filters members and invites",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("admin-member@example.com", "Admin", app.RoleTypeOrgAdmin)
				s.createMember("ro-member@example.com", "RO", app.RoleTypeOrgReadOnly)
				s.createInvite("admin-invite@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgAdmin)
				s.createInvite("ro-invite@example.com", app.OrgInviteStatusPending, app.RoleTypeOrgReadOnly)
			},
			queryParams:   "?role_type=org_admin",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(members []app.OrgMember, _ *httptest.ResponseRecorder) {
				for _, m := range members {
					assert.Equal(s.T(), app.RoleTypeOrgAdmin, m.RoleType)
				}
			},
		},
		{
			name: "pagination pages without duplicates or gaps and flips has-next",
			setupFunc: func() {
				s.ensureOrgRoles()
				s.createMember("page-aaa@example.com", "A", app.RoleTypeOrgAdmin)
				s.createMember("page-bbb@example.com", "B", app.RoleTypeOrgAdmin)
				s.createMember("page-ccc@example.com", "C", app.RoleTypeOrgAdmin)
			},
			queryParams:   "?limit=2",
			expectedCount: 2,
			expectedCode:  http.StatusOK,
			validateFunc: func(page1 []app.OrgMember, rr *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "true", rr.Header().Get("X-Nuon-Page-Next"))
				assert.Equal(s.T(), "page-aaa@example.com", page1[0].Email)
				assert.Equal(s.T(), "page-bbb@example.com", page1[1].Email)

				rr2 := s.makeRequest(http.MethodGet, "/v1/orgs/current/members?limit=2&offset=2")
				require.Equal(s.T(), http.StatusOK, rr2.Code, "body: %s", rr2.Body.String())
				assert.Equal(s.T(), "false", rr2.Header().Get("X-Nuon-Page-Next"))

				var page2 []app.OrgMember
				require.NoError(s.T(), json.Unmarshal(rr2.Body.Bytes(), &page2))
				require.Len(s.T(), page2, 1)
				assert.Equal(s.T(), "page-ccc@example.com", page2[0].Email)

				seen := map[string]struct{}{
					page1[0].ID: {},
					page1[1].ID: {},
					page2[0].ID: {},
				}
				assert.Len(s.T(), seen, 3)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupFunc()

			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/members"+tc.queryParams)
			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode != http.StatusOK {
				return
			}

			var response []app.OrgMember
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.Len(s.T(), response, tc.expectedCount)

			if tc.validateFunc != nil {
				tc.validateFunc(response, rr)
			}
		})
	}
}
