package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

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
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// ===========================
// DeviceCodePage Handler Tests
// ===========================

type DeviceCodePageTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type DeviceCodePageTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeviceCodePageTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestDeviceCodePageSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeviceCodePageTestSuite))
}

func (s *DeviceCodePageTestSuite) SetupSuite() {
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

func (s *DeviceCodePageTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AuthService.RegisterAuthRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeviceCodePageTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeviceCodePageTestSuite) setupTestData() {
	ctx := context.Background()

	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

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
}

func (s *DeviceCodePageTestSuite) makeRequestWithCookie(method, path string, authToken string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	if authToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  NuonAuthCookieName,
			Value: authToken,
		})
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeviceCodePageTestSuite) createTestToken() string {
	now := time.Now()
	tokenValue := domains.NewUserTokenID()

	token := app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   s.testAcc.ID,
		Token:       tokenValue,
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   now.Add(24 * time.Hour),
		IssuedAt:    now,
		Issuer:      "test",
	}

	err := s.service.DB.Create(&token).Error
	require.NoError(s.T(), err)

	return tokenValue
}

func (s *DeviceCodePageTestSuite) TestDeviceCodePage() {
	testCases := []struct {
		name           string
		code           string
		withToken      bool
		expectedCode   int
		expectedError  bool
		errorSubstring string
		validateFunc   func(*httptest.ResponseRecorder)
	}{
		{
			name:           "missing code parameter",
			code:           "",
			withToken:      false,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "device code is required",
		},
		{
			name:           "invalid code format - too short",
			code:           "ABC",
			withToken:      false,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "invalid device code format",
		},
		{
			name:           "invalid code format - lowercase",
			code:           "abcd-efgh",
			withToken:      false,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "invalid device code format",
		},
		{
			name:           "invalid code format - missing hyphen",
			code:           "ABCDEFGH",
			withToken:      false,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "invalid device code format",
		},
		{
			name:           "invalid code format - special characters",
			code:           "ABC$-EFGH",
			withToken:      false,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "invalid device code format",
		},
		{
			name:         "valid code without auth redirects to login",
			code:         "ABCD-1234",
			withToken:    false,
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.Contains(s.T(), location, "/login", "should redirect to login")
				assert.Contains(s.T(), location, "provider=", "should include provider param")
				assert.Contains(s.T(), location, "url=", "should include return URL")

				// Verify return URL contains the device code
				parsedURL, err := url.Parse(location)
				require.NoError(s.T(), err)
				returnURL := parsedURL.Query().Get("url")
				assert.Contains(s.T(), returnURL, "ABCD-1234", "return URL should contain device code")
			},
		},
		{
			name:         "valid code with auth shows approval page",
			code:         "WXYZ-5678",
			withToken:    true,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "WXYZ-5678", "should show device code")
				assert.Contains(s.T(), body, s.testAcc.Email, "should show user email")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var authToken string
			if tc.withToken {
				authToken = s.createTestToken()
			}

			path := "/device/code"
			if tc.code != "" {
				path += "?code=" + tc.code
			}

			rr := s.makeRequestWithCookie("GET", path, authToken)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedError {
				body := rr.Body.String()
				assert.Contains(s.T(), body, tc.errorSubstring)
			}

			if tc.validateFunc != nil {
				tc.validateFunc(rr)
			}
		})
	}
}

func (s *DeviceCodePageTestSuite) TestDeviceCodePageWithExpiredToken() {
	// Create expired token
	now := time.Now()
	tokenValue := domains.NewUserTokenID()
	token := app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   s.testAcc.ID,
		Token:       tokenValue,
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   now.Add(-1 * time.Hour), // Expired
		IssuedAt:    now.Add(-2 * time.Hour),
		Issuer:      "test",
	}
	err := s.service.DB.Create(&token).Error
	require.NoError(s.T(), err)

	rr := s.makeRequestWithCookie("GET", "/device/code?code=ABCD-1234", tokenValue)

	// Should redirect to login (expired token treated as no auth)
	require.Equal(s.T(), http.StatusFound, rr.Code)
	location := rr.Header().Get("Location")
	assert.Contains(s.T(), location, "/login")
}

// ===========================
// DeviceCodeApprove Handler Tests
// ===========================

type DeviceCodeApproveTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type DeviceCodeApproveTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeviceCodeApproveTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestDeviceCodeApproveSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeviceCodeApproveTestSuite))
}

func (s *DeviceCodeApproveTestSuite) SetupSuite() {
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

func (s *DeviceCodeApproveTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AuthService.RegisterAuthRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeviceCodeApproveTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeviceCodeApproveTestSuite) setupTestData() {
	ctx := context.Background()

	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

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
}

func (s *DeviceCodeApproveTestSuite) makePostRequestWithCookie(path, body string, authToken string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", path, strings.NewReader(body))
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if authToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  NuonAuthCookieName,
			Value: authToken,
		})
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeviceCodeApproveTestSuite) createTestToken() string {
	now := time.Now()
	tokenValue := domains.NewUserTokenID()

	token := app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   s.testAcc.ID,
		Token:       tokenValue,
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   now.Add(24 * time.Hour),
		IssuedAt:    now,
		Issuer:      "test",
	}

	err := s.service.DB.Create(&token).Error
	require.NoError(s.T(), err)

	return tokenValue
}

func (s *DeviceCodeApproveTestSuite) TestDeviceCodeApprove() {
	testCases := []struct {
		name           string
		code           string
		withToken      bool
		expectedCode   int
		expectedError  bool
		errorSubstring string
		validateFunc   func(*httptest.ResponseRecorder, string)
	}{
		{
			name:           "missing code parameter",
			code:           "",
			withToken:      true,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "device code is required",
		},
		{
			name:           "invalid code format",
			code:           "invalid",
			withToken:      true,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "invalid device code format",
		},
		{
			name:           "no authentication token",
			code:           "ABCD-1234",
			withToken:      false,
			expectedCode:   http.StatusUnauthorized,
			expectedError:  true,
			errorSubstring: "must be logged in",
		},
		{
			name:         "successful approval creates device code",
			code:         "WXYZ-5678",
			withToken:    true,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "CLI Authorized", "should show success message")
				assert.Contains(s.T(), body, s.testAcc.Email, "should show user email")

				// Verify device code was created in database
				var deviceCode app.DeviceCode
				err := s.service.DB.Where("code = ?", code).First(&deviceCode).Error
				require.NoError(s.T(), err, "device code should exist in database")
				assert.Equal(s.T(), s.testAcc.ID, deviceCode.AccountID)
				assert.False(s.T(), deviceCode.Consumed)
				assert.True(s.T(), deviceCode.ExpiresAt.After(time.Now()))
			},
		},
		{
			name:         "approving same code twice shows success",
			code:         "DUPL-1234",
			withToken:    true,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				// First approval
				authToken := s.createTestToken()
				rr1 := s.makePostRequestWithCookie("/device/code/approve", "code="+code, authToken)
				require.Equal(s.T(), http.StatusOK, rr1.Code)

				// Second approval (same code)
				rr2 := s.makePostRequestWithCookie("/device/code/approve", "code="+code, authToken)
				require.Equal(s.T(), http.StatusOK, rr2.Code)

				body := rr2.Body.String()
				assert.Contains(s.T(), body, "CLI Authorized")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var authToken string
			if tc.withToken {
				authToken = s.createTestToken()
			}

			body := ""
			if tc.code != "" {
				body = "code=" + tc.code
			}

			rr := s.makePostRequestWithCookie("/device/code/approve", body, authToken)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedError {
				responseBody := rr.Body.String()
				assert.Contains(s.T(), responseBody, tc.errorSubstring)
			}

			if tc.validateFunc != nil {
				tc.validateFunc(rr, tc.code)
			}
		})
	}
}

func (s *DeviceCodeApproveTestSuite) TestDeviceCodeApproveWithExpiredToken() {
	// Create expired token
	now := time.Now()
	tokenValue := domains.NewUserTokenID()
	token := app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   s.testAcc.ID,
		Token:       tokenValue,
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   now.Add(-1 * time.Hour),
		IssuedAt:    now.Add(-2 * time.Hour),
		Issuer:      "test",
	}
	err := s.service.DB.Create(&token).Error
	require.NoError(s.T(), err)

	rr := s.makePostRequestWithCookie("/device/code/approve", "code=ABCD-1234", tokenValue)

	require.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	body := rr.Body.String()
	assert.Contains(s.T(), body, "must be logged in")
}

func (s *DeviceCodeApproveTestSuite) TestDeviceCodeApproveConsumedCode() {
	authToken := s.createTestToken()
	code := "CONS-UMED"

	// Create a device code that's already consumed
	deviceCode := &app.DeviceCode{
		Code:      code,
		AccountID: s.testAcc.ID,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Consumed:  true,
	}
	err := s.service.DB.Create(deviceCode).Error
	require.NoError(s.T(), err)

	rr := s.makePostRequestWithCookie("/device/code/approve", "code="+code, authToken)

	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	body := rr.Body.String()
	assert.Contains(s.T(), body, "already used")
}

func (s *DeviceCodeApproveTestSuite) TestDeviceCodeApproveExpiredCode() {
	authToken := s.createTestToken()
	code := "EXPR-IRED"

	// Create a device code that's already expired
	deviceCode := &app.DeviceCode{
		Code:      code,
		AccountID: s.testAcc.ID,
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired
		Consumed:  false,
	}
	err := s.service.DB.Create(deviceCode).Error
	require.NoError(s.T(), err)

	rr := s.makePostRequestWithCookie("/device/code/approve", "code="+code, authToken)

	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	body := rr.Body.String()
	assert.Contains(s.T(), body, "expired")
}

// ===========================
// DeviceCodeToken Handler Tests
// ===========================

type DeviceCodeTokenTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type DeviceCodeTokenTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeviceCodeTokenTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestDeviceCodeTokenSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeviceCodeTokenTestSuite))
}

func (s *DeviceCodeTokenTestSuite) SetupSuite() {
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

func (s *DeviceCodeTokenTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AuthService.RegisterAuthRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeviceCodeTokenTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeviceCodeTokenTestSuite) setupTestData() {
	ctx := context.Background()

	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

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
}

func (s *DeviceCodeTokenTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeviceCodeTokenTestSuite) TestDeviceCodeToken() {
	testCases := []struct {
		name         string
		code         string
		setupFunc    func() string // Returns device code
		expectedCode int
		validateFunc func(*httptest.ResponseRecorder, string)
	}{
		{
			name:         "missing code parameter",
			code:         "",
			expectedCode: http.StatusBadRequest,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				var resp map[string]interface{}
				err := s.unmarshalJSON(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "missing_code", resp["error"])
				assert.Contains(s.T(), resp["error_description"], "device code is required")
			},
		},
		{
			name:         "invalid code format",
			code:         "invalid",
			expectedCode: http.StatusBadRequest,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				var resp map[string]interface{}
				err := s.unmarshalJSON(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "invalid_code", resp["error"])
			},
		},
		{
			name: "authorization pending - code not approved yet",
			code: "PEND-ING1",
			setupFunc: func() string {
				return "PEND-ING1"
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				var resp map[string]interface{}
				err := s.unmarshalJSON(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "authorization_pending", resp["error"])
				assert.Contains(s.T(), resp["error_description"], "waiting for user approval")
			},
		},
		{
			name: "expired device code",
			code: "EXPR-IRED",
			setupFunc: func() string {
				code := "EXPR-IRED"
				deviceCode := &app.DeviceCode{
					Code:      code,
					AccountID: s.testAcc.ID,
					ExpiresAt: time.Now().Add(-1 * time.Minute),
					Consumed:  false,
				}
				err := s.service.DB.Create(deviceCode).Error
				require.NoError(s.T(), err)
				return code
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				var resp map[string]interface{}
				err := s.unmarshalJSON(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "expired_token", resp["error"])
			},
		},
		{
			name: "already consumed device code",
			code: "CONS-UMED",
			setupFunc: func() string {
				code := "CONS-UMED"
				deviceCode := &app.DeviceCode{
					Code:      code,
					AccountID: s.testAcc.ID,
					ExpiresAt: time.Now().Add(5 * time.Minute),
					Consumed:  true,
				}
				err := s.service.DB.Create(deviceCode).Error
				require.NoError(s.T(), err)
				return code
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				var resp map[string]interface{}
				err := s.unmarshalJSON(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "access_denied", resp["error"])
				assert.Contains(s.T(), resp["error_description"], "already been used")
			},
		},
		{
			name: "successful token issuance",
			code: "SUCC-ESS1",
			setupFunc: func() string {
				code := "SUCC-ESS1"
				deviceCode := &app.DeviceCode{
					Code:      code,
					AccountID: s.testAcc.ID,
					ExpiresAt: time.Now().Add(5 * time.Minute),
					Consumed:  false,
				}
				err := s.service.DB.Create(deviceCode).Error
				require.NoError(s.T(), err)
				return code
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder, code string) {
				var resp map[string]interface{}
				err := s.unmarshalJSON(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)

				// Verify response structure
				assert.NotEmpty(s.T(), resp["access_token"], "should have access token")
				assert.Equal(s.T(), "Bearer", resp["token_type"])
				assert.Equal(s.T(), s.testAcc.Email, resp["email"])

				// Verify token was created in database
				tokenValue := resp["access_token"].(string)
				var token app.Token
				err = s.service.DB.Where("token = ?", tokenValue).First(&token).Error
				require.NoError(s.T(), err, "token should exist in database")
				assert.Equal(s.T(), s.testAcc.ID, token.AccountID)
				assert.Equal(s.T(), app.TokenTypeNuon, token.TokenType)

				// Verify device code was marked as consumed
				var deviceCode app.DeviceCode
				err = s.service.DB.Where("code = ?", code).First(&deviceCode).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), deviceCode.Consumed, "device code should be marked as consumed")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var code string
			if tc.setupFunc != nil {
				code = tc.setupFunc()
			} else {
				code = tc.code
			}

			path := "/device/token"
			if code != "" {
				path += "?code=" + code
			}

			rr := s.makeRequest("GET", path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(rr, code)
			}
		})
	}
}

func (s *DeviceCodeTokenTestSuite) TestDeviceCodeTokenPollingBehavior() {
	code := "POLL-TEST"

	// Poll before approval - should get authorization_pending
	rr1 := s.makeRequest("GET", "/device/token?code="+code)
	require.Equal(s.T(), http.StatusOK, rr1.Code)

	var resp1 map[string]interface{}
	err := s.unmarshalJSON(rr1.Body.Bytes(), &resp1)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "authorization_pending", resp1["error"])

	// Approve the device code
	deviceCode := &app.DeviceCode{
		Code:      code,
		AccountID: s.testAcc.ID,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Consumed:  false,
	}
	err = s.service.DB.Create(deviceCode).Error
	require.NoError(s.T(), err)

	// Poll after approval - should get token
	rr2 := s.makeRequest("GET", "/device/token?code="+code)
	require.Equal(s.T(), http.StatusOK, rr2.Code)

	var resp2 map[string]interface{}
	err = s.unmarshalJSON(rr2.Body.Bytes(), &resp2)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), resp2["access_token"])
	assert.Equal(s.T(), "Bearer", resp2["token_type"])

	// Poll again - should get access_denied (already consumed)
	rr3 := s.makeRequest("GET", "/device/token?code="+code)
	require.Equal(s.T(), http.StatusOK, rr3.Code)

	var resp3 map[string]interface{}
	err = s.unmarshalJSON(rr3.Body.Bytes(), &resp3)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "access_denied", resp3["error"])
}

func (s *DeviceCodeTokenTestSuite) unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
