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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// CreateOrgTestSuite is the testify suite for create org endpoint.
type CreateOrgTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      TestService
	router       *gin.Engine
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
	orgsService  *service
}

func TestCreateOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CreateOrgTestSuite))
}

func (s *CreateOrgTestSuite) SetupSuite() {
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

func (s *CreateOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.orgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *CreateOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateOrgTestSuite) setupTestData() {
	ctx := context.Background()
	_, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
}

func (s *CreateOrgTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *CreateOrgTestSuite) TestCreateOrg() {
	// Generate unique names for each test case to avoid cross-run collisions
	minimalName := fmt.Sprintf("test-org-minimal-%s", domains.NewOrgID()[:8])
	sandboxName := fmt.Sprintf("test-org-sandbox-%s", domains.NewOrgID()[:8])
	notificationsName := fmt.Sprintf("test-org-notifications-%s", domains.NewOrgID()[:8])

	testCases := []struct {
		name           string
		request        CreateOrgRequest
		expectedStatus int
		validateFunc   func(*app.Org)
		validateSignal bool
	}{
		{
			name: "successfully creates org with minimal data",
			request: CreateOrgRequest{
				Name: minimalName,
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), minimalName, org.Name)
				// OrgType has json:"-" tag, so it's not returned in API response
				// Check database record instead
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), app.OrgTypeDefault, dbOrg.OrgType)
				require.False(s.T(), org.SandboxMode)
				require.NotEmpty(s.T(), org.ID)
				require.NotEmpty(s.T(), org.CreatedByID)
				require.Equal(s.T(), s.testAcc.ID, org.CreatedByID)

				// Cleanup
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})
			},
			validateSignal: true,
		},
		{
			name: "successfully creates org with sandbox mode",
			request: CreateOrgRequest{
				Name:           sandboxName,
				UseSandboxMode: true,
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), sandboxName, org.Name)
				// OrgType has json:"-" tag, so it's not returned in API response
				// Check database record instead
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), app.OrgTypeSandbox, dbOrg.OrgType)
				require.True(s.T(), org.SandboxMode)

				// Cleanup
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})
			},
			validateSignal: true,
		},
		{
			name: "successfully creates org with notifications config",
			request: CreateOrgRequest{
				Name: notificationsName,
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(org *app.Org) {
				require.Equal(s.T(), notificationsName, org.Name)

				// Verify notifications config was created
				var notifConfig app.NotificationsConfig
				err := s.service.DB.Where("owner_id = ?", org.ID).First(&notifConfig).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), org.ID, notifConfig.OrgID)
				require.True(s.T(), notifConfig.EnableSlackNotifications)
				require.True(s.T(), notifConfig.EnableEmailNotifications)

				// Cleanup
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})
			},
			validateSignal: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/orgs", tc.request)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response
			var response app.Org
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Verify org exists in database
			var dbOrg app.Org
			err = s.service.DB.First(&dbOrg, "id = ?", response.ID).Error
			require.NoError(s.T(), err)
			require.Equal(s.T(), tc.request.Name, dbOrg.Name)

			// Run validation
			if tc.validateFunc != nil {
				tc.validateFunc(&response)
			}

			// Validate signals were sent
			if tc.validateSignal {
				signals := s.mockEvClient.GetSignals()
				// 3 signals total: OperationCreated (runner), OperationCreated (org), OperationProvision (org)
				// Order: runner creation happens first in CreateOrgRunnerGroup, then org signals
				require.Len(s.T(), signals, 3, "expected runner and org signals")

				// Find org signals (skip runner signal at index 0)
				var orgCreatedSignal, orgProvisionSignal *sigs.Signal
				for _, sig := range signals {
					if sig.ID == response.ID {
						if s, ok := sig.Signal.(*sigs.Signal); ok {
							if s.Type == sigs.OperationCreated {
								orgCreatedSignal = s
							} else if s.Type == sigs.OperationProvision {
								orgProvisionSignal = s
							}
						}
					}
				}

				// Verify both org signals were sent
				require.NotNil(s.T(), orgCreatedSignal, "expected OperationCreated signal for org")
				require.NotNil(s.T(), orgProvisionSignal, "expected OperationProvision signal for org")
			}
		})
	}
}

func (s *CreateOrgTestSuite) TestCreateOrgValidation() {
	testCases := []struct {
		name           string
		request        CreateOrgRequest
		expectedStatus int
	}{
		{
			name: "fails with empty name",
			request: CreateOrgRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/orgs", tc.request)

			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// No signals should be sent on validation failure
			signals := s.mockEvClient.GetSignals()
			assert.Len(s.T(), signals, 0)
		})
	}
}

func (s *CreateOrgTestSuite) TestCreateOrgServiceAccountRestriction() {
	// Create service account
	serviceAccID := domains.NewAccountID()
	serviceAcc := &app.Account{
		ID:          serviceAccID,
		Email:       fmt.Sprintf("%s@test.nuon.co", serviceAccID),
		Subject:     "service-subject",
		AccountType: app.AccountTypeService,
	}
	err := s.service.DB.Create(serviceAcc).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", serviceAcc.ID)
	})

	// Create router with service account context
	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: serviceAcc,
	})
	err = s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	// Make request
	request := CreateOrgRequest{
		Name: fmt.Sprintf("service-org-%s", domains.NewOrgID()[:8]),
	}
	jsonBytes, err := json.Marshal(request)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, "/v1/orgs", bytes.NewBuffer(jsonBytes))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should fail for service accounts
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	require.Contains(s.T(), rr.Body.String(), "not allowed to create new orgs")
}

func (s *CreateOrgTestSuite) TestCreateOrgCreatesRunnerGroup() {
	request := CreateOrgRequest{
		Name: fmt.Sprintf("test-org-runner-group-%s", domains.NewOrgID()[:8]),
	}

	// Reset mock
	s.mockEvClient.Reset()

	// Make request
	rr := s.makeRequest(http.MethodPost, "/v1/orgs", request)

	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Parse response
	var response app.Org
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", response.ID)
	})

	// Verify runner group was created
	var runnerGroup app.RunnerGroup
	err = s.service.DB.Where("owner_id = ? AND owner_type = ?", response.ID, "orgs").First(&runnerGroup).Error
	require.NoError(s.T(), err)
	require.Equal(s.T(), response.ID, runnerGroup.OwnerID)
}

func (s *CreateOrgTestSuite) TestCreateOrgCreatesRoles() {
	request := CreateOrgRequest{
		Name: fmt.Sprintf("test-org-roles-%s", domains.NewOrgID()[:8]),
	}

	// Reset mock
	s.mockEvClient.Reset()

	// Make request with account context
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	rr := s.makeRequest(http.MethodPost, "/v1/orgs", request)

	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Parse response
	var response app.Org
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", response.ID)
	})

	// Verify roles were created
	var roles []app.Role
	err = s.service.DB.Where("org_id = ?", response.ID).Find(&roles).Error
	require.NoError(s.T(), err)
	require.GreaterOrEqual(s.T(), len(roles), 1, "at least one role should be created for the org")

	// Verify user was assigned to org admin role
	var accountRoles []app.AccountRole
	err = s.service.DB.Joins("JOIN roles ON roles.id = account_roles.role_id").
		Where("account_roles.account_id = ? AND roles.org_id = ?", s.testAcc.ID, response.ID).
		Find(&accountRoles).Error
	require.NoError(s.T(), err)
	require.GreaterOrEqual(s.T(), len(accountRoles), 1, "user should be assigned to at least one org role")
}

func (s *CreateOrgTestSuite) TestCreateOrgIntegrationAccountType() {
	// Create integration account
	integrationAccID := domains.NewAccountID()
	integrationAcc := &app.Account{
		ID:          integrationAccID,
		Email:       fmt.Sprintf("%s@test.nuon.co", integrationAccID),
		Subject:     "integration-subject",
		AccountType: app.AccountTypeIntegration,
	}
	err := s.service.DB.Create(integrationAcc).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", integrationAcc.ID)
	})

	// Create router with integration account context
	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: integrationAcc,
	})
	err = s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	// Make request
	request := CreateOrgRequest{
		Name: fmt.Sprintf("integration-org-%s", domains.NewOrgID()[:8]),
	}
	jsonBytes, err := json.Marshal(request)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, "/v1/orgs", bytes.NewBuffer(jsonBytes))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	// Reset mock
	s.mockEvClient.Reset()

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Parse response
	var response app.Org
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", response.ID)
	})

	// Verify org type is integration (check database since OrgType has json:"-")
	var dbOrg app.Org
	err = s.service.DB.First(&dbOrg, "id = ?", response.ID).Error
	require.NoError(s.T(), err)
	require.Equal(s.T(), app.OrgTypeIntegration, dbOrg.OrgType)
}

func (s *CreateOrgTestSuite) TestCreateOrgDuplicateName() {
	// Create first org
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	dupName := fmt.Sprintf("duplicate-org-%s", domains.NewOrgID()[:8])
	existingOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        dupName,
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
	}
	err := s.service.DB.WithContext(ctx).Create(existingOrg).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", existingOrg.ID)
	})

	// Try to create org with same name
	request := CreateOrgRequest{
		Name: dupName,
	}

	rr := s.makeRequest(http.MethodPost, "/v1/orgs", request)

	// Should fail with conflict/error
	require.Equal(s.T(), http.StatusConflict, rr.Code)
	s.T().Logf("Duplicate org creation status: %d, Body: %s", rr.Code, rr.Body.String())
}

func (s *CreateOrgTestSuite) TestCreateOrgWithoutAccountContext() {
	// Create router without account context
	router := tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
		// TestAcc intentionally omitted
	})

	err := s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	request := CreateOrgRequest{
		Name: fmt.Sprintf("no-context-org-%s", domains.NewOrgID()[:8]),
	}
	jsonBytes, err := json.Marshal(request)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, "/v1/orgs", bytes.NewBuffer(jsonBytes))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusInternalServerError, rr.Code, "should not successfully create org without account context")
	s.T().Logf("Status without account context: %d, Body: %s", rr.Code, rr.Body.String())
}
