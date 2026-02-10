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
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// AppConfigTypesTestSuite is the testify suite for app config-type endpoints.
type AppConfigTypesTestSuite struct {
	tests.BaseDBTestSuite

	app           *fxtest.App
	service       TestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testApp       *app.App
	testAppConfig *app.AppConfig
}

func TestAppConfigTypesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AppConfigTypesTestSuite))
}

func (s *AppConfigTypesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AppConfigTypesTestSuite) SetupTest() {
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

func (s *AppConfigTypesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AppConfigTypesTestSuite) setupTestData() {
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "user@example.com",
		Subject:     "subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "test-org",
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg

	testApp := &app.App{
		Name:        "test-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
	}
	err = s.service.DB.Create(testApp).Error
	require.NoError(s.T(), err)
	s.testApp = testApp

	ctx = cctx.SetOrgContext(ctx, testOrg)
	testAppConfig := &app.AppConfig{
		OrgID:             s.testOrg.ID,
		AppID:             s.testApp.ID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "test",
	}
	err = s.service.DB.WithContext(ctx).Create(testAppConfig).Error
	require.NoError(s.T(), err)
	s.testAppConfig = testAppConfig
}

func (s *AppConfigTypesTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// TestCreateAppSecretsConfig tests the CreateAppSecretsConfig endpoint.
func (s *AppConfigTypesTestSuite) TestCreateAppSecretsConfig() {
	testCases := []struct {
		name         string
		setupFunc    func() CreateAppSecretsConfigRequest
		expectedCode int
	}{
		{
			name: "successfully creates secrets config",
			setupFunc: func() CreateAppSecretsConfigRequest {
				return CreateAppSecretsConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					Secrets: []AppSecretConfig{
						{
							Name:        "my_secret",
							DisplayName: "My Secret",
							Description: "Test secret",
							Required:    true,
						},
						{
							Name:                      "k8s_secret",
							DisplayName:               "Kubernetes Secret",
							Description:               "Secret synced to k8s",
							KubernetesSync:            true,
							KubernetesSecretNamespace: "default",
							KubernetesSecretName:      "my-k8s-secret",
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "fails without app_config_id",
			setupFunc: func() CreateAppSecretsConfigRequest {
				return CreateAppSecretsConfigRequest{
					Secrets: []AppSecretConfig{
						{
							Name:        "my-secret",
							DisplayName: "My Secret",
							Description: "Test secret",
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "fails with invalid secret name (uppercase)",
			setupFunc: func() CreateAppSecretsConfigRequest {
				return CreateAppSecretsConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					Secrets: []AppSecretConfig{
						{
							Name:        "MY-SECRET",
							DisplayName: "My Secret",
							Description: "Test secret",
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "fails with invalid secret name (spaces)",
			setupFunc: func() CreateAppSecretsConfigRequest {
				return CreateAppSecretsConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					Secrets: []AppSecretConfig{
						{
							Name:        "my secret",
							DisplayName: "My Secret",
							Description: "Test secret",
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := tc.setupFunc()
			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+s.testApp.ID+"/secrets-configs", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var response app.AppSecretsConfig
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), response.ID)
				assert.Equal(s.T(), s.testApp.ID, response.AppID)
				assert.Equal(s.T(), s.testAppConfig.ID, response.AppConfigID)
				assert.NotEmpty(s.T(), response.Secrets)

				var dbConfig app.AppSecretsConfig
				err = s.service.DB.Preload("Secrets").First(&dbConfig, "id = ?", response.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testApp.ID, dbConfig.AppID)
			}
		})
	}
}
