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
	orginvitecreated "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/invite_created"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type CreateOrgInviteTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	Seeder          *testseed.Seeder
}

type CreateOrgInviteTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     CreateOrgInviteTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestCreateOrgInviteSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CreateOrgInviteTestSuite))
}

func (s *CreateOrgInviteTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *CreateOrgInviteTestSuite) SetupTest() {
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

func (s *CreateOrgInviteTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateOrgInviteTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *CreateOrgInviteTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *CreateOrgInviteTestSuite) TestCreateOrgInvite() {
	testCases := []struct {
		name             string
		setupFunc        func() interface{}
		expectedStatus   int
		validateResponse func(*httptest.ResponseRecorder)
		validateSignal   bool
		validateDB       func(*app.OrgInvite)
	}{
		{
			name: "successfully creates invite with valid email",
			setupFunc: func() interface{} {
				testEmail := fmt.Sprintf("invite-%s@test.nuon.co", domains.NewAccountID()[:8])
				return CreateOrgInviteRequest{
					Email: testEmail,
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), invite.ID)
				assert.NotEmpty(s.T(), invite.Email)
				assert.Equal(s.T(), app.OrgInviteStatusPending, invite.Status)
				assert.Equal(s.T(), app.RoleTypeOrgAdmin, invite.RoleType)
				assert.Equal(s.T(), s.testOrg.ID, invite.OrgID)
				assert.NotEmpty(s.T(), invite.CreatedByID)
			},
			validateSignal: true,
			validateDB: func(invite *app.OrgInvite) {
				var dbInvite app.OrgInvite
				err := s.service.DB.Where("id = ?", invite.ID).First(&dbInvite).Error
				require.NoError(s.T(), err)

				assert.Equal(s.T(), invite.ID, dbInvite.ID)
				assert.Equal(s.T(), invite.Email, dbInvite.Email)
				assert.Equal(s.T(), app.OrgInviteStatusPending, dbInvite.Status)
				assert.Equal(s.T(), app.RoleTypeOrgAdmin, dbInvite.RoleType)
				assert.Equal(s.T(), s.testOrg.ID, dbInvite.OrgID)
				assert.Equal(s.T(), s.testAcc.ID, dbInvite.CreatedByID)
			},
		},
		{
			name: "validation error when email is empty",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "",
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "email is required")
			},
			validateSignal: false,
		},
		{
			name: "validation error when email is missing from JSON",
			setupFunc: func() interface{} {
				return map[string]string{}
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "email is required")
			},
			validateSignal: false,
		},
		{
			name: "validation error when email format is invalid - no domain",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "invalidemail",
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invalid email")
			},
			validateSignal: false,
		},
		{
			name: "validation error when email format is invalid - no @",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "invalid.example.com",
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invalid email")
			},
			validateSignal: false,
		},
		{
			name: "validation error when email format is invalid - no TLD",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "test@domain",
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invalid email")
			},
			validateSignal: false,
		},
		{
			name: "validation error when email has spaces",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "invalid email@example.com",
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invalid email")
			},
			validateSignal: false,
		},
		{
			name: "invalid JSON handling",
			setupFunc: func() interface{} {
				return nil
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "invalid")
			},
			validateSignal: false,
		},
		{
			name: "accepts valid email with subdomain",
			setupFunc: func() interface{} {
				testEmail := fmt.Sprintf("user-%s@subdomain.test.nuon.co", domains.NewAccountID()[:8])
				return CreateOrgInviteRequest{
					Email: testEmail,
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)
				assert.NotEmpty(s.T(), invite.Email)
			},
			validateSignal: true,
		},
		{
			name: "accepts valid email with plus addressing",
			setupFunc: func() interface{} {
				testEmail := fmt.Sprintf("user+tag-%s@test.nuon.co", domains.NewAccountID()[:8])
				return CreateOrgInviteRequest{
					Email: testEmail,
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)
				assert.NotEmpty(s.T(), invite.Email)
			},
			validateSignal: true,
		},
		{
			name: "accepts valid email with dots and numbers",
			setupFunc: func() interface{} {
				testEmail := fmt.Sprintf("user.name123-%s@test.nuon.co", domains.NewAccountID()[:8])
				return CreateOrgInviteRequest{
					Email: testEmail,
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)
				assert.NotEmpty(s.T(), invite.Email)
			},
			validateSignal: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			reqBody := tc.setupFunc()

			var rr *httptest.ResponseRecorder
			if tc.name == "invalid JSON handling" {
				req, err := http.NewRequest(http.MethodPost, "/v1/orgs/current/invites", bytes.NewBufferString("{invalid json}"))
				require.NoError(s.T(), err)
				req.Header.Set("Content-Type", "application/json")
				rr = httptest.NewRecorder()
				s.router.ServeHTTP(rr, req)
			} else {
				rr = s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", reqBody)
			}

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if tc.validateResponse != nil {
				tc.validateResponse(rr)
			}

			if tc.validateSignal {
				signals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), signals, 1, "expected exactly one signal to be sent")

				signal := signals[0]
				assert.Equal(s.T(), s.testOrg.ID, signal.OwnerID, "signal should be sent to correct org ID")
				assert.Equal(s.T(), orginvitecreated.SignalType, signal.Type, "signal type should be OperationInviteCreated")
			} else {
				signals := tests.GetQueueSignals(s.T(), s.service.DB)
				assert.Len(s.T(), signals, 0, "no signal should be sent for failed validation")
			}

			if tc.validateDB != nil && rr.Code == http.StatusCreated {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)

				tc.validateDB(&invite)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite.ID)
				})
			}
		})
	}
}

func (s *CreateOrgInviteTestSuite) TestCreateOrgInvite_SetsCorrectDefaults() {
	testEmail := fmt.Sprintf("defaults-%s@test.nuon.co", domains.NewAccountID()[:8])
	req := CreateOrgInviteRequest{
		Email: testEmail,
	}

	rr := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var invite app.OrgInvite
	err := json.Unmarshal(rr.Body.Bytes(), &invite)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), app.OrgInviteStatusPending, invite.Status, "status should default to pending")
	assert.Equal(s.T(), app.RoleTypeOrgAdmin, invite.RoleType, "role_type should default to org_admin")
	assert.Equal(s.T(), s.testOrg.ID, invite.OrgID, "org_id should be set from context")
	assert.Equal(s.T(), s.testAcc.ID, invite.CreatedByID, "created_by_id should be set from context")
	assert.NotEmpty(s.T(), invite.ID, "ID should be generated")
	assert.NotZero(s.T(), invite.CreatedAt, "created_at should be set")
	assert.NotZero(s.T(), invite.UpdatedAt, "updated_at should be set")

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite.ID)
	})
}

func (s *CreateOrgInviteTestSuite) TestCreateOrgInvite_UniqueConstraint() {
	testEmail := fmt.Sprintf("unique-%s@test.nuon.co", domains.NewAccountID()[:8])
	req := CreateOrgInviteRequest{
		Email: testEmail,
	}

	rr1 := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)
	require.Equal(s.T(), http.StatusCreated, rr1.Code)

	var invite1 app.OrgInvite
	err := json.Unmarshal(rr1.Body.Bytes(), &invite1)
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite1.ID)
	})

	rr2 := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)

	assert.Equal(s.T(), http.StatusConflict, rr2.Code)
	body := rr2.Body.String()
	assert.Contains(s.T(), body, "duplicate key")
}

func (s *CreateOrgInviteTestSuite) TestCreateOrgInvite_DifferentOrgsCanInviteSameEmail() {
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

	testEmail := fmt.Sprintf("shared-%s@test.nuon.co", domains.NewAccountID()[:8])
	req := CreateOrgInviteRequest{
		Email: testEmail,
	}
	rr1 := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)
	require.Equal(s.T(), http.StatusCreated, rr1.Code)

	var invite1 app.OrgInvite
	err = json.Unmarshal(rr1.Body.Bytes(), &invite1)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite1.ID)
	})

	router2 := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: org2,
		TestAcc: acc2,
	})
	err = s.orgsService.RegisterPublicRoutes(router2)
	require.NoError(s.T(), err)

	bodyBytes, err := json.Marshal(req)
	require.NoError(s.T(), err)
	req2, err := http.NewRequest(http.MethodPost, "/v1/orgs/current/invites", bytes.NewBuffer(bodyBytes))
	require.NoError(s.T(), err)
	req2.Header.Set("Content-Type", "application/json")

	rr2 := httptest.NewRecorder()
	router2.ServeHTTP(rr2, req2)

	require.Equal(s.T(), http.StatusCreated, rr2.Code, fmt.Sprintf("Body: %s", rr2.Body.String()))

	var invite2 app.OrgInvite
	err = json.Unmarshal(rr2.Body.Bytes(), &invite2)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), testEmail, invite2.Email)
	assert.Equal(s.T(), org2.ID, invite2.OrgID)
	assert.NotEqual(s.T(), invite1.ID, invite2.ID)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite2.ID)
	})
}
