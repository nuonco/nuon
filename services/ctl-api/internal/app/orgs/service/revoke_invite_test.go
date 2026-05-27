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

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type RevokeOrgInviteTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
}

type RevokeOrgInviteTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      RevokeOrgInviteTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
	orgsService  *service
}

func TestRevokeOrgInviteSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(RevokeOrgInviteTestSuite))
}

func (s *RevokeOrgInviteTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.mockEvClient = tests.NewMockEventLoopClient()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			Mocks:           &tests.TestMocks{MockEv: s.mockEvClient},
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *RevokeOrgInviteTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()
	s.mockEvClient.Reset()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.orgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *RevokeOrgInviteTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *RevokeOrgInviteTestSuite) setupTestData() {
	accID := domains.NewAccountID()
	testAcc := &app.Account{
		ID:          accID,
		Email:       fmt.Sprintf("%s@test.nuon.co", accID),
		Subject:     accID,
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)

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

	role := &app.Role{
		OrgID:    generics.NewNullString(orgID),
		RoleType: app.RoleTypeOrgAdmin,
	}
	err = s.service.DB.WithContext(ctx).Create(role).Error
	require.NoError(s.T(), err)

	err = s.service.DB.Exec(
		"INSERT INTO account_roles (account_id, role_id) VALUES (?, ?)",
		testAcc.ID, role.ID,
	).Error
	require.NoError(s.T(), err)

	testAcc.Roles = []app.Role{*role}
	s.testAcc = testAcc
	s.testOrg = testOrg
}

func (s *RevokeOrgInviteTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *RevokeOrgInviteTestSuite) createPendingInvite(email string) *app.OrgInvite {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	invite := &app.OrgInvite{
		Email:    email,
		OrgID:    s.testOrg.ID,
		Status:   app.OrgInviteStatusPending,
		RoleType: app.RoleTypeOrgAdmin,
	}
	err := s.service.DB.WithContext(ctx).Create(invite).Error
	require.NoError(s.T(), err)

	inviteID := invite.ID
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
	})

	return invite
}

func (s *RevokeOrgInviteTestSuite) TestRevokeOrgInvite() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedStatus   int
		validateResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "successfully revokes pending invite",
			setupFunc: func() string {
				invite := s.createPendingInvite(
					fmt.Sprintf("revoke-success-%s@test.nuon.co", domains.NewAccountID()[:8]),
				)
				return invite.ID
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), invite.ID)
				assert.Equal(s.T(), app.OrgInviteStatusRevoked, invite.Status)
				assert.Equal(s.T(), s.testOrg.ID, invite.OrgID)

				var dbInvite app.OrgInvite
				err = s.service.DB.Unscoped().Where("id = ?", invite.ID).First(&dbInvite).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.OrgInviteStatusRevoked, dbInvite.Status)
				assert.NotZero(s.T(), dbInvite.DeletedAt, "invite should be soft-deleted")
			},
		},
		{
			name: "returns error for non-existent invite ID",
			setupFunc: func() string {
				return domains.NewOrgID()
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "record not found")
			},
		},
		{
			name: "returns 400 for already accepted invite",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				invite := &app.OrgInvite{
					Email:    fmt.Sprintf("accepted-%s@test.nuon.co", domains.NewAccountID()[:8]),
					OrgID:    s.testOrg.ID,
					Status:   app.OrgInviteStatusAccepted,
					RoleType: app.RoleTypeOrgAdmin,
				}
				err := s.service.DB.WithContext(ctx).Create(invite).Error
				require.NoError(s.T(), err)
				inviteID := invite.ID
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", inviteID)
				})
				return invite.ID
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "only pending invites can be revoked")
			},
		},
		{
			name: "returns 400 for empty invite_id",
			setupFunc: func() string {
				return ""
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invite_id is required")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			inviteID := tc.setupFunc()

			path := fmt.Sprintf("/v1/orgs/current/invites/%s/revoke", inviteID)
			rr := s.makeRequest(http.MethodPost, path, nil)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if tc.validateResponse != nil {
				tc.validateResponse(rr)
			}
		})
	}
}

func (s *RevokeOrgInviteTestSuite) TestRevokeOrgInvite_NonAdminForbidden() {
	invite := s.createPendingInvite(
		fmt.Sprintf("non-admin-%s@test.nuon.co", domains.NewAccountID()[:8]),
	)

	nonAdminAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       fmt.Sprintf("nonadmin-%s@test.nuon.co", domains.NewAccountID()[:8]),
		Subject:     domains.NewAccountID(),
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(nonAdminAcc).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", nonAdminAcc.ID)
	})

	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: nonAdminAcc,
	})
	err = s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	path := fmt.Sprintf("/v1/orgs/current/invites/%s/revoke", invite.ID)
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer([]byte{}))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusForbidden, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "only org admins can revoke invites")
}

func (s *RevokeOrgInviteTestSuite) TestRevokeOrgInvite_ReInviteAfterRevoke() {
	email := fmt.Sprintf("reinvite-%s@test.nuon.co", domains.NewAccountID()[:8])
	invite := s.createPendingInvite(email)

	path := fmt.Sprintf("/v1/orgs/current/invites/%s/revoke", invite.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	newInvite := &app.OrgInvite{
		Email:    email,
		OrgID:    s.testOrg.ID,
		Status:   app.OrgInviteStatusPending,
		RoleType: app.RoleTypeOrgAdmin,
	}
	err := s.service.DB.WithContext(ctx).Create(newInvite).Error
	require.NoError(s.T(), err, "should be able to re-invite the same email after revocation")
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", newInvite.ID)
	})

	assert.NotEqual(s.T(), invite.ID, newInvite.ID)
	assert.Equal(s.T(), app.OrgInviteStatusPending, newInvite.Status)
}
