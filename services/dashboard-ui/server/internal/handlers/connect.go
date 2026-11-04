package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/oidc/githubauth"
	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

var githubAppSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)

var orgIDPattern = regexp.MustCompile(`^[a-z0-9]+$`)

var installationIDPattern = regexp.MustCompile(`^[0-9]+$`)

func githubAppSlug(name string) string {
	slug := githubAppSlugPattern.ReplaceAllString(strings.ToLower(name), "-")
	return strings.Trim(slug, "-")
}

type ConnectHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewConnectHandler(cfg *internal.Config, l *zap.Logger) *ConnectHandler {
	return &ConnectHandler{cfg: cfg, l: l}
}

func (h *ConnectHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/connect", h.Handle)
	e.GET("/connect/oauth", h.HandleOAuth)
	e.GET("/api/connect-github", h.StartConnectGithub)
	return nil
}

func (h *ConnectHandler) StartConnectGithub(c *gin.Context) {
	orgID := c.Query("org_id")
	if !orgIDPattern.MatchString(orgID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing org_id"})
		return
	}
	onboarding := c.Query("onboarding") != ""

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		h.redirectToLogin(c)
		return
	}
	if h.cfg.GithubAppName == "" {
		h.l.Error("github connect is not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "github app not configured"})
		return
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.ServerAPIURL()), nuon.WithAuthToken(token), nuon.WithOrgID(orgID))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		h.redirectToLogin(c)
		return
	}
	if err := client.ValidateToken(c.Request.Context()); err != nil {
		h.l.Warn("github connect token rejected", zap.Error(err))
		h.redirectToLogin(c)
		return
	}
	if _, err := client.GetOrg(c.Request.Context()); err != nil {
		h.l.Warn("github connect org access rejected", zap.String("org_id", orgID), zap.Error(err))
		c.Redirect(http.StatusFound, connectFallback(orgID, onboarding))
		return
	}

	nonce, err := h.setConnectState(c, orgID, token, onboarding)
	if err != nil {
		h.l.Error("failed to set github connect state", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "github app not configured"})
		return
	}
	target := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", githubAppSlug(h.cfg.GithubAppName), url.QueryEscape(nonce))
	c.Redirect(http.StatusFound, target)
}

// Handle is the GitHub App setup URL. GitHub sends the user here after an
// install (or an update, with "Redirect on update" enabled) carrying
// installation_id and our state nonce. The id is not trusted on its own: when
// an OAuth client is configured the user is sent on to GitHub's authorize
// page, and HandleOAuth finishes the connection with the resulting code.
func (h *ConnectHandler) Handle(c *gin.Context) {
	installationID := c.Query("installation_id")
	token, _ := c.Cookie(authCookie)
	flow, err := h.openConnectState(c)
	fallback := connectFallback(flow.OrgID, flow.Onboarding)
	if err != nil || token == "" || !hmacEqual(tokenHash(token), flow.TokenHash) || !hmacEqual(flow.Nonce, c.Query("state")) {
		h.clearConnectState(c)
		c.Redirect(http.StatusFound, fallback)
		return
	}
	if !installationIDPattern.MatchString(installationID) {
		h.clearConnectState(c)
		h.l.Warn("github setup redirect without a usable installation_id", zap.String("org_id", flow.OrgID))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	if h.cfg.GithubAppClientID == "" {
		// No OAuth client: hand the id to ctl-api as-is. ctl-api rejects it
		// unless its own ownership check is disabled (local stacks only).
		h.clearConnectState(c)
		h.finishConnect(c, flow, token, installationID, "")
		return
	}

	flow.InstallationID = installationID
	nonce, err := h.writeConnectState(c, flow)
	if err != nil {
		h.l.Error("failed to set github oauth state", zap.Error(err))
		h.clearConnectState(c)
		c.Redirect(http.StatusFound, fallback)
		return
	}
	c.Redirect(http.StatusFound, githubauth.AuthorizeURL(h.cfg.GithubAppClientID, h.oauthRedirectURI(), nonce))
}

// HandleOAuth is the GitHub App callback URL for the user-authorization flow
// started by Handle. It forwards the code together with the installation id
// stashed in the state cookie, so ctl-api can prove the two belong together.
func (h *ConnectHandler) HandleOAuth(c *gin.Context) {
	code := c.Query("code")
	token, _ := c.Cookie(authCookie)
	flow, err := h.openConnectState(c)
	h.clearConnectState(c)
	fallback := connectFallback(flow.OrgID, flow.Onboarding)
	if err != nil || token == "" || !hmacEqual(tokenHash(token), flow.TokenHash) || !hmacEqual(flow.Nonce, c.Query("state")) {
		c.Redirect(http.StatusFound, fallback)
		return
	}
	if code == "" || !installationIDPattern.MatchString(flow.InstallationID) {
		h.l.Warn("github oauth redirect without code or installation",
			zap.String("org_id", flow.OrgID),
			zap.Bool("has_code", code != ""),
			zap.Bool("has_installation", flow.InstallationID != ""),
		)
		c.Redirect(http.StatusFound, fallback)
		return
	}

	h.finishConnect(c, flow, token, flow.InstallationID, code)
}

func (h *ConnectHandler) finishConnect(c *gin.Context, flow connectState, token, installationID, code string) {
	fallback := connectFallback(flow.OrgID, flow.Onboarding)

	client, err := nuon.New(nuon.WithURL(h.cfg.ServerAPIURL()), nuon.WithAuthToken(token), nuon.WithOrgID(flow.OrgID))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	connection, err := client.CreateVCSConnectionCallback(c.Request.Context(), &models.ServiceCreateConnectionCallbackRequest{
		GithubInstallID: &installationID,
		OrgID:           &flow.OrgID,
		Code:            code,
	})
	if err != nil {
		h.l.Error("vcs connection callback failed", zap.Error(err))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	if flow.Onboarding {
		c.Redirect(http.StatusFound, "/onboarding?vcs-connected="+url.QueryEscape(connection.ID))
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/%s/apps?vcs-connected=%s", flow.OrgID, connection.ID))
}

// oauthRedirectURI must match a callback URL registered on the GitHub App.
func (h *ConnectHandler) oauthRedirectURI() string {
	return strings.TrimRight(h.cfg.AppUrl, "/") + "/connect/oauth"
}

func (h *ConnectHandler) redirectToLogin(c *gin.Context) {
	if h.cfg.AuthServiceUrl == "" || h.cfg.AppUrl == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}
	returnURL := h.cfg.AppUrl + c.Request.URL.RequestURI()
	c.Redirect(http.StatusFound, h.cfg.AuthServiceUrl+"/?url="+url.QueryEscape(returnURL))
}

func connectFallback(orgID string, onboarding bool) string {
	if onboarding {
		return "/onboarding?vcs-error=1"
	}
	if orgIDPattern.MatchString(orgID) {
		return "/" + orgID + "/apps"
	}
	return "/"
}
