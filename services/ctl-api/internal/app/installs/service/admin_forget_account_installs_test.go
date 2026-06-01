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
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/forgotten"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AdminForgetAccountInstallsTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type AdminForgetAccountInstallsTestSuite struct {
	tests.BaseDBTestSuite
	app         *fxtest.App
	service     AdminForgetAccountInstallsTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	testApp     *app.App
	testInstall *app.Install
}

func TestAdminForgetAccountInstallsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminForgetAccountInstallsTestSuite))
}

func (s *AdminForgetAccountInstallsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *AdminForgetAccountInstallsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Admin routes do NOT use TestOrg/TestAcc context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.InstallsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminForgetAccountInstallsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminForgetAccountInstallsTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
}

func (s *AdminForgetAccountInstallsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(s.T(), err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminForgetAccountInstallsTestSuite) TestForgetAccountInstalls() {
	testCases := []struct {
		name             string
		setupFunc        func() AdminForgetAccountInstallsRequest
		expectedCode     int
		expectedSignal   bool
		validateFunc     func(AdminForgetAccountInstallsRequest)
		expectedNotFound bool
	}{
		{
			name: "successfully forget installs for account with AWS account",
			setupFunc: func() AdminForgetAccountInstallsRequest {
				ctx := context.Background()

				// Must set account context before creating entities with created_by_id
				ctx, _ = s.service.Seeder.EnsureAccount(ctx, s.T())
				ctx, org := s.service.Seeder.EnsureOrg(ctx, s.T())
				testApp := s.service.Seeder.CreateApp(ctx, s.T())
				s.service.Seeder.CreateAppConfig(ctx, s.T(), testApp.ID)
				install := s.service.Seeder.CreateInstall(ctx, s.T(), testApp)

				// Update the install's AWS account with specific IAMRoleARN
				testAccountID := "123456789012"
				err := s.service.DB.WithContext(ctx).
					Model(&app.AWSAccount{}).
					Where("install_id = ?", install.ID).
					Update("iam_role_arn", "arn:aws:iam::"+testAccountID+":role/test-role").Error
				require.NoError(s.T(), err)

				// Toggle sandbox mode off so the handler's filter doesn't skip it
				err = s.service.DB.WithContext(ctx).Model(org).Update("sandbox_mode", false).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(install)
					s.service.DB.Unscoped().Delete(testApp)
					s.service.DB.Unscoped().Delete(org)
				})

				return AdminForgetAccountInstallsRequest{
					AccountID: testAccountID,
				}
			},
			expectedCode:   http.StatusOK,
			expectedSignal: true,
			validateFunc: func(req AdminForgetAccountInstallsRequest) {
				// Verify signals were sent
				sigs := tests.GetQueueSignals(s.T(), s.service.DB)
				assert.GreaterOrEqual(s.T(), len(sigs), 1, "expected at least one signal")

				for _, qs := range sigs {
					assert.Equal(s.T(), forgotten.SignalType, qs.Type)
				}
			},
		},
		{
			name: "account with no installs returns success",
			setupFunc: func() AdminForgetAccountInstallsRequest {
				return AdminForgetAccountInstallsRequest{
					AccountID: "999999999999",
				}
			},
			expectedCode:   http.StatusOK,
			expectedSignal: false,
		},
		{
			name: "missing account ID returns validation error",
			setupFunc: func() AdminForgetAccountInstallsRequest {
				return AdminForgetAccountInstallsRequest{
					AccountID: "",
				}
			},
			expectedCode:     http.StatusBadRequest,
			expectedSignal:   false,
			expectedNotFound: true,
		},
		{
			name: "invalid request body returns error",
			setupFunc: func() AdminForgetAccountInstallsRequest {
				return AdminForgetAccountInstallsRequest{
					AccountID: "invalid",
				}
			},
			expectedCode:   http.StatusOK,
			expectedSignal: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/installs/admin-forget-account-installs", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusOK {
				tc.validateFunc(req)
			}

			// Verify signal presence matches expectation
			capturedSignals := tests.GetQueueSignals(s.T(), s.service.DB)
			if tc.expectedSignal {
				assert.GreaterOrEqual(s.T(), len(capturedSignals), 1, "expected at least one signal to be sent")
			} else {
				assert.Len(s.T(), capturedSignals, 0, "expected no signals to be sent")
			}
		})
	}
}
