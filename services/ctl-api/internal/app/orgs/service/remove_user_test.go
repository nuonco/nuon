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
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type RemoveUserTestService struct {
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

type RemoveUserTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     RemoveUserTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestRemoveUserSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(RemoveUserTestSuite))
}

func (s *RemoveUserTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *RemoveUserTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.orgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *RemoveUserTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *RemoveUserTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *RemoveUserTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *RemoveUserTestSuite) TestRemoveUser() {
	testCases := []struct {
		name               string
		setupFunc          func() string
		requestBody        interface{}
		expectedStatus     int
		validateFunc       func(string)
		shouldRemoveRoles  bool
		shouldRemoveInvite bool
	}{
		{
			name: "removes user and associated invite",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				userToRemoveID := domains.NewAccountID()
				userToRemove := &app.Account{
					ID:          userToRemoveID,
					Email:       fmt.Sprintf("%s@test.nuon.co", userToRemoveID),
					Subject:     "invited-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(userToRemove).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", userToRemove.ID)
				})

				err = s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				err = s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, userToRemove.ID)
				require.NoError(s.T(), err)

				invite := &app.OrgInvite{
					OrgID: s.testOrg.ID,
					Email: userToRemove.Email,
				}
				err = s.service.DB.WithContext(ctx).Create(invite).Error
				require.NoError(s.T(), err)

				var inviteCount int64
				err = s.service.DB.Model(&app.OrgInvite{}).
					Where("org_id = ? AND email = ?", s.testOrg.ID, userToRemove.Email).
					Count(&inviteCount).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(1), inviteCount, "invite should exist before removal")

				return userToRemove.ID
			},
			requestBody:        nil,
			expectedStatus:     http.StatusAccepted,
			shouldRemoveRoles:  true,
			shouldRemoveInvite: true,
			validateFunc: func(userID string) {
				var account app.Account
				err := s.service.DB.Select("email").Where("id = ?", userID).First(&account).Error
				require.NoError(s.T(), err)

				var inviteCount int64
				err = s.service.DB.Unscoped().Model(&app.OrgInvite{}).
					Where("org_id = ? AND email = ?", s.testOrg.ID, account.Email).
					Count(&inviteCount).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), inviteCount, "invite should be removed")

				var roleCount int64
				err = s.service.DB.Unscoped().Model(&app.AccountRole{}).
					Where("account_id = ? AND org_id = ?", userID, s.testOrg.ID).
					Count(&roleCount).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), roleCount, "role should be removed")
			},
		},
		{
			name: "removes user with multiple roles",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				userToRemoveID := domains.NewAccountID()
				userToRemove := &app.Account{
					ID:          userToRemoveID,
					Email:       fmt.Sprintf("%s@test.nuon.co", userToRemoveID),
					Subject:     "multirole-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(userToRemove).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", userToRemove.ID)
				})

				err = s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				err = s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, userToRemove.ID)
				require.NoError(s.T(), err)

				err = s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeInstaller, s.testOrg.ID, userToRemove.ID)
				require.NoError(s.T(), err)

				var roleCount int64
				err = s.service.DB.Model(&app.AccountRole{}).
					Where("account_id = ? AND org_id = ?", userToRemove.ID, s.testOrg.ID).
					Count(&roleCount).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(2), roleCount, "should have 2 roles before removal")

				return userToRemove.ID
			},
			requestBody:       nil,
			expectedStatus:    http.StatusAccepted,
			shouldRemoveRoles: true,
			validateFunc: func(userID string) {
				var roleCount int64
				err := s.service.DB.Unscoped().Model(&app.AccountRole{}).
					Where("account_id = ? AND org_id = ?", userID, s.testOrg.ID).
					Count(&roleCount).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), roleCount, "all roles should be removed")
			},
		},
		{
			name: "returns error when user_id is missing",
			setupFunc: func() string {
				return ""
			},
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusNotFound,
			validateFunc:   nil,
		},
		{
			name: "returns error when user_id is empty string",
			setupFunc: func() string {
				return ""
			},
			requestBody: map[string]interface{}{
				"user_id": "",
			},
			expectedStatus: http.StatusNotFound,
			validateFunc:   nil,
		},
		{
			name: "user can be re-invited after removal",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				userToRemoveID := domains.NewAccountID()
				userToRemove := &app.Account{
					ID:          userToRemoveID,
					Email:       fmt.Sprintf("%s@test.nuon.co", userToRemoveID),
					Subject:     "reinvite-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(userToRemove).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", userToRemove.ID)
				})

				err = s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
				require.NoError(s.T(), err)

				err = s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, userToRemove.ID)
				require.NoError(s.T(), err)

				invite := &app.OrgInvite{
					OrgID: s.testOrg.ID,
					Email: userToRemove.Email,
				}
				err = s.service.DB.WithContext(ctx).Create(invite).Error
				require.NoError(s.T(), err)

				return userToRemove.ID
			},
			requestBody:    nil,
			expectedStatus: http.StatusAccepted,
			validateFunc: func(userID string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				var account app.Account
				err := s.service.DB.Select("email").Where("id = ?", userID).First(&account).Error
				require.NoError(s.T(), err)

				var inviteCount int64
				err = s.service.DB.Unscoped().Model(&app.OrgInvite{}).
					Where("org_id = ? AND email = ?", s.testOrg.ID, account.Email).
					Count(&inviteCount).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), inviteCount, "old invite should be removed")

				newInvite := &app.OrgInvite{
					OrgID: s.testOrg.ID,
					Email: account.Email,
				}
				err = s.service.DB.WithContext(ctx).Create(newInvite).Error
				assert.NoError(s.T(), err, "should be able to create new invite after removal")

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "org_id = ? AND email = ?", s.testOrg.ID, account.Email)
				})
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			userID := tc.setupFunc()

			var reqBody interface{}
			if tc.requestBody == nil && userID != "" {
				reqBody = RemoveOrgUserRequest{UserID: userID}
			} else if tc.requestBody != nil {
				reqBody = tc.requestBody
			} else {
				reqBody = RemoveOrgUserRequest{UserID: userID}
			}

			rr := s.makeRequest(http.MethodPost, "/v1/orgs/current/remove-user", reqBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusAccepted {
				var response app.Account
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)
			}

			if tc.validateFunc != nil {
				tc.validateFunc(userID)
			}
		})
	}
}

func (s *RemoveUserTestSuite) TestRemoveUserInvalidJSON() {
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "invalid json structure",
			requestBody:    `{"user_id": 123}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed json",
			requestBody:    `{"user_id": "test"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "completely invalid json",
			requestBody:    `not json at all`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := http.NewRequest(http.MethodPost, "/v1/orgs/current/remove-user", bytes.NewBufferString(tc.requestBody))
			require.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)
		})
	}
}

func (s *RemoveUserTestSuite) TestRemoveUserWithoutOrgContext() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	userToRemoveID := domains.NewAccountID()
	userToRemove := &app.Account{
		ID:          userToRemoveID,
		Email:       fmt.Sprintf("%s@test.nuon.co", userToRemoveID),
		Subject:     "nocontext-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(userToRemove).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", userToRemove.ID)
	})

	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err = s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	reqBody := RemoveOrgUserRequest{UserID: userToRemove.ID}
	jsonData, err := json.Marshal(reqBody)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, "/v1/orgs/current/remove-user", bytes.NewBuffer(jsonData))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusInternalServerError, rr.Code)
}

func (s *RemoveUserTestSuite) TestRemoveUserAcrossOrgs() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	acc2ID := domains.NewAccountID()
	acc2 := &app.Account{
		ID:          acc2ID,
		Email:       fmt.Sprintf("%s@test.nuon.co", acc2ID),
		Subject:     "org2-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(acc2).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
	})

	ctx2 := context.Background()
	ctx2 = cctx.SetAccountContext(ctx2, acc2)

	org2 := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "test-org-2",
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx2).Create(org2).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)
	})

	sharedUserID := domains.NewAccountID()
	sharedUser := &app.Account{
		ID:          sharedUserID,
		Email:       fmt.Sprintf("%s@test.nuon.co", sharedUserID),
		Subject:     "shared-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err = s.service.DB.Create(sharedUser).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", sharedUser.ID)
	})

	err = s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
	require.NoError(s.T(), err)

	err = s.service.AuthzClient.CreateOrgRoles(ctx2, org2.ID)
	require.NoError(s.T(), err)

	err = s.service.AuthzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, sharedUser.ID)
	require.NoError(s.T(), err)

	err = s.service.AuthzClient.AddAccountOrgRole(ctx2, app.RoleTypeOrgAdmin, org2.ID, sharedUser.ID)
	require.NoError(s.T(), err)

	reqBody := RemoveOrgUserRequest{UserID: sharedUser.ID}
	rr := s.makeRequest(http.MethodPost, "/v1/orgs/current/remove-user", reqBody)

	require.Equal(s.T(), http.StatusAccepted, rr.Code)

	var org1RoleCount int64
	err = s.service.DB.Model(&app.AccountRole{}).
		Where("account_id = ? AND org_id = ?", sharedUser.ID, s.testOrg.ID).
		Count(&org1RoleCount).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), org1RoleCount, "user should be removed from first org")

	var org2RoleCount int64
	err = s.service.DB.Model(&app.AccountRole{}).
		Where("account_id = ? AND org_id = ?", sharedUser.ID, org2.ID).
		Count(&org2RoleCount).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), org2RoleCount, "user should still have access to second org")
}
