package service

import (
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

// StackTokenResponse carries the install stack's API token. The value is returned in
// full because its only purpose is to be pasted into a Terraform provider block —
// app.Token tags the field json:"-", so it has to be surfaced explicitly here.
type StackTokenResponse struct {
	ID        string    `json:"id,omitzero"`
	APIToken  string    `json:"api_token,omitzero"`
	ExpiresAt time.Time `json:"expires_at,omitzero"`
}

// @ID						GetStackToken
// @Summary				get an install stack's API token
// @Description			Return the API token the install stack's Terraform module uses to authenticate against the runner API. Read-only: the token is minted during stack version generation and is not created or rotated by this endpoint.
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
// @Success				200	{object}	StackTokenResponse
// @Router					/v1/stacks/{install_id}/token [get]
func (s *service) GetStackToken(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve org from request: %w", err))
		return
	}

	// Scoped to the caller's org, and reported as not-found on a mismatch so this
	// cannot be used to probe which install IDs exist elsewhere.
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
			ctx.Error(stderr.ErrNotFound{Err: err, Description: "this stack has no service account yet"})
			return
		}
		ctx.Error(fmt.Errorf("load stack service account: %w", err))
		return
	}

	// Deliberately does not mint. Minting belongs to stack version generation, which
	// is convergent and idempotent; issuing one here would let a dashboard refresh
	// create credentials, and the customer's Terraform holds whichever came first.
	var token app.Token
	if res := s.db.WithContext(ctx).
		Where(app.Token{AccountID: acct.ID}).
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&token); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{
				Err:         res.Error,
				Description: "this stack has no live API token; reprovision the install to mint one",
			})
			return
		}
		ctx.Error(fmt.Errorf("load stack token: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, StackTokenResponse{
		ID:        token.ID,
		APIToken:  token.Token,
		ExpiresAt: token.ExpiresAt,
	})
}
