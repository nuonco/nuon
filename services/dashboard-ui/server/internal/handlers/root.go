package handlers

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const authCookie = "X-Nuon-Auth"
const orgCookie = "org_session"

func orgLandingPath(org *models.AppOrg) string {
	if org.Features != nil && org.Features["org-dashboard"] {
		return "/" + org.ID + ""
	}
	return "/" + org.ID + "/apps"
}

type RootHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewRootHandler(cfg *internal.Config, l *zap.Logger) *RootHandler {
	return &RootHandler{cfg: cfg, l: l}
}

func (h *RootHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/", h.Handle)
	e.GET("/api/auth/login", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})
	return nil
}

func isRetryableError(err error) bool {
	type coder interface {
		Code() int
	}
	if c, ok := err.(coder); ok {
		code := c.Code()
		return code >= 500 && code < 600
	}
	return isTimeoutError(err)
}

func isTimeoutError(err error) bool {
	type timeouter interface {
		Timeout() bool
	}
	if t, ok := err.(timeouter); ok {
		return t.Timeout()
	}
	return false
}

const maxRetries = 2
const retryDelay = 500 * time.Millisecond

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
			if org, err := orgClient.GetOrg(c.Request.Context()); err == nil {
				c.Redirect(http.StatusFound, orgLandingPath(org))
				return
			} else {
				h.l.Warn("failed to get org from session cookie",
					zap.String("org_id", orgId),
					zap.Error(err),
				)
			}
		}
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.Redirect(http.StatusFound, "/onboarding")
		return
	}

	var orgs []*models.AppOrg
	for attempt := 0; attempt <= maxRetries; attempt++ {
		orgs, _, err = client.GetOrgs(c.Request.Context(), &models.GetPaginatedQuery{Limit: 1})
		if err == nil {
			break
		}
		h.l.Warn("failed to get orgs",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
		)
		if !isRetryableError(err) || attempt == maxRetries {
			break
		}
		time.Sleep(retryDelay)
	}

	if err != nil {
		h.l.Error("all attempts to get orgs failed, redirecting to onboarding",
			zap.Error(err),
		)
		c.Redirect(http.StatusFound, "/onboarding")
		return
	}

	if len(orgs) > 0 {
		c.Redirect(http.StatusFound, orgLandingPath(orgs[0]))
		return
	}

	h.l.Info("no orgs found for account, redirecting to onboarding")
	c.Redirect(http.StatusFound, "/onboarding")
}
