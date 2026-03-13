package handlers

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const authCookie = "X-Nuon-Auth"
const orgCookie = "org_session"

type RootHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewRootHandler(cfg *internal.Config, l *zap.Logger) *RootHandler {
	return &RootHandler{cfg: cfg, l: l}
}

func (h *RootHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/", h.Handle)
	return nil
}

func (h *RootHandler) Handle(c *gin.Context) {
	token, err := c.Cookie(authCookie)
	if (err != nil || token == "") && h.cfg.AuthServiceUrl != "" {
		c.Redirect(http.StatusFound, h.cfg.AuthServiceUrl+"/?url="+url.QueryEscape(h.cfg.AppUrl))
		return
	}

	if orgId, err := c.Cookie(orgCookie); err == nil && orgId != "" {
		orgClient, err := nuon.New(
			nuon.WithURL(h.cfg.APIUrl),
			nuon.WithAuthToken(token),
			nuon.WithOrgID(orgId),
		)
		if err == nil {
			if _, err := orgClient.GetOrg(c.Request.Context()); err == nil {
				c.Redirect(http.StatusFound, "/"+orgId+"/apps")
				return
			}
		}
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.Redirect(http.StatusFound, "/onboarding")
		return
	}

	orgs, _, err := client.GetOrgs(c.Request.Context(), nil)
	if err == nil && len(orgs) > 0 {
		c.Redirect(http.StatusFound, "/"+orgs[0].ID+"/apps")
		return
	}

	c.Redirect(http.StatusFound, "/onboarding")
}
