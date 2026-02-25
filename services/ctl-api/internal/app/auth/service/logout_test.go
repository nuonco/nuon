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

type LogoutTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type LogoutTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service LogoutTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestLogoutSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(LogoutTestSuite))
}

func (s *LogoutTestSuite) SetupSuite() {
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

func (s *LogoutTestSuite) SetupTest() {
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

func (s *LogoutTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *LogoutTestSuite) setupTestData() {
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

func (s *LogoutTestSuite) makeRequestWithCookie(method, path string, authToken string) *httptest.ResponseRecorder {
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

func (s *LogoutTestSuite) createTestToken() string {
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

func (s *LogoutTestSuite) TestLogout() {
	testCases := []struct {
		name         string
		queryParams  string
		withToken    bool
		expectedCode int
		validateFunc func(*httptest.ResponseRecorder)
	}{
		{
			name:         "logout without token shows HTML page",
			queryParams:  "",
			withToken:    false,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "logged out", "should show logout success message")

				// Verify cookies are cleared
				cookies := rr.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == NuonAuthCookieName || cookie.Name == NuonAuthSessionName {
						assert.Equal(s.T(), -1, cookie.MaxAge, "cookie should be cleared")
					}
				}
			},
		},
		{
			name:         "logout with valid token deletes token from DB",
			queryParams:  "",
			withToken:    true,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "logged out")

				// Verify cookies are cleared
				cookies := rr.Result().Cookies()
				var foundAuthCookie, foundSessionCookie bool
				for _, cookie := range cookies {
					if cookie.Name == NuonAuthCookieName {
						foundAuthCookie = true
						assert.Equal(s.T(), -1, cookie.MaxAge)
					}
					if cookie.Name == NuonAuthSessionName {
						foundSessionCookie = true
						assert.Equal(s.T(), -1, cookie.MaxAge)
					}
				}
				s.T().Logf("Cleared cookies: auth=%v, session=%v", foundAuthCookie, foundSessionCookie)
			},
		},
		{
			name:         "logout with valid redirect URL",
			queryParams:  "?url=http://localhost:4000/login",
			withToken:    false,
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.Equal(s.T(), "http://localhost:4000/login", location,
					"should redirect to provided URL")

				// Verify cookies are cleared
				cookies := rr.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == NuonAuthCookieName || cookie.Name == NuonAuthSessionName {
						assert.Equal(s.T(), -1, cookie.MaxAge)
					}
				}
			},
		},
		{
			name:         "logout with invalid redirect URL ignores it",
			queryParams:  "?url=javascript:alert(1)",
			withToken:    false,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				// Should show HTML page instead of redirecting to invalid URL
				body := rr.Body.String()
				assert.Contains(s.T(), body, "logged out")

				// Should NOT have Location header
				location := rr.Header().Get("Location")
				assert.Empty(s.T(), location, "should not redirect to invalid URL")
			},
		},
		{
			name:         "logout with malformed URL",
			queryParams:  "?url=not-a-valid-url",
			withToken:    false,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, "logged out")

				location := rr.Header().Get("Location")
				assert.Empty(s.T(), location)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var authToken string
			if tc.withToken {
				authToken = s.createTestToken()

				// Verify token exists before logout
				var token app.Token
				err := s.service.DB.Where("token = ?", authToken).First(&token).Error
				require.NoError(s.T(), err, "token should exist before logout")
			}

			rr := s.makeRequestWithCookie("GET", "/logout"+tc.queryParams, authToken)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(rr)
			}

			if tc.withToken && authToken != "" {
				// Verify token was soft-deleted from DB
				var token app.Token
				err := s.service.DB.Where("token = ?", authToken).First(&token).Error
				assert.Error(s.T(), err, "token should be deleted after logout")
			}
		})
	}
}

func (s *LogoutTestSuite) TestLogoutClearsSessionCookie() {
	// Create a session cookie
	sessionData := &SessionData{
		State:      "test-state",
		ProviderID: "auth0",
	}
	encoded, err := s.service.AuthService.encodeSession(sessionData)
	require.NoError(s.T(), err)

	req, err := http.NewRequest("GET", "/logout", nil)
	require.NoError(s.T(), err)

	req.AddCookie(&http.Cookie{
		Name:  NuonAuthSessionName,
		Value: encoded,
	})

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Verify session cookie was cleared
	cookies := rr.Result().Cookies()
	var foundSessionCookie bool
	for _, cookie := range cookies {
		if cookie.Name == NuonAuthSessionName {
			foundSessionCookie = true
			assert.Equal(s.T(), -1, cookie.MaxAge, "session cookie should be cleared")
			break
		}
	}
	s.T().Logf("Session cookie cleared: %v", foundSessionCookie)
}

func (s *LogoutTestSuite) TestLogoutPreservesHTTPSRedirects() {
	testCases := []struct {
		name        string
		redirectURL string
		expectRedir bool
	}{
		{
			name:        "http localhost URL is preserved",
			redirectURL: "http://localhost:4000/login",
			expectRedir: true,
		},
		{
			name:        "http 127.0.0.1 URL is preserved",
			redirectURL: "http://127.0.0.1:4000/login",
			expectRedir: true,
		},
		{
			name:        "URL without scheme is rejected",
			redirectURL: "localhost:4000/login",
			expectRedir: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequestWithCookie("GET", "/logout?url="+tc.redirectURL, "")

			if tc.expectRedir {
				require.Equal(s.T(), http.StatusFound, rr.Code)
				location := rr.Header().Get("Location")
				assert.Equal(s.T(), tc.redirectURL, location)
			} else {
				require.Equal(s.T(), http.StatusOK, rr.Code)
				location := rr.Header().Get("Location")
				assert.Empty(s.T(), location)
			}
		})
	}
}
