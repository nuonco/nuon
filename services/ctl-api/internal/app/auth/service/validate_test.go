package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
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

type ValidateTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type ValidateTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service ValidateTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestValidateSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(ValidateTestSuite))
}

func (s *ValidateTestSuite) SetupSuite() {
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

func (s *ValidateTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = gin.New()
	err := s.service.AuthService.RegisterAuthRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *ValidateTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *ValidateTestSuite) setupTestData() {
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

func (s *ValidateTestSuite) makeRequestWithCookie(method, path string, authToken string) *httptest.ResponseRecorder {
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

func (s *ValidateTestSuite) createTestToken() string {
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

func (s *ValidateTestSuite) TestValidate() {
	testCases := []struct {
		name         string
		withToken    bool
		tokenExpired bool
		expectedCode int
		validateFunc func(*httptest.ResponseRecorder)
	}{
		{
			name:         "no token returns unauthorized",
			withToken:    false,
			expectedCode: http.StatusUnauthorized,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				// Verify error headers are set
				assert.Equal(s.T(), "false", rr.Header().Get(HeaderNuonAuthSuccess))

				// Should not have user info headers
				assert.Empty(s.T(), rr.Header().Get(HeaderNuonAuthUser))
				assert.Empty(s.T(), rr.Header().Get(HeaderNuonAuthEmail))
			},
		},
		{
			name:         "valid token returns OK with headers",
			withToken:    true,
			tokenExpired: false,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				// Verify success headers
				assert.Equal(s.T(), "true", rr.Header().Get(HeaderNuonAuthSuccess))

				// Verify user info headers
				assert.Equal(s.T(), s.testAcc.Email, rr.Header().Get(HeaderNuonAuthUser))
				assert.Equal(s.T(), s.testAcc.Email, rr.Header().Get(HeaderNuonAuthEmail))

				// Body should be empty (headers-only response)
				assert.Empty(s.T(), rr.Body.String())
			},
		},
		{
			name:         "expired token returns unauthorized",
			withToken:    true,
			tokenExpired: true,
			expectedCode: http.StatusUnauthorized,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "false", rr.Header().Get(HeaderNuonAuthSuccess))
				assert.Empty(s.T(), rr.Header().Get(HeaderNuonAuthUser))
			},
		},
		{
			name:         "invalid token returns unauthorized",
			withToken:    false,
			expectedCode: http.StatusUnauthorized,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				assert.Equal(s.T(), "false", rr.Header().Get(HeaderNuonAuthSuccess))
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var authToken string
			if tc.withToken {
				if tc.tokenExpired {
					// Create expired token
					now := time.Now()
					tokenValue := domains.NewUserTokenID()
					token := app.Token{
						CreatedByID: s.testAcc.ID,
						AccountID:   s.testAcc.ID,
						Token:       tokenValue,
						TokenType:   app.TokenTypeNuon,
						ExpiresAt:   now.Add(-1 * time.Hour), // Expired 1 hour ago
						IssuedAt:    now.Add(-2 * time.Hour),
						Issuer:      "test",
					}
					err := s.service.DB.Create(&token).Error
					require.NoError(s.T(), err)
					authToken = tokenValue
				} else {
					authToken = s.createTestToken()
				}
			}

			rr := s.makeRequestWithCookie("GET", "/validate", authToken)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Headers: %v", rr.Code, rr.Header())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(rr)
			}
		})
	}
}

func (s *ValidateTestSuite) TestValidateResponseHeaders() {
	// Create valid token
	authToken := s.createTestToken()

	rr := s.makeRequestWithCookie("GET", "/validate", authToken)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Verify all expected headers are present
	headers := rr.Header()

	// Success flag
	assert.Equal(s.T(), "true", headers.Get(HeaderNuonAuthSuccess))

	// User information
	assert.Equal(s.T(), s.testAcc.Email, headers.Get(HeaderNuonAuthUser),
		"should set user header")
	assert.Equal(s.T(), s.testAcc.Email, headers.Get(HeaderNuonAuthEmail),
		"should set email header")

	// Body should be empty (this is a validation endpoint for reverse proxies)
	assert.Empty(s.T(), rr.Body.String(), "body should be empty")
}

func (s *ValidateTestSuite) TestValidateWithInvalidTokenFormat() {
	// Use invalid token that doesn't exist in DB
	invalidToken := "invalid-token-12345"

	rr := s.makeRequestWithCookie("GET", "/validate", invalidToken)

	require.Equal(s.T(), http.StatusUnauthorized, rr.Code)

	// Verify error headers
	assert.Equal(s.T(), "false", rr.Header().Get(HeaderNuonAuthSuccess))
	assert.Empty(s.T(), rr.Header().Get(HeaderNuonAuthUser))
	assert.Empty(s.T(), rr.Header().Get(HeaderNuonAuthEmail))
}

func (s *ValidateTestSuite) TestValidateUsedByReverseProxy() {
	// This test simulates how a reverse proxy (nginx) would use the validate endpoint
	authToken := s.createTestToken()

	// Reverse proxy makes validation request
	req, err := http.NewRequest("GET", "/validate", nil)
	require.NoError(s.T(), err)

	// Pass auth cookie from original request
	req.AddCookie(&http.Cookie{
		Name:  NuonAuthCookieName,
		Value: authToken,
	})

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	// Proxy checks status code
	require.Equal(s.T(), http.StatusOK, rr.Code,
		"proxy should see 200 for valid auth")

	// Proxy can forward user info headers to upstream
	userHeader := rr.Header().Get(HeaderNuonAuthUser)
	emailHeader := rr.Header().Get(HeaderNuonAuthEmail)
	successHeader := rr.Header().Get(HeaderNuonAuthSuccess)

	assert.Equal(s.T(), s.testAcc.Email, userHeader)
	assert.Equal(s.T(), s.testAcc.Email, emailHeader)
	assert.Equal(s.T(), "true", successHeader)

	s.T().Logf("Proxy would forward headers: User=%s, Email=%s, Success=%s",
		userHeader, emailHeader, successHeader)
}

func (s *ValidateTestSuite) TestValidateMultipleTimes() {
	// Verify validate endpoint can be called multiple times with same token
	authToken := s.createTestToken()

	// First validation
	rr1 := s.makeRequestWithCookie("GET", "/validate", authToken)
	require.Equal(s.T(), http.StatusOK, rr1.Code)
	assert.Equal(s.T(), "true", rr1.Header().Get(HeaderNuonAuthSuccess))

	// Second validation (token should still be valid)
	rr2 := s.makeRequestWithCookie("GET", "/validate", authToken)
	require.Equal(s.T(), http.StatusOK, rr2.Code)
	assert.Equal(s.T(), "true", rr2.Header().Get(HeaderNuonAuthSuccess))

	// Verify token still exists in DB
	var token app.Token
	err := s.service.DB.Where("token = ?", authToken).First(&token).Error
	require.NoError(s.T(), err, "token should still exist after multiple validations")
}

func (s *ValidateTestSuite) TestValidateHeaderConstants() {
	// Verify header constant values match expected format
	assert.Equal(s.T(), "X-Nuon-Auth-User", HeaderNuonAuthUser)
	assert.Equal(s.T(), "X-Nuon-Auth-Email", HeaderNuonAuthEmail)
	assert.Equal(s.T(), "X-Nuon-Auth-Success", HeaderNuonAuthSuccess)

	// These headers should be used by nginx/reverse proxy for auth_request
	s.T().Logf("Validation headers: User=%s, Email=%s, Success=%s",
		HeaderNuonAuthUser, HeaderNuonAuthEmail, HeaderNuonAuthSuccess)
}

func (s *ValidateTestSuite) TestValidateWithDifferentAccounts() {
	// Create second account
	testAcc2 := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test2@example.com",
		Subject:     "test-subject-2",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc2).Error
	require.NoError(s.T(), err)

	// Create token for second account
	now := time.Now()
	tokenValue2 := domains.NewUserTokenID()
	token2 := app.Token{
		CreatedByID: testAcc2.ID,
		AccountID:   testAcc2.ID,
		Token:       tokenValue2,
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   now.Add(24 * time.Hour),
		IssuedAt:    now,
		Issuer:      "test",
	}
	err = s.service.DB.Create(&token2).Error
	require.NoError(s.T(), err)

	// Validate with first account's token
	authToken1 := s.createTestToken()
	rr1 := s.makeRequestWithCookie("GET", "/validate", authToken1)
	require.Equal(s.T(), http.StatusOK, rr1.Code)
	assert.Equal(s.T(), s.testAcc.Email, rr1.Header().Get(HeaderNuonAuthEmail))

	// Validate with second account's token
	rr2 := s.makeRequestWithCookie("GET", "/validate", tokenValue2)
	require.Equal(s.T(), http.StatusOK, rr2.Code)
	assert.Equal(s.T(), testAcc2.Email, rr2.Header().Get(HeaderNuonAuthEmail))

	// Verify tokens are isolated per account
	assert.NotEqual(s.T(), rr1.Header().Get(HeaderNuonAuthEmail),
		rr2.Header().Get(HeaderNuonAuthEmail),
		"different accounts should have different emails")
}
