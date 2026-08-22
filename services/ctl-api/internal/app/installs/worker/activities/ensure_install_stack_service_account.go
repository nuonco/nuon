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

// stackTokenTimeout is the lifetime of an install stack's API token. Deliberately
// short: unlike the phone-home token — which backs a Lambda that can be re-invoked
// years after the stack was applied, and is therefore effectively permanent — this
// one is handed to a customer to paste into a shell or a CI secret, where it is far
// more exposed and far harder to account for.
//
// One day means a leaked token is worth little, and it pushes anything recurring
// toward OIDC, which mints per-run credentials and stores nothing. The cost is that
// a customer returning to re-apply needs a fresh token from the dashboard;
// reconciliation mints one whenever no live token exists.
const stackTokenTimeout = time.Hour * 24

type EnsureInstallStackServiceAccountRequest struct {
	InstallStackID string `json:"install_stack_id" validate:"required"`
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
// @by-field InstallStackID
// @start-to-close-timeout 2m
func (a *Activities) EnsureInstallStackServiceAccount(
	ctx context.Context, req *EnsureInstallStackServiceAccountRequest,
) (*EnsureInstallStackServiceAccountResponse, error) {
	if err := a.v.StructCtx(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	var stack app.InstallStack
	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{ID: req.InstallStackID}).
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

	hasLive, err := hasLiveStackToken(ctx, a.db, acct.ID)
	if err != nil {
		return nil, err
	}
	if hasLive {
		return resp, nil
	}

	if _, err := a.acctClient.CreateToken(
		ctx, account.ServiceAccountEmail(stack.ID), stackTokenTimeout,
	); err != nil {
		return nil, fmt.Errorf("unable to mint stack token: %w", err)
	}
	resp.TokenMinted = true

	return resp, nil
}

// hasLiveStackToken reports whether an account still holds a usable token. This is
// the whole mint decision, so every way a token can stop counting has to be handled
// here: gorm's soft delete filters revoked rows out of the query, and the expiry
// bound filters out ones that aged out. A false here re-mints.
func hasLiveStackToken(ctx context.Context, db *gorm.DB, accountID string) (bool, error) {
	var live app.Token
	res := db.WithContext(ctx).
		Where(app.Token{AccountID: accountID}).
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&live)
	if res.Error == nil {
		return true, nil
	}
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return false, generics.TemporalGormError(res.Error, "unable to look up stack token: %w")
}
