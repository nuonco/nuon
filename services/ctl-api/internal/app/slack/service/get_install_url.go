package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// defaultInstallScopes are the bot scopes the Nuon Slack app requests at
// install time. These match the surface area exercised by Phase 4
// (chat.postMessage / chat.update for lifecycle + approval messages,
// conversations.list for the /nuon subscribe channel picker, slash commands).
const defaultInstallScopes = "chat:write,channels:read,groups:read,commands"

// GetInstallURLResponse is the response body for the install-url endpoint.
type GetInstallURLResponse struct {
	URL string `json:"url"`
}

// @ID						GetSlackInstallURL
// @Summary				Get the Slack OAuth install URL for the current org
// @Description			Returns a Slack OAuth v2 authorize URL with a signed state JWT bound to the calling account and org. The dashboard redirects the user to this URL to begin the install flow.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Success				200	{object}	GetInstallURLResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/install-url [GET]
func (s *service) GetInstallURL(ctx *gin.Context) {
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if s.cfg.SlackClientID == "" {
		ctx.Error(stderr.ErrSystem{Err: fmt.Errorf("slack client id not configured")})
		return
	}
	if s.cfg.SlackOAuthRedirectURL == "" {
		ctx.Error(stderr.ErrSystem{Err: fmt.Errorf("slack oauth redirect url not configured")})
		return
	}

	nonce, err := newNonce()
	if err != nil {
		ctx.Error(fmt.Errorf("unable to generate nonce: %w", err))
		return
	}

	state, err := s.stateJWT.Issue(acct.ID, org.ID, nonce)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to issue state jwt: %w", err))
		return
	}

	q := url.Values{}
	q.Set("client_id", s.cfg.SlackClientID)
	q.Set("scope", defaultInstallScopes)
	q.Set("redirect_uri", s.cfg.SlackOAuthRedirectURL)
	q.Set("state", state)

	installURL := "https://slack.com/oauth/v2/authorize?" + q.Encode()
	ctx.JSON(http.StatusOK, GetInstallURLResponse{URL: installURL})
}

// newNonce returns a hex-encoded 16-byte random string for state-JWT binding.
func newNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
