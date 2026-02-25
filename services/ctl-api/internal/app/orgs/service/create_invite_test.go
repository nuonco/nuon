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
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// CreateOrgInviteTestService holds all fx-injected dependencies for create org invite tests.
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

// CreateOrgInviteTestSuite is the testify suite for create org invite endpoint.
type CreateOrgInviteTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      CreateOrgInviteTestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
	orgsService  *service
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

func (s *CreateOrgInviteTestSuite) SetupTest() {
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
				return CreateOrgInviteRequest{
					Email: "invite@example.com",
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), invite.ID)
				assert.Equal(s.T(), "invite@example.com", invite.Email)
				assert.Equal(s.T(), app.OrgInviteStatusPending, invite.Status)
				assert.Equal(s.T(), app.RoleTypeOrgAdmin, invite.RoleType)
				assert.Equal(s.T(), s.testOrg.ID, invite.OrgID)
				assert.NotEmpty(s.T(), invite.CreatedByID)
			},
			validateSignal: true,
			validateDB: func(invite *app.OrgInvite) {
				var dbInvite app.OrgInvite
				err := s.service.DB.Where("email = ?", "invite@example.com").First(&dbInvite).Error
				require.NoError(s.T(), err)

				assert.Equal(s.T(), invite.ID, dbInvite.ID)
				assert.Equal(s.T(), "invite@example.com", dbInvite.Email)
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
				// Return nil to signal we'll send raw invalid JSON
				return nil
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				// BindJSON returns JSON parsing errors with "invalid character" message
				assert.Contains(s.T(), body, "invalid")
			},
			validateSignal: false,
		},
		{
			name: "accepts valid email with subdomain",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "user@subdomain.example.com",
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "user@subdomain.example.com", invite.Email)
			},
			validateSignal: true,
		},
		{
			name: "accepts valid email with plus addressing",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "user+tag@example.com",
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "user+tag@example.com", invite.Email)
			},
			validateSignal: true,
		},
		{
			name: "accepts valid email with dots and numbers",
			setupFunc: func() interface{} {
				return CreateOrgInviteRequest{
					Email: "user.name123@example.co.uk",
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(rr *httptest.ResponseRecorder) {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "user.name123@example.co.uk", invite.Email)
			},
			validateSignal: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

			// Setup test data
			reqBody := tc.setupFunc()

			// Make request
			var rr *httptest.ResponseRecorder
			if tc.name == "invalid JSON handling" {
				// Send invalid JSON for that specific test
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
			} else {
				assert.Len(s.T(), signals, 0, "no signal should be sent for failed validation")
			}

			// Validate database state for successful creations
			if tc.validateDB != nil && rr.Code == http.StatusCreated {
				var invite app.OrgInvite
				err := json.Unmarshal(rr.Body.Bytes(), &invite)
				require.NoError(s.T(), err)

				tc.validateDB(&invite)

				// Cleanup
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite.ID)
				})
			}
		})
	}
}

func (s *CreateOrgInviteTestSuite) TestCreateOrgInvite_SetsCorrectDefaults() {
	// This test specifically validates default values
	req := CreateOrgInviteRequest{
		Email: "defaults@example.com",
	}

	rr := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var invite app.OrgInvite
	err := json.Unmarshal(rr.Body.Bytes(), &invite)
	require.NoError(s.T(), err)

	// Verify all default values
	assert.Equal(s.T(), app.OrgInviteStatusPending, invite.Status, "status should default to pending")
	assert.Equal(s.T(), app.RoleTypeOrgAdmin, invite.RoleType, "role_type should default to org_admin")
	assert.Equal(s.T(), s.testOrg.ID, invite.OrgID, "org_id should be set from context")
	assert.Equal(s.T(), s.testAcc.ID, invite.CreatedByID, "created_by_id should be set from context")
	assert.NotEmpty(s.T(), invite.ID, "ID should be generated")
	assert.NotZero(s.T(), invite.CreatedAt, "created_at should be set")
	assert.NotZero(s.T(), invite.UpdatedAt, "updated_at should be set")

	// Cleanup
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite.ID)
	})
}

func (s *CreateOrgInviteTestSuite) TestCreateOrgInvite_UniqueConstraint() {
	// Create first invite
	req := CreateOrgInviteRequest{
		Email: "unique@example.com",
	}

	rr1 := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)
	require.Equal(s.T(), http.StatusCreated, rr1.Code)

	var invite1 app.OrgInvite
	err := json.Unmarshal(rr1.Body.Bytes(), &invite1)
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite1.ID)
	})

	// Try to create second invite with same email in same org
	rr2 := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)

	// Should fail due to unique constraint (org_id, email, deleted_at)
	// Stderr middleware automatically returns 409 Conflict for duplicate keys
	assert.Equal(s.T(), http.StatusConflict, rr2.Code)
	body := rr2.Body.String()
	assert.Contains(s.T(), body, "duplicate key")
}

func (s *CreateOrgInviteTestSuite) TestCreateOrgInvite_DifferentOrgsCanInviteSameEmail() {
	// Create second account and org
	acc2 := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "other@example.com",
		Subject:     "other-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(acc2).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
	})

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, acc2)
	org2 := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "other-org",
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

	// Create invite in first org
	req := CreateOrgInviteRequest{
		Email: "shared@example.com",
	}
	rr1 := s.makeRequest(http.MethodPost, "/v1/orgs/current/invites", req)
	require.Equal(s.T(), http.StatusCreated, rr1.Code)

	var invite1 app.OrgInvite
	err = json.Unmarshal(rr1.Body.Bytes(), &invite1)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite1.ID)
	})

	// Create invite in second org with same email (should succeed)
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
	assert.Equal(s.T(), "shared@example.com", invite2.Email)
	assert.Equal(s.T(), org2.ID, invite2.OrgID)
	assert.NotEqual(s.T(), invite1.ID, invite2.ID)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.OrgInvite{}, "id = ?", invite2.ID)
	})
}
