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

// AdminRenameOrgTestService holds all fx-injected dependencies for admin rename org tests.
type AdminRenameOrgTestService struct {
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

// AdminRenameOrgTestSuite is the testify suite for admin rename org endpoint.
type AdminRenameOrgTestSuite struct {
	tests.BaseDBTestSuite

	app         *fxtest.App
	service     AdminRenameOrgTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	orgsService *service
}

func TestAdminRenameOrgSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminRenameOrgTestSuite))
}

func (s *AdminRenameOrgTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service, &s.orgsService),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *AdminRenameOrgTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.orgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminRenameOrgTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminRenameOrgTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *AdminRenameOrgTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminRenameOrgTestSuite) TestAdminRenameOrg() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Org
		requestBody    interface{}
		expectedStatus int
		validateFunc   func(*app.Org, *httptest.ResponseRecorder)
	}{
		{
			name: "successfully renames org",
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
			requestBody: RenameOrgRequest{
				Name: "new-org-name",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org, rr *httptest.ResponseRecorder) {
				// Verify HTTP response is true
				var response bool
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				require.True(s.T(), response)

				// Verify database state - name changed
				var updatedOrg app.Org
				err = s.service.DB.Where("id = ?", org.ID).First(&updatedOrg).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), "new-org-name", updatedOrg.Name)
			},
		},
		{
			name: "validation error when name is empty",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "existing-org",
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
			requestBody: RenameOrgRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(org *app.Org, rr *httptest.ResponseRecorder) {
				// Verify database state - name unchanged
				var unchangedOrg app.Org
				err := s.service.DB.Where("id = ?", org.ID).First(&unchangedOrg).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), "existing-org", unchangedOrg.Name)
			},
		},
		{
			name: "validation error when name is missing",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "existing-org",
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
			requestBody: map[string]interface{}{
				// name field intentionally omitted
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(org *app.Org, rr *httptest.ResponseRecorder) {
				// Verify database state - name unchanged
				var unchangedOrg app.Org
				err := s.service.DB.Where("id = ?", org.ID).First(&unchangedOrg).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), "existing-org", unchangedOrg.Name)
			},
		},
		{
			name: "returns error when org_id not found",
			setupFunc: func() *app.Org {
				// Return org with non-existent ID
				return &app.Org{
					ID:          "org_nonexistent_id_12345",
					Name:        "does-not-exist",
					SandboxMode: true,
				}
			},
			requestBody: RenameOrgRequest{
				Name: "attempted-new-name",
			},
			expectedStatus: http.StatusNotFound,
			validateFunc: func(org *app.Org, rr *httptest.ResponseRecorder) {
				// Verify org does not exist in database
				var count int64
				err := s.service.DB.Model(&app.Org{}).Where("id = ?", org.ID).Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(0), count)
			},
		},
		{
			name: "invalid JSON handling",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "existing-org",
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
			requestBody:    "invalid-json-string",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(org *app.Org, rr *httptest.ResponseRecorder) {
				// Verify database state - name unchanged
				var unchangedOrg app.Org
				err := s.service.DB.Where("id = ?", org.ID).First(&unchangedOrg).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), "existing-org", unchangedOrg.Name)
			},
		},
		{
			name: "only updates specified org (not other orgs)",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create target org
				targetOrg := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "target-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(targetOrg).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", targetOrg.ID)
				})

				// Create other org that should NOT be affected
				otherOrg := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err = s.service.DB.WithContext(ctx).Create(otherOrg).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrg.ID)
				})

				// Store other org ID for validation
				s.T().Cleanup(func() {
					// Verify other org name unchanged in final validation
					var unchangedOtherOrg app.Org
					err := s.service.DB.Where("id = ?", otherOrg.ID).First(&unchangedOtherOrg).Error
					require.NoError(s.T(), err)
					require.Equal(s.T(), "other-org", unchangedOtherOrg.Name)
				})

				return targetOrg
			},
			requestBody: RenameOrgRequest{
				Name: "renamed-target-org",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(org *app.Org, rr *httptest.ResponseRecorder) {
				// Verify HTTP response is true
				var response bool
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				require.True(s.T(), response)

				// Verify target org was renamed
				var updatedOrg app.Org
				err = s.service.DB.Where("id = ?", org.ID).First(&updatedOrg).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), "renamed-target-org", updatedOrg.Name)

				// Additional validation of other org happens in cleanup
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			org := tc.setupFunc()

			// Update router context to use the test org
			s.router = tests.NewTestRouter(tests.RouterOptions{
				L:       s.service.L,
				DB:      s.service.DB,
				TestOrg: org,
				TestAcc: s.testAcc,
			})
			err := s.orgsService.RegisterInternalRoutes(s.router)
			require.NoError(s.T(), err)

			// Make request
			rr := s.makeRequest(http.MethodPost, fmt.Sprintf("/v1/orgs/%s/admin-rename", org.ID), tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Run validation
			if tc.validateFunc != nil {
				tc.validateFunc(org, rr)
			}
		})
	}
}
