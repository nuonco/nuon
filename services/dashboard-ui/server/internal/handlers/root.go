package handlers

import (
	"context"
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

const firstRunJourney = "first_run"

func orgLandingPath(org *models.AppOrg) string {
	return "/" + org.ID
}

func firstRunPending(account *models.AppAccount) bool {
	if account == nil {
		return false
	}
	for _, journey := range account.UserJourneys {
		if journey == nil || journey.Name != firstRunJourney {
			continue
		}
		for _, step := range journey.Steps {
			if step != nil && !step.Complete {
				return true
			}
		}
		return false
	}
	return false
}

type currentUserGetter interface {
	GetCurrentUser(ctx context.Context) (*models.AppAccount, error)
}

func (h *RootHandler) orgDestination(ctx context.Context, client currentUserGetter, orgPath string) string {
	if !h.cfg.OnboardingFirstRun {
		return orgPath
	}
	account, err := client.GetCurrentUser(ctx)
	if err != nil {
		h.l.Warn("root: unable to read account journeys, skipping onboarding check", zap.Error(err))
		return orgPath
	}
	if firstRunPending(account) {
		return "/onboarding"
	}
	return orgPath
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
	start := time.Now()

	token, err := c.Cookie(authCookie)
	if (err != nil || token == "") && h.cfg.AuthServiceUrl != "" {
		h.l.Info("root: no auth token, redirecting to auth service")
		c.Redirect(http.StatusFound, h.cfg.AuthServiceUrl+"/?url="+url.QueryEscape(h.cfg.AppUrl))
		return
	}

	hasToken := token != ""
	_, orgCookieErr := c.Cookie(orgCookie)
	hasOrgCookie := orgCookieErr == nil

	h.l.Info("root: handling request",
		zap.Bool("has_token", hasToken),
		zap.Bool("has_org_cookie", hasOrgCookie),
	)

	if !h.cfg.OnboardingFirstRun {
		if orgId, err := c.Cookie(orgCookie); err == nil && orgId != "" {
			h.l.Info("root: redirecting to org from session cookie",
				zap.String("org_id", orgId),
				zap.Duration("duration", time.Since(start)),
			)
			c.Redirect(http.StatusFound, "/"+orgId)
			return
		}
	}

	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
	if err != nil {
		h.l.Error("root: failed to create nuon client",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)),
		)
		c.Redirect(http.StatusFound, "/error?reason=api-error")
		return
	}

	if orgId, err := c.Cookie(orgCookie); err == nil && orgId != "" {
		dest := h.orgDestination(c.Request.Context(), client, "/"+orgId)
		h.l.Info("root: redirecting to org from session cookie",
			zap.String("org_id", orgId),
			zap.String("destination", dest),
			zap.Duration("duration", time.Since(start)),
		)
		c.Redirect(http.StatusFound, dest)
		return
	}

	h.l.Info("root: no org cookie, fetching orgs from API")

	var orgs []*models.AppOrg
	for attempt := 0; attempt <= maxRetries; attempt++ {
		attemptStart := time.Now()
		orgs, _, err = client.GetOrgs(c.Request.Context(), &models.GetPaginatedQuery{Limit: 1})
		if err == nil {
			h.l.Info("root: GetOrgs succeeded",
				zap.Int("attempt", attempt+1),
				zap.Int("org_count", len(orgs)),
				zap.Duration("attempt_duration", time.Since(attemptStart)),
			)
			break
		}
		h.l.Warn("root: GetOrgs failed",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Bool("retryable", isRetryableError(err)),
			zap.Duration("attempt_duration", time.Since(attemptStart)),
		)
		if !isRetryableError(err) || attempt == maxRetries {
			break
		}
		time.Sleep(retryDelay)
	}

	if err != nil {
		h.l.Error("root: all attempts to get orgs failed, redirecting to error page",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)),
		)
		c.Redirect(http.StatusFound, "/error?reason=orgs-failed")
		return
	}

	if len(orgs) > 0 {
		dest := h.orgDestination(c.Request.Context(), client, orgLandingPath(orgs[0]))
		h.l.Info("root: redirecting to org",
			zap.String("org_id", orgs[0].ID),
			zap.String("destination", dest),
			zap.Duration("duration", time.Since(start)),
		)
		c.Redirect(http.StatusFound, dest)
		return
	}

	if h.cfg.IsBYOC {
		h.l.Info("root: no orgs found (BYOC), redirecting to setup",
			zap.Duration("duration", time.Since(start)),
		)
		c.Redirect(http.StatusFound, "/byoc-setup")
		return
	}

	h.l.Info("root: no orgs found, redirecting to onboarding",
		zap.Duration("duration", time.Since(start)),
	)
	c.Redirect(http.StatusFound, "/onboarding")
}
