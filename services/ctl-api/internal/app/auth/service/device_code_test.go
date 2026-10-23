package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
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
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

const (
	testRequesterIP        = "203.0.113.7"
	testRequesterUserAgent = "nuon-cli/test"
)

var userCodePattern = regexp.MustCompile(`^[BCDFGHJKLMNPQRSTVWXZ]{4}-[BCDFGHJKLMNPQRSTVWXZ]{4}$`)

type DeviceAuthorizationTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type DeviceAuthorizationTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeviceAuthorizationTestService
	router  *gin.Engine
	testAcc *app.Account
}

type deviceAuthorizeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type deviceTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Email            string `json:"email"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func TestDeviceAuthorizationSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeviceAuthorizationTestSuite))
}

func (s *DeviceAuthorizationTestSuite) SetupSuite() {
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

func (s *DeviceAuthorizationTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	accID := domains.NewAccountID()
	s.testAcc = &app.Account{
		ID:          accID,
		Email:       fmt.Sprintf("%s@test.nuon.co", accID),
		Subject:     accID,
		AccountType: app.AccountTypeAuth0,
	}
	require.NoError(s.T(), s.service.DB.Create(s.testAcc).Error)

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})
	require.NoError(s.T(), s.service.AuthService.RegisterAuthRoutes(s.router))
}

func (s *DeviceAuthorizationTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeviceAuthorizationTestSuite) do(method, path string, form url.Values, sessionToken string) *httptest.ResponseRecorder {
	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	req, err := http.NewRequest(method, path, body)
	require.NoError(s.T(), err)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("User-Agent", testRequesterUserAgent)
	req.RemoteAddr = testRequesterIP + ":51234"
	if sessionToken != "" {
		req.AddCookie(&http.Cookie{Name: NuonAuthCookieName, Value: sessionToken})
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeviceAuthorizationTestSuite) createSessionToken(tokenType app.TokenType, expiresAt time.Time) app.Token {
	token := app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   s.testAcc.ID,
		Token:       domains.NewUserTokenID(),
		TokenType:   tokenType,
		ExpiresAt:   expiresAt,
		IssuedAt:    time.Now(),
		Issuer:      "test",
	}
	require.NoError(s.T(), s.service.DB.Create(&token).Error)
	return token
}

func (s *DeviceAuthorizationTestSuite) createSession() app.Token {
	return s.createSessionToken(app.TokenTypeNuon, time.Now().Add(24*time.Hour))
}

func (s *DeviceAuthorizationTestSuite) createAuthorization(status app.DeviceAuthorizationStatus, expiresAt time.Time, sourceTokenID string) *app.DeviceAuthorization {
	deviceCode, err := generateDeviceCode()
	require.NoError(s.T(), err)
	userCode, err := generateUserCode()
	require.NoError(s.T(), err)

	authz := &app.DeviceAuthorization{
		DeviceCode:         deviceCode,
		UserCode:           userCode,
		RequesterIP:        testRequesterIP,
		RequesterUserAgent: testRequesterUserAgent,
		Status:             status,
		ExpiresAt:          expiresAt,
	}
	if status != app.DeviceAuthorizationStatusPending {
		authz.AccountID = s.testAcc.ID
		authz.SourceTokenID = sourceTokenID
	}
	require.NoError(s.T(), s.service.DB.Create(authz).Error)
	return authz
}

func (s *DeviceAuthorizationTestSuite) reload(id string) app.DeviceAuthorization {
	var authz app.DeviceAuthorization
	require.NoError(s.T(), s.service.DB.Where(&app.DeviceAuthorization{ID: id}).First(&authz).Error)
	return authz
}

func (s *DeviceAuthorizationTestSuite) authorize(form url.Values) deviceAuthorizeResponse {
	rr := s.do(http.MethodPost, "/device/authorize", form, "")
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp deviceAuthorizeResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	return resp
}

func (s *DeviceAuthorizationTestSuite) pollToken(deviceCode string) (int, deviceTokenResponse) {
	rr := s.do(http.MethodPost, "/device/token", url.Values{
		"grant_type":  {deviceCodeGrantType},
		"device_code": {deviceCode},
	}, "")

	var resp deviceTokenResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp), rr.Body.String())
	return rr.Code, resp
}

func (s *DeviceAuthorizationTestSuite) TestAuthorizeIssuesCodes() {
	resp := s.authorize(url.Values{})

	assert.Len(s.T(), resp.DeviceCode, 43, "device code should carry 256 bits")
	assert.Regexp(s.T(), userCodePattern, resp.UserCode)
	assert.Equal(s.T(), s.service.AuthService.buildDeviceURL(), resp.VerificationURI)
	assert.NotContains(s.T(), resp.VerificationURI, resp.UserCode)
	assert.Equal(s.T(), s.service.AuthService.buildDeviceURL()+"?user_code="+resp.UserCode, resp.VerificationURIComplete)
	assert.NotContains(s.T(), resp.VerificationURIComplete, resp.DeviceCode)
	assert.Equal(s.T(), int(deviceAuthorizationExpiry.Seconds()), resp.ExpiresIn)
	assert.Equal(s.T(), int(deviceAuthorizationInterval.Seconds()), resp.Interval)

	var authz app.DeviceAuthorization
	require.NoError(s.T(), s.service.DB.Where(&app.DeviceAuthorization{DeviceCode: resp.DeviceCode}).First(&authz).Error)
	assert.Equal(s.T(), resp.UserCode, authz.UserCode)
	assert.Equal(s.T(), app.DeviceAuthorizationStatusPending, authz.Status)
	assert.Equal(s.T(), testRequesterIP, authz.RequesterIP)
	assert.Equal(s.T(), testRequesterUserAgent, authz.RequesterUserAgent)
	assert.Empty(s.T(), authz.AccountID)
}

func (s *DeviceAuthorizationTestSuite) TestAuthorizeIgnoresClientChosenCodes() {
	resp := s.authorize(url.Values{
		"code":        {"ATTK-0001"},
		"user_code":   {"BCDF-GHJK"},
		"device_code": {"attacker-chosen"},
	})

	assert.NotEqual(s.T(), "BCDF-GHJK", resp.UserCode)
	assert.NotEqual(s.T(), "attacker-chosen", resp.DeviceCode)

	second := s.authorize(url.Values{})
	assert.NotEqual(s.T(), resp.DeviceCode, second.DeviceCode)
	assert.NotEqual(s.T(), resp.UserCode, second.UserCode)
}

func (s *DeviceAuthorizationTestSuite) TestDevicePage() {
	s.Run("without a session redirects to login and back to /device", func() {
		rr := s.do(http.MethodGet, "/device", nil, "")
		require.Equal(s.T(), http.StatusFound, rr.Code)
		s.assertDeviceLoginRedirect(rr.Header().Get("Location"), s.service.AuthService.buildDeviceURL())
	})

	s.Run("without a session keeps the pre-filled code through login", func() {
		rr := s.do(http.MethodGet, "/device?user_code=bcdf-ghjk", nil, "")
		require.Equal(s.T(), http.StatusFound, rr.Code)
		s.assertDeviceLoginRedirect(rr.Header().Get("Location"), s.service.AuthService.buildDeviceURL()+"?user_code=BCDF-GHJK")
	})

	s.Run("with an expired session redirects to login", func() {
		session := s.createSessionToken(app.TokenTypeNuon, time.Now().Add(-time.Hour))
		rr := s.do(http.MethodGet, "/device", nil, session.Token)
		require.Equal(s.T(), http.StatusFound, rr.Code)
		s.assertDeviceLoginRedirect(rr.Header().Get("Location"), s.service.AuthService.buildDeviceURL())
	})

	s.Run("with a non-session token redirects to login", func() {
		session := s.createSessionToken(app.TokenTypeOAuth, time.Now().Add(time.Hour))
		rr := s.do(http.MethodGet, "/device", nil, session.Token)
		require.Equal(s.T(), http.StatusFound, rr.Code)
	})

	s.Run("with a session shows an empty entry form", func() {
		session := s.createSession()

		rr := s.do(http.MethodGet, "/device", nil, session.Token)
		require.Equal(s.T(), http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(s.T(), body, `name="user_code"`)
		assert.Contains(s.T(), body, `value=""`)
		assert.Contains(s.T(), body, s.testAcc.Email)
	})

	s.Run("a code in the link only pre-fills the form", func() {
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")
		session := s.createSession()

		rr := s.do(http.MethodGet, "/device?user_code="+authz.UserCode, nil, session.Token)
		require.Equal(s.T(), http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(s.T(), body, `value="`+authz.UserCode+`"`)
		assert.Contains(s.T(), body, `action="/device"`)
		assert.NotContains(s.T(), body, "/device/approve")
		assert.Equal(s.T(), app.DeviceAuthorizationStatusPending, s.reload(authz.ID).Status)
	})

	s.Run("a malformed code in the link is dropped", func() {
		session := s.createSession()

		rr := s.do(http.MethodGet, `/device?user_code=%22%3E%3Cscript%3E`, nil, session.Token)
		require.Equal(s.T(), http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(s.T(), body, `value=""`)
		assert.NotContains(s.T(), body, `"><script>`)
		assert.NotContains(s.T(), body, `&#34;&gt;&lt;script&gt;`)
	})
}

func (s *DeviceAuthorizationTestSuite) TestLookup() {
	session := s.createSession()

	s.Run("without a session is rejected", func() {
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")
		rr := s.do(http.MethodPost, "/device", url.Values{"user_code": {authz.UserCode}}, "")
		require.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	})

	testCases := map[string]struct {
		userCode     func() string
		errSubstring string
	}{
		"empty code": {
			userCode:     func() string { return "  " },
			errSubstring: "enter the code",
		},
		"malformed code": {
			userCode:     func() string { return "ABC$-1234" },
			errSubstring: "invalid or has expired",
		},
		"unknown code": {
			userCode: func() string {
				code, err := generateUserCode()
				require.NoError(s.T(), err)
				return code
			},
			errSubstring: "invalid or has expired",
		},
		"expired code": {
			userCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(-time.Minute), "").UserCode
			},
			errSubstring: "invalid or has expired",
		},
		"already approved code": {
			userCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusApproved, time.Now().Add(time.Minute), session.ID).UserCode
			},
			errSubstring: "invalid or has expired",
		},
	}
	for name, tc := range testCases {
		s.Run(name, func() {
			rr := s.do(http.MethodPost, "/device", url.Values{"user_code": {tc.userCode()}}, session.Token)
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
			assert.Contains(s.T(), rr.Body.String(), tc.errSubstring)
			assert.NotContains(s.T(), rr.Body.String(), "/device/approve")
		})
	}

	s.Run("pending code shows who is asking", func() {
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")
		typed := strings.ToLower(strings.ReplaceAll(authz.UserCode, "-", " "))

		rr := s.do(http.MethodPost, "/device", url.Values{"user_code": {typed}}, session.Token)
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
		body := rr.Body.String()
		assert.Contains(s.T(), body, authz.UserCode)
		assert.Contains(s.T(), body, testRequesterIP)
		assert.Contains(s.T(), body, testRequesterUserAgent)
		assert.Contains(s.T(), body, authz.CreatedAt.UTC().Format("2006-01-02 15:04"))
		assert.Contains(s.T(), body, s.testAcc.Email)
		assert.Contains(s.T(), body, `action="/device/approve"`)
		assert.Contains(s.T(), body, "matches the code in your terminal")

		assert.Equal(s.T(), app.DeviceAuthorizationStatusPending, s.reload(authz.ID).Status, "lookup must not approve")
	})
}

func (s *DeviceAuthorizationTestSuite) TestApprove() {
	s.Run("approve binds the request to the signed-in account", func() {
		session := s.createSession()
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")

		rr := s.do(http.MethodPost, "/device/approve", url.Values{
			"user_code": {authz.UserCode},
			"action":    {"approve"},
		}, session.Token)
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
		assert.Contains(s.T(), rr.Body.String(), "CLI Authorized")

		updated := s.reload(authz.ID)
		assert.Equal(s.T(), app.DeviceAuthorizationStatusApproved, updated.Status)
		assert.Equal(s.T(), s.testAcc.ID, updated.AccountID)
		assert.Equal(s.T(), session.ID, updated.SourceTokenID)
	})

	s.Run("deny marks the request denied", func() {
		session := s.createSession()
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")

		rr := s.do(http.MethodPost, "/device/approve", url.Values{
			"user_code": {authz.UserCode},
			"action":    {"deny"},
		}, session.Token)
		require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
		assert.Contains(s.T(), rr.Body.String(), "CLI Access Denied")

		updated := s.reload(authz.ID)
		assert.Equal(s.T(), app.DeviceAuthorizationStatusDenied, updated.Status)
		assert.Empty(s.T(), updated.AccountID)
	})

	s.Run("a request can only be decided once", func() {
		session := s.createSession()
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")
		form := url.Values{"user_code": {authz.UserCode}, "action": {"approve"}}

		require.Equal(s.T(), http.StatusOK, s.do(http.MethodPost, "/device/approve", form, session.Token).Code)
		rr := s.do(http.MethodPost, "/device/approve", form, session.Token)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		assert.Contains(s.T(), rr.Body.String(), "invalid or has expired")
	})

	s.Run("expired request cannot be approved", func() {
		session := s.createSession()
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(-time.Minute), "")

		rr := s.do(http.MethodPost, "/device/approve", url.Values{
			"user_code": {authz.UserCode},
			"action":    {"approve"},
		}, session.Token)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		assert.Equal(s.T(), app.DeviceAuthorizationStatusPending, s.reload(authz.ID).Status)
	})

	s.Run("without a session is rejected", func() {
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")

		rr := s.do(http.MethodPost, "/device/approve", url.Values{
			"user_code": {authz.UserCode},
			"action":    {"approve"},
		}, "")
		require.Equal(s.T(), http.StatusUnauthorized, rr.Code)
		assert.Equal(s.T(), app.DeviceAuthorizationStatusPending, s.reload(authz.ID).Status)
	})

	s.Run("a limited oauth token cannot approve", func() {
		limited := s.createSessionToken(app.TokenTypeOAuth, time.Now().Add(time.Hour))
		authz := s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "")

		rr := s.do(http.MethodPost, "/device/approve", url.Values{
			"user_code": {authz.UserCode},
			"action":    {"approve"},
		}, limited.Token)
		require.Equal(s.T(), http.StatusUnauthorized, rr.Code)
		assert.Equal(s.T(), app.DeviceAuthorizationStatusPending, s.reload(authz.ID).Status)
	})
}

func (s *DeviceAuthorizationTestSuite) TestTokenErrors() {
	session := s.createSession()

	s.Run("wrong grant type", func() {
		rr := s.do(http.MethodPost, "/device/token", url.Values{"device_code": {"x"}}, "")
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		assert.Contains(s.T(), rr.Body.String(), "unsupported_grant_type")
	})

	testCases := map[string]struct {
		deviceCode func() string
		error      string
	}{
		"missing device code": {
			deviceCode: func() string { return "" },
			error:      "invalid_request",
		},
		"unknown device code": {
			deviceCode: func() string {
				code, err := generateDeviceCode()
				require.NoError(s.T(), err)
				return code
			},
			error: "invalid_grant",
		},
		"user code is not accepted as a device code": {
			deviceCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusApproved, time.Now().Add(time.Minute), session.ID).UserCode
			},
			error: "invalid_grant",
		},
		"pending": {
			deviceCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(time.Minute), "").DeviceCode
			},
			error: "authorization_pending",
		},
		"denied": {
			deviceCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusDenied, time.Now().Add(time.Minute), session.ID).DeviceCode
			},
			error: "access_denied",
		},
		"expired while pending": {
			deviceCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusPending, time.Now().Add(-time.Minute), "").DeviceCode
			},
			error: "expired_token",
		},
		"expired after approval": {
			deviceCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusApproved, time.Now().Add(-time.Minute), session.ID).DeviceCode
			},
			error: "expired_token",
		},
		"already consumed": {
			deviceCode: func() string {
				return s.createAuthorization(app.DeviceAuthorizationStatusConsumed, time.Now().Add(time.Minute), session.ID).DeviceCode
			},
			error: "invalid_grant",
		},
	}
	for name, tc := range testCases {
		s.Run(name, func() {
			status, resp := s.pollToken(tc.deviceCode())
			require.Equal(s.T(), http.StatusBadRequest, status)
			assert.Equal(s.T(), tc.error, resp.Error)
			assert.Empty(s.T(), resp.AccessToken)
		})
	}
}

func (s *DeviceAuthorizationTestSuite) TestFullFlow() {
	session := s.createSession()
	start := s.authorize(url.Values{})

	status, resp := s.pollToken(start.DeviceCode)
	require.Equal(s.T(), http.StatusBadRequest, status)
	assert.Equal(s.T(), "authorization_pending", resp.Error)

	rr := s.do(http.MethodPost, "/device", url.Values{"user_code": {start.UserCode}}, session.Token)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	rr = s.do(http.MethodPost, "/device/approve", url.Values{
		"user_code": {start.UserCode},
		"action":    {"approve"},
	}, session.Token)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	status, resp = s.pollToken(start.DeviceCode)
	require.Equal(s.T(), http.StatusOK, status)
	require.NotEmpty(s.T(), resp.AccessToken)
	assert.Equal(s.T(), "Bearer", resp.TokenType)
	assert.Equal(s.T(), s.testAcc.Email, resp.Email)
	assert.Equal(s.T(), s.service.Cfg.NuonAuthTokenTTL*60, resp.ExpiresIn)

	var token app.Token
	require.NoError(s.T(), s.service.DB.Where(&app.Token{Token: resp.AccessToken}).First(&token).Error)
	assert.Equal(s.T(), s.testAcc.ID, token.AccountID)
	assert.Equal(s.T(), app.TokenTypeNuon, token.TokenType)
	assert.Equal(s.T(), app.TokenSourceTypeToken, token.SourceType)
	assert.Equal(s.T(), session.ID, token.SourceID)

	status, resp = s.pollToken(start.DeviceCode)
	require.Equal(s.T(), http.StatusBadRequest, status)
	assert.Equal(s.T(), "invalid_grant", resp.Error)
}

func (s *DeviceAuthorizationTestSuite) TestLegacyEndpoints() {
	session := s.createSession()

	s.Run("old CLI polling gets an upgrade error", func() {
		rr := s.do(http.MethodGet, "/device/token?code=ATTK-0001", nil, "")
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var resp deviceTokenResponse
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(s.T(), "unsupported_client", resp.Error)
		assert.Contains(s.T(), resp.ErrorDescription, "upgrade")
		assert.Empty(s.T(), resp.AccessToken)
	})

	s.Run("old approval link cannot approve anything", func() {
		rr := s.do(http.MethodGet, "/device/code?code=ATTK-0001", nil, session.Token)
		require.Equal(s.T(), http.StatusGone, rr.Code)
		assert.Contains(s.T(), rr.Body.String(), "upgrade")
		assert.NotContains(s.T(), rr.Body.String(), "ATTK-0001")
	})
}

// assertDeviceLoginRedirect checks the unauthenticated redirect. Which shape is correct depends on
// how many providers are enabled, and this database is shared with other suites, so the
// expectation is derived rather than hardcoded.
func (s *DeviceAuthorizationTestSuite) assertDeviceLoginRedirect(location, returnURL string) {
	identityProviders, err := s.service.AuthService.getIdentityProviders(context.Background())
	require.NoError(s.T(), err)

	if len(identityProviders) == 1 {
		assert.Contains(s.T(), location, "/login", "a single provider should be linked directly")
		assert.Contains(s.T(), location, "provider="+identityProviders[0].ID)
	} else {
		assert.Contains(s.T(), location, "/?url=", "several providers should land on the picker")
		assert.NotContains(s.T(), location, "provider=")
	}

	parsedURL, err := url.Parse(location)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), returnURL, parsedURL.Query().Get("url"))
}
