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
	"github.com/lib/pq"
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
)

// AdminAddTagsTestService holds all fx-injected dependencies for admin add tags tests.
type AdminAddTagsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
}

// AdminAddTagsTestSuite is the testify suite for the AdminAddTags endpoint.
type AdminAddTagsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service AdminAddTagsTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestAdminAddTagsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminAddTagsTestSuite))
}

func (s *AdminAddTagsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *AdminAddTagsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares (no org context for admin endpoints)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminAddTagsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminAddTagsTestSuite) setupTestData() {
	// Create test account
	testAccID := domains.NewAccountID()
	testAcc := &app.Account{
		ID:          testAccID,
		Email:       fmt.Sprintf("%s@test.nuon.co", testAccID),
		Subject:     fmt.Sprintf("add-tags-%s", testAccID),
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org with account context (required for BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)

	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("add-tags-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *AdminAddTagsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
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

func (s *AdminAddTagsTestSuite) TestAdminAddTags() {
	testCases := []struct {
		name          string
		setupFunc     func() *app.Org
		requestBody   interface{}
		expectedCode  int
		validateFunc  func(*app.Org)
		checkDBFunc   func(*app.Org)
		errorContains string
	}{
		{
			name: "successfully adds tags to org",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-add-tags-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminAddTagsRequest{
				Tags: []string{"enterprise", "beta"},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.ElementsMatch(s.T(), []string{"enterprise", "beta"}, org.Tags)
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.ElementsMatch(s.T(), []string{"enterprise", "beta"}, dbOrg.Tags)
			},
		},
		{
			name: "duplicate tags are ignored",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-duplicate-tags",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Tags: pq.StringArray{"enterprise", "beta"},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminAddTagsRequest{
				Tags: []string{"enterprise", "beta"},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.ElementsMatch(s.T(), []string{"enterprise", "beta"}, org.Tags)
				assert.Len(s.T(), org.Tags, 2) // No duplicates
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.ElementsMatch(s.T(), []string{"enterprise", "beta"}, dbOrg.Tags)
				assert.Len(s.T(), dbOrg.Tags, 2)
			},
		},
		{
			name: "adds additional tags to existing",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-additional-tags",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Tags: pq.StringArray{"enterprise"},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminAddTagsRequest{
				Tags: []string{"beta", "production"},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.ElementsMatch(s.T(), []string{"enterprise", "beta", "production"}, org.Tags)
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.ElementsMatch(s.T(), []string{"enterprise", "beta", "production"}, dbOrg.Tags)
			},
		},
		{
			name: "fails with empty tags array",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody: AdminAddTagsRequest{
				Tags: []string{},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid request",
		},
		{
			name: "fails with invalid JSON",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody:   "invalid-json-string",
			expectedCode:  http.StatusBadRequest,
			errorContains: "unable to parse request",
		},
		{
			name: "fails when org not found",
			setupFunc: func() *app.Org {
				// Return org with ID that doesn't exist
				return &app.Org{
					ID:          domains.NewOrgID(), // Non-existent org ID
					Name:        "nonexistent",
					SandboxMode: true,
				}
			},
			requestBody: AdminAddTagsRequest{
				Tags: []string{"test"},
			},
			expectedCode:  http.StatusNotFound,
			errorContains: "org not found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test org
			org := tc.setupFunc()
			require.NotNil(s.T(), org)

			// Make request with org_id path parameter
			path := "/v1/orgs/" + org.ID + "/admin-add-tags"
			rr := s.makeRequest(http.MethodPost, path, tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// For success cases
			if tc.expectedCode == http.StatusOK {
				// Parse response
				var response app.Org
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)

				// Run validations
				if tc.validateFunc != nil {
					tc.validateFunc(&response)
				}

				// Check database state
				if tc.checkDBFunc != nil {
					tc.checkDBFunc(&response)
				}
			}

			// For error cases
			if tc.errorContains != "" {
				body := rr.Body.String()
				assert.Contains(s.T(), body, tc.errorContains,
					"Expected error message to contain: %s", tc.errorContains)
			}
		})
	}
}

// AdminRemoveTagsTestService holds all fx-injected dependencies for admin remove tags tests.
type AdminRemoveTagsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
}

// AdminRemoveTagsTestSuite is the testify suite for the AdminRemoveTags endpoint.
type AdminRemoveTagsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service AdminRemoveTagsTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestAdminRemoveTagsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminRemoveTagsTestSuite))
}

func (s *AdminRemoveTagsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *AdminRemoveTagsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares (no org context for admin endpoints)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminRemoveTagsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminRemoveTagsTestSuite) setupTestData() {
	// Create test account
	testAccID := domains.NewAccountID()
	testAcc := &app.Account{
		ID:          testAccID,
		Email:       fmt.Sprintf("%s@test.nuon.co", testAccID),
		Subject:     fmt.Sprintf("rm-tags-%s", testAccID),
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org with account context (required for BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)

	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("rm-tags-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *AdminRemoveTagsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
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

func (s *AdminRemoveTagsTestSuite) TestAdminRemoveTags() {
	testCases := []struct {
		name          string
		setupFunc     func() *app.Org
		requestBody   interface{}
		expectedCode  int
		validateFunc  func(*app.Org)
		checkDBFunc   func(*app.Org)
		errorContains string
	}{
		{
			name: "successfully removes tags from org",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-remove-tags-org",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Tags: pq.StringArray{"enterprise", "beta", "production"},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminRemoveTagsRequest{
				Tags: []string{"beta"},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.ElementsMatch(s.T(), []string{"enterprise", "production"}, org.Tags)
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.ElementsMatch(s.T(), []string{"enterprise", "production"}, dbOrg.Tags)
			},
		},
		{
			name: "removing non-existent tags succeeds silently",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-nonexistent-tags",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Tags: pq.StringArray{"enterprise"},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminRemoveTagsRequest{
				Tags: []string{"nonexistent"},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.ElementsMatch(s.T(), []string{"enterprise"}, org.Tags)
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.ElementsMatch(s.T(), []string{"enterprise"}, dbOrg.Tags)
			},
		},
		{
			name: "removes all tags",
			setupFunc: func() *app.Org {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				org := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "test-remove-all-tags",
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/test",
					},
					Tags: pq.StringArray{"a", "b"},
				}
				err := s.service.DB.WithContext(ctx).Create(org).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
				})

				return org
			},
			requestBody: AdminRemoveTagsRequest{
				Tags: []string{"a", "b"},
			},
			expectedCode: http.StatusOK,
			validateFunc: func(org *app.Org) {
				require.NotNil(s.T(), org)
				assert.Empty(s.T(), org.Tags)
			},
			checkDBFunc: func(org *app.Org) {
				var dbOrg app.Org
				err := s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
				require.NoError(s.T(), err)
				assert.Empty(s.T(), dbOrg.Tags)
			},
		},
		{
			name: "fails with empty tags array",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody: AdminRemoveTagsRequest{
				Tags: []string{},
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid request",
		},
		{
			name: "fails with invalid JSON",
			setupFunc: func() *app.Org {
				return s.testOrg
			},
			requestBody:   "invalid-json-string",
			expectedCode:  http.StatusBadRequest,
			errorContains: "unable to parse request",
		},
		{
			name: "fails when org not found",
			setupFunc: func() *app.Org {
				// Return org with ID that doesn't exist
				return &app.Org{
					ID:          domains.NewOrgID(), // Non-existent org ID
					Name:        "nonexistent",
					SandboxMode: true,
				}
			},
			requestBody: AdminRemoveTagsRequest{
				Tags: []string{"test"},
			},
			expectedCode:  http.StatusNotFound,
			errorContains: "org not found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test org
			org := tc.setupFunc()
			require.NotNil(s.T(), org)

			// Make request with org_id path parameter
			path := "/v1/orgs/" + org.ID + "/admin-remove-tags"
			rr := s.makeRequest(http.MethodPost, path, tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			// For success cases
			if tc.expectedCode == http.StatusOK {
				// Parse response
				var response app.Org
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
				}
				require.NoError(s.T(), err)

				// Run validations
				if tc.validateFunc != nil {
					tc.validateFunc(&response)
				}

				// Check database state
				if tc.checkDBFunc != nil {
					tc.checkDBFunc(&response)
				}
			}

			// For error cases
			if tc.errorContains != "" {
				body := rr.Body.String()
				assert.Contains(s.T(), body, tc.errorContains,
					"Expected error message to contain: %s", tc.errorContains)
			}
		})
	}
}
