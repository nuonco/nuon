package service

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// AppCRUDTestSuite tests GetApp, UpdateApp, and DeleteApp endpoints.
type AppCRUDTestSuite struct {
	tests.BaseDBTestSuite

	app          *fxtest.App
	service      TestService
	router       *gin.Engine
	testOrg      *app.Org
	testAcc      *app.Account
	mockEvClient *tests.MockEventLoopClient
}

func TestAppCRUDSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AppCRUDTestSuite))
}

func (s *AppCRUDTestSuite) SetupSuite() {
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
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AppCRUDTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Reset mock before each test
	s.mockEvClient.Reset()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AppCRUDTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AppCRUDTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *AppCRUDTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// TestGetApp tests the GetApp endpoint.
func (s *AppCRUDTestSuite) TestGetApp() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(*models.AppApp)
	}{
		{
			name: "get app by ID returns 200 with correct data",
			setupFunc: func() string {
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-get",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusProvisioning,
				}
				err := s.service.DB.Create(testApp).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				return testApp.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(appResp *models.AppApp) {
				assert.Equal(s.T(), "test-app-get", appResp.Name)
				assert.Equal(s.T(), s.testOrg.ID, appResp.OrgID)
			},
		},
		{
			name: "get app by name returns 200",
			setupFunc: func() string {
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-by-name",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusProvisioning,
				}
				err := s.service.DB.Create(testApp).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				return testApp.Name
			},
			expectedCode: http.StatusOK,
			validateFunc: func(appResp *models.AppApp) {
				assert.Equal(s.T(), "test-app-by-name", appResp.Name)
			},
		},
		{
			name: "get non-existent app returns 404",
			setupFunc: func() string {
				return domains.NewAppID()
			},
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
		{
			name: "get app from different org returns 404",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				otherOrg := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org-" + domains.NewOrgID(),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(otherOrg).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrg.ID)
				})

				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "other-org-app",
					OrgID:       otherOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusProvisioning,
				}
				err = s.service.DB.Create(testApp).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				return testApp.ID
			},
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			appID := tc.setupFunc()
			rr := s.makeRequest(http.MethodGet, "/v1/apps/"+appID, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				var response models.AppApp
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				tc.validateFunc(&response)
			}
		})
	}
}
