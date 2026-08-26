package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// StackServiceAccountResponse identifies the account an install stack authenticates
// as, and whether it holds a usable token. Never the token value — those are
// returned once, by POST /v1/service-accounts/{account_id}/tokens.
type StackServiceAccountResponse struct {
	AccountID string `json:"account_id,omitzero"`
	Email     string `json:"email,omitzero"`

	// HasLiveToken is false whether no token was ever created or every one has
	// expired or been revoked; the caller fixes both the same way.
	HasLiveToken bool `json:"has_live_token"`

	// ExpiresAt is the expiry of the longest-lived usable token; zero when
	// HasLiveToken is false.
	ExpiresAt time.Time `json:"expires_at,omitzero"`
}

// @ID						GetStackServiceAccount
// @Summary				get an install stack's service account
// @Description			Return the service account an install stack's Terraform module authenticates as, and whether it holds a usable API token. Never returns a token value: create one with POST /v1/service-accounts/{account_id}/tokens, which returns it once.
// @Param					install_id	path	string	true	"install ID"
// @Tags					stacks
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	StackServiceAccountResponse
// @Router					/v1/stacks/{install_id}/service-account [get]
func (s *service) GetStackServiceAccount(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve org from request: %w", err))
		return
	}

	// Scoped to the caller's org, and not-found on a mismatch so this cannot probe
	// which install IDs exist elsewhere.
	var stack app.InstallStack
	if res := s.db.WithContext(ctx).
		Where(app.InstallStack{InstallID: installID, OrgID: orgID}).
		Order("created_at DESC").
		First(&stack); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "install stack not found"})
			return
		}
		ctx.Error(fmt.Errorf("load install stack: %w", res.Error))
		return
	}

	acct, err := s.acctClient.FindAccount(ctx, account.ServiceAccountEmail(stack.ID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{
				Err:         err,
				Description: "this stack has no service account yet; reprovision the install to create one",
			})
			return
		}
		ctx.Error(fmt.Errorf("load stack service account: %w", err))
		return
	}

	expiresAt, err := liveStackTokenExpiry(ctx, s.db, acct.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, StackServiceAccountResponse{
		AccountID:    acct.ID,
		Email:        acct.Email,
		HasLiveToken: !expiresAt.IsZero(),
		ExpiresAt:    expiresAt,
	})
}

// liveStackTokenExpiry returns the expiry of the account's longest-lived usable
// token, or zero if it has none. Ordered by expiry, not creation: a fresh 1-day
// token can sit alongside an older 1-year one.
func liveStackTokenExpiry(ctx context.Context, db *gorm.DB, accountID string) (time.Time, error) {
	var live app.Token
	res := db.WithContext(ctx).
		Where(app.Token{AccountID: accountID}).
		Where("expires_at > ?", time.Now()).
		Order("expires_at DESC").
		First(&live)
	if res.Error == nil {
		return live.ExpiresAt, nil
	}
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return time.Time{}, nil
	}

	return time.Time{}, fmt.Errorf("load stack token: %w", res.Error)
}
