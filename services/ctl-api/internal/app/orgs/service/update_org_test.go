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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// UpdateOrgTestService holds all fx-injected dependencies for update org tests.
type UpdateOrgTestService struct {
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

// UpdateOrgTestSuite is the testify suite for update org endpoint.
type UpdateOrgTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     UpdateOrgTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestUpdateOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(UpdateOrgTestSuite))
}

func (s *UpdateOrgTestSuite) SetupSuite() {
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

func (s *UpdateOrgTestSuite) SetupTest() {
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

func (s *UpdateOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateOrgTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *UpdateOrgTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *UpdateOrgTestSuite) TestUpdateOrg() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Org
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(*app.Org)
	}{
		{
			name: "successfully updates org name",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "original-name",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: UpdateOrgRequest{
				Name: "updated-name",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org) {
				// Verify database state
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "updated-name", dbOrg.Name)
			},
		},
		{
			name: "validation error when name is empty",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody: UpdateOrgRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(org *app.Org) {
				// Verify org was NOT updated in database
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", s.testOrg.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testOrg.Name, dbOrg.Name, "org name should not have changed")
			},
		},
		{
			name: "validation error when name is missing",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody: map[string]interface{}{
				// name field intentionally omitted
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(org *app.Org) {
				// Verify org was NOT updated in database
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", s.testOrg.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testOrg.Name, dbOrg.Name, "org name should not have changed")
			},
		},
		{
			name: "only updates current org not other orgs",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create a second org
				otherOrgID := domains.NewOrgID()
				otherOrg := &app.Org{
					ID:          otherOrgID,
					Name:        fmt.Sprintf("other-org-%s", otherOrgID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(otherOrg).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrg.ID)
				})

				return s.testOrg
			},
			requestBody: UpdateOrgRequest{
				Name: "updated-current-org",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org) {
				// Verify current org was updated
				var currentOrg app.Org
				err := s.service.DB.First(&currentOrg, "id = ?", s.testOrg.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "updated-current-org", currentOrg.Name)

				// Verify other org was NOT updated (still has its original name)
				var otherOrg app.Org
				err = s.service.DB.Where("name LIKE ?", "other-org-%").First(&otherOrg).Error
				require.NoError(s.T(), err)
				assert.Contains(s.T(), otherOrg.Name, "other-org-", "other org should not be affected")
			},
		},
		{
			name: "updates name with special characters",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody: UpdateOrgRequest{
				Name: "Test Org - Production (2024)",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org) {
				// Verify database state (special characters preserved)
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "Test Org - Production (2024)", dbOrg.Name)
			},
		},
		{
			name: "updates name with unicode characters",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody: UpdateOrgRequest{
				Name: "組織名前 🚀",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org) {
				// Verify database state (unicode characters preserved)
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "組織名前 🚀", dbOrg.Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Always recreate router with correct org context for this test case
			s.router = tests.NewTestRouter(tests.RouterOptions{
				L:       s.service.L,
				DB:      s.service.DB,
				TestOrg: org,
				TestAcc: s.testAcc,
			})
			err := s.orgsService.RegisterPublicRoutes(s.router)
			require.NoError(s.T(), err)

			// Make request
			rr := s.makeRequest(http.MethodPatch, "/v1/orgs/current", tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Parse response for successful updates
			if tc.expectedStatus == http.StatusOK {
				var response app.Org
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)

				// Run validation function
				if tc.validateFunc != nil {
					tc.validateFunc(&response)
				}
			} else if tc.validateFunc != nil {
				// Run validation for non-success cases (verify no changes)
				tc.validateFunc(nil)
			}
		})
	}
}

func (s *UpdateOrgTestSuite) TestUpdateOrgInvalidJSON() {
	// Test with malformed JSON
	reqBody := bytes.NewBufferString(`{"name": "test", invalid json}`)

	req, err := http.NewRequest(http.MethodPatch, "/v1/orgs/current", reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)

	// Verify org was NOT updated in database
	var dbOrg app.Org
	err = s.service.DB.First(&dbOrg, "id = ?", s.testOrg.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testOrg.Name, dbOrg.Name, "org name should not have changed")
}

func (s *UpdateOrgTestSuite) TestUpdateOrgNonExistentOrg() {
	// Create a non-existent org ID for context
	nonExistentOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "non-existent",
		SandboxMode: true,
	}

	// Create router with non-existent org context
	router := tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: nonExistentOrg,
		TestAcc: s.testAcc,
	})
	err := s.orgsService.RegisterPublicRoutes(router)
	require.NoError(s.T(), err)

	// Try to update non-existent org
	reqBody := UpdateOrgRequest{
		Name: "updated-name",
	}
	jsonBytes, err := json.Marshal(reqBody)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPatch, "/v1/orgs/current", bytes.NewBuffer(jsonBytes))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *UpdateOrgTestSuite) TestUpdateOrgWhitespaceHandling() {
	testCases := []struct {
		name           string
		orgName        string
		expectedStatus int
		expectedName   string
	}{
		{
			name:           "preserves leading whitespace",
			orgName:        "  leading-space",
			expectedStatus: http.StatusOK,
			expectedName:   "  leading-space",
		},
		{
			name:           "preserves trailing whitespace",
			orgName:        "trailing-space  ",
			expectedStatus: http.StatusOK,
			expectedName:   "trailing-space  ",
		},
		{
			name:           "preserves internal whitespace",
			orgName:        "name   with   spaces",
			expectedStatus: http.StatusOK,
			expectedName:   "name   with   spaces",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := UpdateOrgRequest{
				Name: tc.orgName,
			}
			rr := s.makeRequest(http.MethodPatch, "/v1/orgs/current", req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response app.Org
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tc.expectedName, response.Name)

				// Verify database state
				var dbOrg app.Org
				err = s.service.DB.First(&dbOrg, "id = ?", s.testOrg.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tc.expectedName, dbOrg.Name)
			}
		})
	}
}
