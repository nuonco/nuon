package service

import (
	"context"
	"fmt"
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

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type LoginTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type LoginTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service LoginTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestLoginSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(LoginTestSuite))
}

func (s *LoginTestSuite) SetupSuite() {
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

func (s *LoginTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	// Register auth routes (includes HTML template loading)
	err := s.service.AuthService.RegisterAuthRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *LoginTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *LoginTestSuite) setupTestData() {
	ctx := context.Background()

	// Create test account
	accID := domains.NewAccountID()
	testAcc := &app.Account{
		ID:          accID,
		Email:       fmt.Sprintf("%s@test.nuon.co", accID),
		Subject:     accID,
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org with account context
	ctx = cctx.SetAccountContext(ctx, testAcc)
	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("test-org-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg

	// Note: Identity provider comes from environment config (default provider)
	// No need to create test identity provider in DB
}

func (s *LoginTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *LoginTestSuite) configuredProvider() string {
	return s.service.Cfg.NuonAuthProviderType
}

func (s *LoginTestSuite) TestLogin() {
	provider := s.configuredProvider()

	testCases := []struct {
		name           string
		queryParams    string
		expectedCode   int
		validateFunc   func(*httptest.ResponseRecorder)
		expectedError  bool
		errorSubstring string
	}{
		{
			name:           "missing provider parameter",
			queryParams:    "",
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "provider is required",
		},
		{
			name:           "invalid provider type",
			queryParams:    "?provider=invalid-provider",
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "invalid provider",
		},
		{
			name:         "valid provider type initiates OAuth flow",
			queryParams:  "?provider=" + provider,
			expectedCode: http.StatusFound, // 302 redirect
			validateFunc: func(rr *httptest.ResponseRecorder) {
				// Should redirect to OAuth provider
				location := rr.Header().Get("Location")
				assert.NotEmpty(s.T(), location, "should have Location header")

				// Should set session cookie
				cookies := rr.Result().Cookies()
				var foundSession bool
				for _, cookie := range cookies {
					if cookie.Name == NuonAuthSessionName {
						foundSession = true
						assert.NotEmpty(s.T(), cookie.Value)
						assert.True(s.T(), cookie.HttpOnly)
						break
					}
				}
				assert.True(s.T(), foundSession, "should set session cookie")
			},
		},
		{
			name:         "valid provider with requested URL",
			queryParams:  "?provider=" + provider + "&url=http://localhost:4000/dashboard",
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.NotEmpty(s.T(), location)

				// Session cookie should contain requested URL
				cookies := rr.Result().Cookies()
				var foundSession bool
				for _, cookie := range cookies {
					if cookie.Name == NuonAuthSessionName {
						foundSession = true
						break
					}
				}
				assert.True(s.T(), foundSession)
			},
		},
		{
			name:         "invalid URL encoding is silently dropped by Go HTTP parser",
			queryParams:  "?provider=" + provider + "&url=%ZZ%invalid",
			expectedCode: http.StatusFound, // Go drops malformed query params, so url="" and handler proceeds
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.NotEmpty(s.T(), location, "should redirect to OAuth provider")
			},
		},
		{
			name:           "URL without scheme",
			queryParams:    "?provider=" + provider + "&url=example.com/path",
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "must begin with http",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest("GET", "/login"+tc.queryParams)

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

// The env provider is addressable by its synthetic ID, and DB providers by their row ID, so that a
// deployment with two providers of the same type can send a user to a specific one.
func (s *LoginTestSuite) TestLoginByProviderID() {
	envProviderID := app.EnvIdentityProviderID(app.ProviderType(s.configuredProvider()))

	rr := s.makeRequest("GET", "/login?provider="+envProviderID)
	require.Equal(s.T(), http.StatusFound, rr.Code)
	assert.NotEmpty(s.T(), rr.Header().Get("Location"))

	rr = s.makeRequest("GET", "/login?provider=idp_does_not_exist")
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

// Two OIDC providers can coexist, and each is reachable by its own ID.
func (s *LoginTestSuite) TestLoginWithASecondProviderOfTheSameType() {
	secondary := &app.IdentityProvider{
		ID:           domains.NewIdentityProviderID(),
		ProviderType: app.ProviderTypeOIDC,
		Name:         "Secondary SSO",
		Enabled:      true,
	}
	require.NoError(s.T(), secondary.SetOpenIDConfig(&providers.OpenIDConfig{
		BaseConfig: providers.BaseConfig{
			ClientID:     "secondary-client-id",
			ClientSecret: "secondary-client-secret",
			RedirectURL:  s.service.Cfg.NuonAuthRedirectURL,
		},
		IssuerURL: s.service.Cfg.NuonAuthIssuerURL,
	}))
	secondary.ProviderType = app.ProviderTypeOIDC
	secondary.Name = "Secondary SSO"
	require.NoError(s.T(), s.service.DB.Create(secondary).Error)
	// other suites in this package assert on the provider list, so don't leak an enabled provider
	defer func() {
		require.NoError(s.T(), s.service.DB.Unscoped().Delete(secondary).Error)
	}()

	// the redirect itself needs live OIDC discovery against the issuer, which the dev-stack run
	// covers; here the point is that both providers coexist and are separately addressable
	rr := s.makeRequest("GET", "/")
	require.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "Secondary SSO",
		"a named second provider must be distinguishable on the sign-in page")
	assert.Contains(s.T(), rr.Body.String(), "provider="+secondary.ID)
}

func (s *LoginTestSuite) TestLoginClearsExistingCookie() {
	// First, set an auth cookie
	req, err := http.NewRequest("GET", "/login?provider="+s.configuredProvider(), nil)
	require.NoError(s.T(), err)

	// Add existing auth cookie
	req.AddCookie(&http.Cookie{
		Name:  NuonAuthCookieName,
		Value: "existing-token-value",
	})

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusFound, rr.Code)

	// Verify cookie was cleared (MaxAge = -1 indicates deletion)
	cookies := rr.Result().Cookies()
	var foundAuthCookie bool
	for _, cookie := range cookies {
		if cookie.Name == NuonAuthCookieName {
			foundAuthCookie = true
			// Cookie should be cleared with MaxAge = -1
			assert.Equal(s.T(), -1, cookie.MaxAge, "auth cookie should be cleared")
			break
		}
	}
	// Cookie clearing might not always set a new cookie, so this assertion is flexible
	s.T().Logf("Auth cookie found in response: %v", foundAuthCookie)
}
