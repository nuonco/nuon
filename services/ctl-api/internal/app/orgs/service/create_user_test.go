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
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// CreateUserTestService holds all fx-injected dependencies for create user tests.
type CreateUserTestService struct {
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

// CreateUserTestSuite is the testify suite for CreateUser endpoint.
type CreateUserTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     CreateUserTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestCreateUserSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CreateUserTestSuite))
}

func (s *CreateUserTestSuite) SetupSuite() {
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

func (s *CreateUserTestSuite) SetupTest() {
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

func (s *CreateUserTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateUserTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())

	// Create org roles (required for role assignment)
	err := s.service.AuthzClient.CreateOrgRoles(ctx, s.testOrg.ID)
	require.NoError(s.T(), err)
}

func (s *CreateUserTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *CreateUserTestSuite) TestCreateUser() {
	testCases := []struct {
		name             string
		setupFunc        func() *app.Account
		requestBody      interface{}
		expectedStatus   int
		validateFunc     func(*app.Account)
		expectedRoleType app.RoleType
	}{
		{
			name: "successfully adds authenticated user to org",
			setupFunc: func() *app.Account {
				// Create a new account that's not yet in the org
				accID := domains.NewAccountID()
				acc := &app.Account{
					ID:          accID,
					Email:       fmt.Sprintf("%s@test.nuon.co", accID),
					Subject:     "new-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(acc).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc.ID)
				})

				return acc
			},
			requestBody:      CreateOrgUserRequest{UserID: "ignored-field"},
			expectedStatus:   http.StatusCreated,
			expectedRoleType: app.RoleTypeOrgAdmin,
			validateFunc: func(acc *app.Account) {
				// Verify AccountRole was created
				var accountRole app.AccountRole
				err := s.service.DB.
					Where("account_id = ? AND org_id = ?", acc.ID, s.testOrg.ID).
					Preload("Role").
					First(&accountRole).Error
				require.NoError(s.T(), err)

				// Verify role type is org_admin
				assert.Equal(s.T(), app.RoleTypeOrgAdmin, accountRole.Role.RoleType)
				assert.Equal(s.T(), s.testOrg.ID, accountRole.OrgID.String)

				// Verify role belongs to correct org
				assert.Equal(s.T(), s.testOrg.ID, accountRole.Role.OrgID.String)
			},
		},
		{
			name: "returns authenticated account in response",
			setupFunc: func() *app.Account {
				accID := domains.NewAccountID()
				acc := &app.Account{
					ID:          accID,
					Email:       fmt.Sprintf("%s@test.nuon.co", accID),
					Subject:     "response-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(acc).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc.ID)
				})

				return acc
			},
			requestBody:      CreateOrgUserRequest{UserID: "ignored"},
			expectedStatus:   http.StatusCreated,
			expectedRoleType: app.RoleTypeOrgAdmin,
			validateFunc: func(expectedAcc *app.Account) {
				// Response validation happens in the main test loop
			},
		},
		{
			name: "handles empty request body",
			setupFunc: func() *app.Account {
				accID := domains.NewAccountID()
				acc := &app.Account{
					ID:          accID,
					Email:       fmt.Sprintf("%s@test.nuon.co", accID),
					Subject:     "empty-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(acc).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc.ID)
				})

				return acc
			},
			requestBody:      CreateOrgUserRequest{},
			expectedStatus:   http.StatusCreated,
			expectedRoleType: app.RoleTypeOrgAdmin,
			validateFunc: func(acc *app.Account) {
				// Verify role assignment still works
				var accountRole app.AccountRole
				err := s.service.DB.
					Where("account_id = ? AND org_id = ?", acc.ID, s.testOrg.ID).
					First(&accountRole).Error
				require.NoError(s.T(), err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test-specific account
			testAccount := tc.setupFunc()

			// Create router with test-specific account context
			router := tests.NewTestRouter(tests.RouterOptions{
				L:       s.service.L,
				DB:      s.service.DB,
				TestOrg: s.testOrg,
				TestAcc: testAccount,
			})
			err := s.orgsService.RegisterPublicRoutes(router)
			require.NoError(s.T(), err)

			// Make request using the test-specific router
			var reqBody *bytes.Buffer
			if tc.requestBody != nil {
				jsonBytes, err := json.Marshal(tc.requestBody)
				require.NoError(s.T(), err)
				reqBody = bytes.NewBuffer(jsonBytes)
			} else {
				reqBody = bytes.NewBuffer(nil)
			}
			req, err := http.NewRequest(http.MethodPost, "/v1/orgs/current/user", reqBody)
			require.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response - should return the authenticated account
			var response app.Account
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Verify response contains authenticated account data
			assert.Equal(s.T(), testAccount.ID, response.ID)
			assert.Equal(s.T(), testAccount.Email, response.Email)
			assert.Equal(s.T(), testAccount.Subject, response.Subject)

			// Run additional validations
			if tc.validateFunc != nil {
				tc.validateFunc(testAccount)
			}
		})
	}
}

func (s *CreateUserTestSuite) TestCreateUserInvalidJSON() {
	// Test with malformed JSON
	req, err := http.NewRequest(http.MethodPost, "/v1/orgs/current/user", bytes.NewBufferString("{invalid json"))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	// Should return 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}
