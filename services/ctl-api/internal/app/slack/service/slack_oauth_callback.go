package service

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

// oauthErrorPage is the HTML shown to the user in their browser when the OAuth
// callback fails. Slack opens the redirect in a browser tab so we render a
// human-readable page rather than a JSON error envelope.
var oauthErrorPage = template.Must(template.New("slack-oauth-error").Parse(`<!doctype html>
<html><head><meta charset="utf-8"><title>Slack install failed</title>
<style>body{font-family:system-ui,sans-serif;max-width:540px;margin:64px auto;color:#222}
h1{font-size:20px}p{line-height:1.45}code{background:#f3f3f3;padding:2px 4px;border-radius:3px}</style>
</head><body>
<h1>Slack install failed</h1>
<p>{{.Message}}</p>
<p>You can close this tab and retry the install from the Nuon dashboard.</p>
</body></html>`))

// SlackOAuthCallback is the redirect target for the Slack OAuth v2 install
// flow. It is registered on the dedicated Slack listener (NOT the public API)
// and must be reachable from the public internet at cfg.SlackOAuthRedirectURL.
//
// Trust model: there is NO Slack signing-secret on this request — Slack does
// not sign OAuth redirects. Instead, we trust the `state` parameter, which is
// an HS256-signed JWT issued by GetInstallURL bound to (account_id, org_id).
// On signature/expiry failure we render an HTML error page.
//
// On success: upserts SlackInstallation by team_id (the workspace identity
// invariant), upserts a verified SlackOrgLink for (team_id, claims.OrgID),
// then redirects the user back to the dashboard.
//
//	@ID						SlackOAuthCallback
//	@Summary				Slack OAuth v2 redirect target
//	@Description			Receives the OAuth `code` + signed-state JWT from Slack, exchanges the code for a workspace bot token via oauth.v2.access, persists the installation + org link, and redirects to the dashboard. NOT signed by Slack — trust comes from the state JWT signature. Enterprise Grid (org-wide) installs are rejected.
//	@Tags					slack
//	@Accept					json
//	@Produce				html
//	@Param					code	query	string	true	"OAuth authorization code from Slack"
//	@Param					state	query	string	true	"Signed state JWT issued by GetInstallURL"
//	@Success				302	"Redirect to dashboard with ?slack=installed"
//	@Failure				400	"HTML error page"
//	@Failure				401	"HTML error page"
//	@Failure				500	"HTML error page"
//	@Router					/slack/oauth/callback [GET]
func (s *service) SlackOAuthCallback(ctx *gin.Context) {
	// Check server-side config before consuming the (single-use) state JWT
	// or the OAuth code. Otherwise a misconfigured server would burn the
	// user's install attempt and force a full retry from the dashboard.
	if s.cfg.SlackClientID == "" || s.cfg.SlackClientSecret == "" {
		s.l.Error("slack oauth callback invoked without client credentials configured")
		s.renderOAuthError(ctx, http.StatusInternalServerError, "Slack integration is not configured on this server.")
		return
	}

	code := ctx.Query("code")
	state := ctx.Query("state")
	if code == "" || state == "" {
		s.renderOAuthError(ctx, http.StatusBadRequest, "Missing code or state parameter from Slack.")
		return
	}

	claims, err := s.stateJWT.Decode(state)
	if err != nil {
		s.l.Warn("slack oauth: state jwt decode failed", zap.Error(err))
		s.renderOAuthError(ctx, http.StatusUnauthorized, "The install link has expired or is invalid. Please retry from the Nuon dashboard.")
		return
	}

	resp, err := s.slackClient.OAuthV2Access(ctx, slackclient.OAuthV2AccessRequest{
		ClientID:     s.cfg.SlackClientID,
		ClientSecret: s.cfg.SlackClientSecret,
		Code:         code,
		RedirectURI:  s.cfg.SlackOAuthRedirectURL,
	})
	if err != nil {
		s.l.Warn("slack oauth.v2.access failed", zap.Error(err))
		s.renderOAuthError(ctx, http.StatusBadRequest, "Slack rejected the install: "+err.Error())
		return
	}

	// Reject Enterprise Grid org-wide installs. Phase 4 invariant: team_id
	// uniquely identifies a workspace; org-wide installs break that.
	if resp.IsEnterpriseInstall {
		s.l.Warn("slack oauth: rejecting enterprise-grid install",
			zap.String("team_id", resp.Team.ID))
		s.renderOAuthError(ctx, http.StatusBadRequest,
			"Enterprise Grid (org-wide) installs are not supported. Please install the Nuon app to a single workspace instead.")
		return
	}

	if err := s.persistSlackInstall(ctx, claims.AccountID, claims.OrgID, resp); err != nil {
		s.l.Error("slack oauth: persist install failed",
			zap.Error(err),
			zap.String("team_id", resp.Team.ID),
			zap.String("account_id", claims.AccountID),
			zap.String("org_id", claims.OrgID))
		s.renderOAuthError(ctx, http.StatusInternalServerError, "We couldn't save the install. Please retry; if it persists contact support.")
		return
	}

	ctx.Redirect(http.StatusFound, s.postInstallRedirectURL(claims.OrgID))
}

// persistSlackInstall upserts the SlackInstallation row keyed by team_id
// (re-install path clears DeletedAt) and upserts the verified SlackOrgLink for
// (team_id, orgID). Both writes happen inside a single transaction so a
// partial failure can't leave a fresh installation row without an org link.
// The calling account is stamped on the gorm context so the BeforeCreate
// hooks resolve CreatedByID via createdByIDFromContext.
func (s *service) persistSlackInstall(
	ctx context.Context,
	accountID, orgID string,
	resp *slackclient.OAuthV2AccessResponse,
) error {
	// The OAuth callback is hit by Slack, not by an authenticated dashboard
	// user, so there is no middleware-resolved account in context. We
	// resolve the account ourselves from the state JWT claim and stamp it
	// on the context so the BeforeCreate hooks can pick it up.
	var acct app.Account
	if err := s.db.WithContext(ctx).
		Where(app.Account{ID: accountID}).
		First(&acct).Error; err != nil {
		return fmt.Errorf("resolve installer account %q: %w", accountID, err)
	}
	ctxWithAcct := cctx.SetAccountContext(ctx, &acct)

	var enterpriseID string
	if resp.Enterprise != nil {
		enterpriseID = resp.Enterprise.ID
	}

	return s.db.WithContext(ctxWithAcct).Transaction(func(tx *gorm.DB) error {
		// Upsert SlackInstallation by team_id. Use Unscoped to find a soft-deleted
		// prior installation (the unique index is on (team_id, deleted_at) so a
		// previously uninstalled record may exist with deleted_at != 0).
		var existing app.SlackInstallation
		res := tx.
			Unscoped().
			Where(app.SlackInstallation{TeamID: resp.Team.ID}).
			First(&existing)
		switch {
		case errors.Is(res.Error, gorm.ErrRecordNotFound):
			install := &app.SlackInstallation{
				TeamID:                 resp.Team.ID,
				TeamName:               resp.Team.Name,
				EnterpriseID:           enterpriseID,
				BotUserID:              resp.BotUserID,
				AppID:                  resp.AppID,
				Scope:                  resp.Scope,
				Status:                 app.SlackInstallationStatusActive,
				BotAccessToken:         resp.AccessToken,
				InstalledBySlackUserID: resp.AuthedUser.ID,
				InstalledByAccountID:   accountID,
			}
			if err := tx.Create(install).Error; err != nil {
				return fmt.Errorf("create installation: %w", err)
			}
		case res.Error != nil:
			return fmt.Errorf("lookup existing installation: %w", res.Error)
		default:
			// Re-install path: clear soft-delete tombstone, refresh tokens /
			// scope / metadata, and reset Status to active.
			existing.TeamName = resp.Team.Name
			existing.EnterpriseID = enterpriseID
			existing.BotUserID = resp.BotUserID
			existing.AppID = resp.AppID
			existing.Scope = resp.Scope
			existing.Status = app.SlackInstallationStatusActive
			existing.BotAccessToken = resp.AccessToken
			existing.InstalledBySlackUserID = resp.AuthedUser.ID
			existing.InstalledByAccountID = accountID
			existing.DeletedAt = 0
			if err := tx.Unscoped().Save(&existing).Error; err != nil {
				return fmt.Errorf("update installation: %w", err)
			}
		}

		// Upsert verified SlackOrgLink for (team_id, orgID). Same Unscoped
		// pattern — a previously revoked / soft-deleted link should be reused
		// rather than a duplicate row created.
		var existingLink app.SlackOrgLink
		linkRes := tx.
			Unscoped().
			Where(app.SlackOrgLink{TeamID: resp.Team.ID, OrgID: orgID}).
			First(&existingLink)
		switch {
		case errors.Is(linkRes.Error, gorm.ErrRecordNotFound):
			link := &app.SlackOrgLink{
				TeamID:            resp.Team.ID,
				OrgID:             orgID,
				Status:            app.SlackOrgLinkStatusVerified,
				LinkedByAccountID: accountID,
			}
			if err := tx.Create(link).Error; err != nil {
				return fmt.Errorf("create org link: %w", err)
			}
		case linkRes.Error != nil:
			return fmt.Errorf("lookup existing org link: %w", linkRes.Error)
		default:
			existingLink.Status = app.SlackOrgLinkStatusVerified
			existingLink.LinkedByAccountID = accountID
			existingLink.DeletedAt = 0
			if err := tx.Unscoped().Save(&existingLink).Error; err != nil {
				return fmt.Errorf("update org link: %w", err)
			}
		}

		return nil
	})
}

// postInstallRedirectURL returns the dashboard URL the user is redirected to
// after a successful install. We deep-link to the org's Slack integration page
// (/<orgID>/slack) and append ?slack=installed so the page can show a success
// toast. If orgID is empty we fall back to the dashboard root.
func (s *service) postInstallRedirectURL(orgID string) string {
	base := strings.TrimRight(s.cfg.AppURL, "/")
	if base == "" {
		// AppURL is required at config validation time, but be defensive:
		// fall back to "/" rather than minting a broken redirect.
		base = "/"
	}
	q := url.Values{}
	q.Set("slack", "installed")
	if orgID == "" {
		return base + "/?" + q.Encode()
	}
	return base + "/" + orgID + "/slack?" + q.Encode()
}

// renderOAuthError writes an HTML error page back to the user's browser.
func (s *service) renderOAuthError(ctx *gin.Context, status int, msg string) {
	ctx.Status(status)
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	if err := oauthErrorPage.Execute(ctx.Writer, struct{ Message string }{Message: msg}); err != nil {
		s.l.Warn("slack oauth: render error page failed", zap.Error(err))
	}
}
