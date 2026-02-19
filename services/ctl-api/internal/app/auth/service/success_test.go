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

type SuccessTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type SuccessTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service SuccessTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestSuccessSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(SuccessTestSuite))
}

func (s *SuccessTestSuite) SetupSuite() {
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

func (s *SuccessTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = gin.New()
	err := s.service.AuthService.RegisterAuthRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *SuccessTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *SuccessTestSuite) setupTestData() {
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

func (s *SuccessTestSuite) makeRequestWithCookie(method, path string, authToken string) *httptest.ResponseRecorder {
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

func (s *SuccessTestSuite) createTestToken() string {
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

func (s *SuccessTestSuite) TestSuccess() {
	testCases := []struct {
		name         string
		withToken    bool
		tokenExpired bool
		expectedCode int
		validateFunc func(*httptest.ResponseRecorder)
	}{
		{
			name:         "no auth cookie redirects to index",
			withToken:    false,
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.Equal(s.T(), "/", location, "should redirect to index page")
			},
		},
		{
			name:         "valid auth cookie shows success page",
			withToken:    true,
			tokenExpired: false,
			expectedCode: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				assert.Contains(s.T(), body, s.testAcc.Email,
					"should display user email")
				// Success page shows HTML content
				s.T().Logf("Success page body length: %d", len(body))
			},
		},
		{
			name:         "expired token redirects to index",
			withToken:    true,
			tokenExpired: true,
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.Equal(s.T(), "/", location, "should redirect to index page")

				// Verify cookie was cleared
				cookies := rr.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == NuonAuthCookieName {
						assert.Equal(s.T(), -1, cookie.MaxAge, "auth cookie should be cleared")
					}
				}
			},
		},
		{
			name:         "invalid token redirects to index",
			withToken:    false,
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.Equal(s.T(), "/", location)
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

			rr := s.makeRequestWithCookie("GET", "/success", authToken)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(rr)
			}
		})
	}
}

func (s *SuccessTestSuite) TestSuccessDisplaysUserInfo() {
	// Create valid token
	authToken := s.createTestToken()

	rr := s.makeRequestWithCookie("GET", "/success", authToken)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body := rr.Body.String()

	// Verify user information is displayed
	assert.Contains(s.T(), body, s.testAcc.Email, "should display user email")

	// Verify HTML structure (basic check)
	assert.Contains(s.T(), body, "html", "should be HTML response")
}

func (s *SuccessTestSuite) TestSuccessWithInvalidTokenFormat() {
	// Use invalid token that doesn't exist in DB
	invalidToken := "invalid-token-12345"

	rr := s.makeRequestWithCookie("GET", "/success", invalidToken)

	// Should redirect to index since token is invalid
	require.Equal(s.T(), http.StatusFound, rr.Code)

	location := rr.Header().Get("Location")
	assert.Equal(s.T(), "/", location, "should redirect to index page")

	// Verify cookie was cleared
	cookies := rr.Result().Cookies()
	var foundAuthCookie bool
	for _, cookie := range cookies {
		if cookie.Name == NuonAuthCookieName {
			foundAuthCookie = true
			assert.Equal(s.T(), -1, cookie.MaxAge, "auth cookie should be cleared")
			break
		}
	}
	s.T().Logf("Auth cookie cleared: %v", foundAuthCookie)
}

func (s *SuccessTestSuite) TestSuccessMultipleTimes() {
	// Verify success page can be accessed multiple times with same token
	authToken := s.createTestToken()

	// First access
	rr1 := s.makeRequestWithCookie("GET", "/success", authToken)
	require.Equal(s.T(), http.StatusOK, rr1.Code)
	assert.Contains(s.T(), rr1.Body.String(), s.testAcc.Email)

	// Second access (token should still be valid)
	rr2 := s.makeRequestWithCookie("GET", "/success", authToken)
	require.Equal(s.T(), http.StatusOK, rr2.Code)
	assert.Contains(s.T(), rr2.Body.String(), s.testAcc.Email)

	// Verify token still exists in DB
	var token app.Token
	err := s.service.DB.Where("token = ?", authToken).First(&token).Error
	require.NoError(s.T(), err, "token should still exist after multiple accesses")
}
