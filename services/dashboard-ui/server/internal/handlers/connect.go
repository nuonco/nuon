package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

var githubAppSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)

var orgIDPattern = regexp.MustCompile(`^[a-z0-9]+$`)

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

	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token), nuon.WithOrgID(orgID))
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

func (h *ConnectHandler) Handle(c *gin.Context) {
	installationID := c.Query("installation_id")
	token, _ := c.Cookie(authCookie)
	flow, err := h.openConnectState(c)
	h.clearConnectState(c)
	fallback := connectFallback(flow.OrgID, flow.Onboarding)
	if err != nil || token == "" || !hmacEqual(tokenHash(token), flow.TokenHash) || !hmacEqual(flow.Nonce, c.Query("state")) {
		c.Redirect(http.StatusFound, fallback)
		return
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token), nuon.WithOrgID(flow.OrgID))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	connection, err := client.CreateVCSConnectionCallback(c.Request.Context(), &models.ServiceCreateConnectionCallbackRequest{
		GithubInstallID: &installationID,
		OrgID:           &flow.OrgID,
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
