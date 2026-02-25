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

// AppConfigsTestSuite is the testify suite for app config endpoints.
type AppConfigsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service TestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestAppConfigsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AppConfigsTestSuite))
}

func (s *AppConfigsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AppConfigsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AppConfigsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AppConfigsTestSuite) setupTestData() {
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "test-org-" + domains.NewOrgID(),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg

	testApp := &app.App{
		ID:          domains.NewAppID(),
		Name:        "test-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	err = s.service.DB.Create(testApp).Error
	require.NoError(s.T(), err)
	s.testApp = testApp
}

func (s *AppConfigsTestSuite) makeGetRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AppConfigsTestSuite) makeRequestWithBody(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// TestGetAppConfigsReturnsEmptyArrayWhenNoConfigs tests GET /v1/apps/:app_id/configs with no configs.
func (s *AppConfigsTestSuite) TestGetAppConfigsReturnsEmptyArrayWhenNoConfigs() {
	path := fmt.Sprintf("/v1/apps/%s/configs", s.testApp.ID)
	rr := s.makeGetRequest(http.MethodGet, path)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []*models.AppAppConfig
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), response)
	require.Len(s.T(), response, 0)
}

func (s *AppConfigsTestSuite) TestGetAppConfigsReturnsConfigs() {
	testCases := []struct {
		name          string
		setupFunc     func() []string
		expectedCount int
	}{
		{
			name: "returns single config",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				cfg := &app.AppConfig{
					ID:                domains.NewAppCfgID(),
					OrgID:             s.testOrg.ID,
					AppID:             s.testApp.ID,
					Status:            app.AppConfigStatusPending,
					StatusDescription: "pending",
					Readme:            "test readme",
					CLIVersion:        "1.0.0",
				}
				err := s.service.DB.WithContext(ctx).Create(cfg).Error
				require.NoError(s.T(), err)
				return []string{cfg.ID}
			},
			expectedCount: 1,
		},
		{
			name: "returns multiple configs",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				cfg1 := &app.AppConfig{
					ID:                domains.NewAppCfgID(),
					OrgID:             s.testOrg.ID,
					AppID:             s.testApp.ID,
					Status:            app.AppConfigStatusPending,
					StatusDescription: "pending",
					Readme:            "test readme 1",
					CLIVersion:        "1.0.0",
				}
				cfg2 := &app.AppConfig{
					ID:                domains.NewAppCfgID(),
					OrgID:             s.testOrg.ID,
					AppID:             s.testApp.ID,
					Status:            app.AppConfigStatusActive,
					StatusDescription: "success",
					Readme:            "test readme 2",
					CLIVersion:        "1.1.0",
				}
				err := s.service.DB.WithContext(ctx).Create(cfg1).Error
				require.NoError(s.T(), err)
				err = s.service.DB.WithContext(ctx).Create(cfg2).Error
				require.NoError(s.T(), err)
				return []string{cfg1.ID, cfg2.ID}
			},
			expectedCount: 2,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			configIDs := tc.setupFunc()

			s.T().Cleanup(func() {
				for _, cfgID := range configIDs {
					capturedID := cfgID
					s.service.DB.Unscoped().Delete(&app.AppConfig{}, "id = ?", capturedID)
				}
			})

			path := fmt.Sprintf("/v1/apps/%s/configs", s.testApp.ID)
			rr := s.makeGetRequest(http.MethodGet, path)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var response []*models.AppAppConfig
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)
			require.Len(s.T(), response, tc.expectedCount)
		})
	}
}

func (s *AppConfigsTestSuite) TestGetAppConfigsDoesNotReturnConfigsFromOtherApps() {
	otherApp := &app.App{
		ID:          domains.NewAppID(),
		Name:        "other-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	err := s.service.DB.Create(otherApp).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", otherApp.ID)

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)

	cfg1 := &app.AppConfig{
		ID:                domains.NewAppCfgID(),
		OrgID:             s.testOrg.ID,
		AppID:             s.testApp.ID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "pending",
		Readme:            "test app config",
	}
	err = s.service.DB.WithContext(ctx).Create(cfg1).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.AppConfig{}, "id = ?", cfg1.ID)

	cfg2 := &app.AppConfig{
		ID:                domains.NewAppCfgID(),
		OrgID:             s.testOrg.ID,
		AppID:             otherApp.ID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "pending",
		Readme:            "other app config",
	}
	err = s.service.DB.WithContext(ctx).Create(cfg2).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.AppConfig{}, "id = ?", cfg2.ID)

	path := fmt.Sprintf("/v1/apps/%s/configs", s.testApp.ID)
	rr := s.makeGetRequest(http.MethodGet, path)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response []*models.AppAppConfig
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
	}
	require.NoError(s.T(), err)
	require.Len(s.T(), response, 1)
	require.Equal(s.T(), s.testApp.ID, response[0].AppID)
	require.Equal(s.T(), "test app config", response[0].Readme)
}
