package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// stackTokenTimeout is the lifetime of an install stack's API token. Unlike the
// phone-home token — which backs a Lambda that can be re-invoked years after the
// stack was applied, and is therefore effectively permanent — this one is handed to
// the customer to put in their Terraform configuration, so it gets the same 90 days
// as a runner token. Rotation is not handled yet; an expired token needs a new one
// minted out of band.
const stackTokenTimeout = time.Hour * 24 * 90

type EnsureInstallStackServiceAccountRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

type EnsureInstallStackServiceAccountResponse struct {
	AccountID string `json:"account_id"`

	// TokenMinted reports whether this call issued a token. The token value is
	// deliberately absent: activity results are persisted in Temporal history, and
	// EnsureInstallPhoneHomeSecret keeps credentials out of it for the same reason.
	// Read the value from the tokens table when it needs to be shown.
	TokenMinted bool `json:"token_minted"`
}

// EnsureInstallStackServiceAccount reconciles the service account and API token that
// let an install stack read its own configuration from the runner API.
//
// The account is keyed on the InstallStack, not the InstallStackVersion. A stack
// regenerates its version on every config change, and re-keying per version would
// mint a new token each time and invalidate the customer's Terraform on every sync.
// This is a different grain from InstallStackVersion.PhoneHomeTokenID by design —
// the two credentials authenticate different directions and should not be merged.
//
// Convergent, not a create hook: desired state is derived from Postgres on every
// call, so a provision, a reprovision, and a retry after a partial failure all reach
// the same place. The mint decision keys on whether a live token exists rather than
// on whether the account does — those are written in the same step but can diverge
// when minting fails after the account is created, and keying on the account would
// strand such a stack with no credential and no way to notice.
//
// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 2m
func (a *Activities) EnsureInstallStackServiceAccount(
	ctx context.Context, req *EnsureInstallStackServiceAccountRequest,
) (*EnsureInstallStackServiceAccountResponse, error) {
	if err := a.v.StructCtx(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	var stack app.InstallStack
	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{InstallID: req.InstallID}).
		First(&stack); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to load install stack: %w")
	}

	acct, err := a.acctClient.EnsureServiceAccount(ctx, stack.ID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to ensure stack service account: %w", err)
	}

	// The stack reads install configuration and reports outputs back, which today has
	// no narrower role than org admin. Re-applied on every call so a stack created
	// before the role existed picks it up on its next reconcile.
	if err := a.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, stack.OrgID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to grant stack service account org role: %w", err)
	}

	resp := &EnsureInstallStackServiceAccountResponse{AccountID: acct.ID}

	var live app.Token
	res := a.db.WithContext(ctx).
		Where(app.Token{AccountID: acct.ID}).
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&live)
	if res.Error == nil {
		return resp, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, generics.TemporalGormError(res.Error, "unable to look up stack token: %w")
	}

	if _, err := a.acctClient.CreateToken(
		ctx, account.ServiceAccountEmail(stack.ID), stackTokenTimeout,
	); err != nil {
		return nil, fmt.Errorf("unable to mint stack token: %w", err)
	}
	resp.TokenMinted = true

	return resp, nil
}
