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

const onboardingStateMarker = "onboarding"

func parseConnectState(state string) (orgID string, onboarding bool) {
	orgID, marker, _ := strings.Cut(state, ":")
	return orgID, marker == onboardingStateMarker
}

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
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing org_id"})
		return
	}
	if h.cfg.GithubAppName == "" {
		h.l.Error("github_app_name not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "github app not configured"})
		return
	}
	state := orgID
	if c.Query("onboarding") != "" {
		state = orgID + ":" + onboardingStateMarker
	}
	target := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", githubAppSlug(h.cfg.GithubAppName), url.QueryEscape(state))
	c.Redirect(http.StatusFound, target)
}

func (h *ConnectHandler) Handle(c *gin.Context) {
	installationID := c.Query("installation_id")
	orgID, onboarding := parseConnectState(c.Query("state"))

	fallback := "/"
	if orgIDPattern.MatchString(orgID) {
		fallback = fmt.Sprintf("/%s/apps", orgID)
	}
	if onboarding {
		fallback = "/onboarding?vcs-error=1"
	}

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.Redirect(http.StatusFound, fallback)
		return
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	connection, err := client.CreateVCSConnectionCallback(c.Request.Context(), &models.ServiceCreateConnectionCallbackRequest{
		GithubInstallID: &installationID,
		OrgID:           &orgID,
	})
	if err != nil {
		h.l.Error("vcs connection callback failed", zap.Error(err))
		c.Redirect(http.StatusFound, fallback)
		return
	}

	if onboarding {
		c.Redirect(http.StatusFound, "/onboarding?vcs-connected="+url.QueryEscape(connection.ID))
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/%s/apps?vcs-connected=%s", orgID, connection.ID))
}
