package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/nuonauth"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type OAuthCLITestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type OAuthCLITestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service OAuthCLITestService
	router  *gin.Engine
	testAcc *app.Account
}

func TestOAuthCLISuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(OAuthCLITestSuite))
}

func (s *OAuthCLITestSuite) SetupSuite() {
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

func (s *OAuthCLITestSuite) SetupTest() {
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

func (s *OAuthCLITestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *OAuthCLITestSuite) do(method, path string, form url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	body := strings.NewReader("")
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, path, body)
	require.NoError(s.T(), err)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *OAuthCLITestSuite) session() app.Token {
	token := app.Token{
		CreatedByID: s.testAcc.ID,
		AccountID:   s.testAcc.ID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		IssuedAt:    time.Now(),
		Issuer:      "test",
	}
	require.NoError(s.T(), s.service.DB.Create(&token).Error)
	return token
}

func (s *OAuthCLITestSuite) TestCLIClientIsSeeded() {
	var client app.OAuthClient
	require.NoError(s.T(), s.service.DB.Where(&app.OAuthClient{ID: nuonauth.CLIClientID}).First(&client).Error)
	assert.Equal(s.T(), nuonauth.CLIClientName, client.ClientName)
	assert.True(s.T(), client.AllowsRedirectURI("http://127.0.0.1:54321/callback"))
}

func (s *OAuthCLITestSuite) TestCLIAuthorizeAllowsLoopbackPort() {
	q := url.Values{}
	q.Set("client_id", nuonauth.CLIClientID)
	q.Set("redirect_uri", "http://127.0.0.1:60000/callback")
	q.Set("response_type", "code")
	q.Set("code_challenge", s256("verifier-verifier-verifier-verifier"))
	q.Set("code_challenge_method", "S256")
	q.Set("state", "st")

	rr := s.do(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil)
	require.Equal(s.T(), http.StatusFound, rr.Code)
	assert.Contains(s.T(), rr.Header().Get("Location"), "/login")
}

func (s *OAuthCLITestSuite) TestMCPAuthorizeRejectsOtherLoopbackPort() {
	client := app.OAuthClient{
		ClientName:              "mcp-test",
		RedirectURIs:            pq.StringArray{"http://127.0.0.1:1111/cb"},
		TokenEndpointAuthMethod: "none",
	}
	require.NoError(s.T(), s.service.DB.Create(&client).Error)

	q := url.Values{}
	q.Set("client_id", client.ID)
	q.Set("redirect_uri", "http://127.0.0.1:2222/cb")
	q.Set("response_type", "code")
	q.Set("code_challenge", s256("verifier-verifier-verifier-verifier"))
	q.Set("code_challenge_method", "S256")

	rr := s.do(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "redirect_uri is not registered")
}

func (s *OAuthCLITestSuite) TestCLIConsentIssuesNuonToken() {
	session := s.session()
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	redirectURI := "http://127.0.0.1:54321/callback"
	requestID := "rid-" + s.testAcc.ID
	pending := app.OAuthAuthorizationCode{
		RequestID:           requestID,
		ClientID:            nuonauth.CLIClientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       s256(verifier),
		CodeChallengeMethod: "S256",
		ClientState:         "cli-state",
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	}
	require.NoError(s.T(), s.service.DB.Create(&pending).Error)

	cookies := []*http.Cookie{
		{Name: NuonAuthCookieName, Value: session.Token},
		{Name: oauthRequestCookieName, Value: s.service.AuthService.oauthRequestCookieValue(requestID)},
	}

	finish := s.do(http.MethodGet, "/oauth/finish?rid="+url.QueryEscape(requestID), nil, cookies...)
	require.Equal(s.T(), http.StatusOK, finish.Code)
	assert.Contains(s.T(), finish.Body.String(), "Sign in to the Nuon CLI on this computer.")
	assert.NotContains(s.T(), finish.Body.String(), "Read only")

	consent := s.do(http.MethodPost, "/oauth/consent", url.Values{"rid": {requestID}}, cookies...)
	require.Equal(s.T(), http.StatusFound, consent.Code)
	loc, err := url.Parse(consent.Header().Get("Location"))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "127.0.0.1:54321", loc.Host)
	assert.Equal(s.T(), "/callback", loc.Path)
	assert.Equal(s.T(), "cli-state", loc.Query().Get("state"))
	code := loc.Query().Get("code")
	require.NotEmpty(s.T(), code)

	bad := s.exchange(code, redirectURI, "wrong-verifier")
	assert.Equal(s.T(), "invalid_grant", bad.Error)

	good := s.exchange(code, redirectURI, verifier)
	require.Empty(s.T(), good.Error)
	assert.Equal(s.T(), "Bearer", good.TokenType)
	assert.Equal(s.T(), s.testAcc.Email, good.Email)
	assert.Empty(s.T(), good.RefreshToken)
	assert.Equal(s.T(), s.service.Cfg.NuonAuthTokenTTL*60, good.ExpiresIn)

	var issued app.Token
	require.NoError(s.T(), s.service.DB.Where(&app.Token{Token: good.AccessToken}).First(&issued).Error)
	assert.Equal(s.T(), app.TokenTypeNuon, issued.TokenType)
	assert.Equal(s.T(), s.testAcc.ID, issued.AccountID)
	assert.Empty(s.T(), issued.Role)

	again := s.exchange(code, redirectURI, verifier)
	assert.Equal(s.T(), "invalid_grant", again.Error)
}

func (s *OAuthCLITestSuite) TestMCPTokenStaysOAuth() {
	session := s.session()
	client := app.OAuthClient{
		ClientName:              "mcp-token",
		RedirectURIs:            pq.StringArray{"https://claude.ai/api/mcp/auth_callback"},
		TokenEndpointAuthMethod: "none",
	}
	require.NoError(s.T(), s.service.DB.Create(&client).Error)

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	authCode := app.OAuthAuthorizationCode{
		RequestID:           "mcp-" + s.testAcc.ID,
		Code:                "mcp-code-" + s.testAcc.ID,
		ClientID:            client.ID,
		RedirectURI:         "https://claude.ai/api/mcp/auth_callback",
		CodeChallenge:       s256(verifier),
		CodeChallengeMethod: "S256",
		Scope:               string(app.RoleTypeOrgAdmin),
		AccountID:           s.testAcc.ID,
		SourceTokenID:       session.ID,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	}
	require.NoError(s.T(), s.service.DB.Create(&authCode).Error)

	resp := s.exchangeFor(client.ID, authCode.Code, authCode.RedirectURI, verifier)
	require.Empty(s.T(), resp.Error)
	assert.NotEmpty(s.T(), resp.RefreshToken)
	assert.Empty(s.T(), resp.Email)

	var issued app.Token
	require.NoError(s.T(), s.service.DB.Where(&app.Token{Token: resp.AccessToken}).First(&issued).Error)
	assert.Equal(s.T(), app.TokenTypeOAuth, issued.TokenType)
	assert.Equal(s.T(), string(app.RoleTypeOrgAdmin), issued.Role)
}

func (s *OAuthCLITestSuite) TestCLICompletePage() {
	session := s.session()

	ok := s.do(http.MethodGet, "/cli/complete?result=success", nil, &http.Cookie{Name: NuonAuthCookieName, Value: session.Token})
	require.Equal(s.T(), http.StatusOK, ok.Code)
	assert.Contains(s.T(), ok.Body.String(), "CLI Authorized")
	assert.Contains(s.T(), ok.Body.String(), s.testAcc.Email)

	denied := s.do(http.MethodGet, "/cli/complete?result=denied", nil)
	require.Equal(s.T(), http.StatusOK, denied.Code)
	assert.Contains(s.T(), denied.Body.String(), "CLI Access Denied")

	failed := s.do(http.MethodGet, "/cli/complete?result=%3Cscript%3E", nil)
	require.Equal(s.T(), http.StatusBadRequest, failed.Code)
	assert.NotContains(s.T(), failed.Body.String(), "<script>")
}

type oauthTokenHTTPResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Email        string `json:"email"`
	RefreshToken string `json:"refresh_token"`
	Error        string `json:"error"`
}

func (s *OAuthCLITestSuite) exchange(code, redirectURI, verifier string) oauthTokenHTTPResponse {
	return s.exchangeFor(nuonauth.CLIClientID, code, redirectURI, verifier)
}

func (s *OAuthCLITestSuite) exchangeFor(clientID, code, redirectURI, verifier string) oauthTokenHTTPResponse {
	rr := s.do(http.MethodPost, "/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	})
	var resp oauthTokenHTTPResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	return resp
}
