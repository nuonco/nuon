package authz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	strictjson "sigs.k8s.io/json"

	"github.com/nuonco/nuon/pkg/analytics/events"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type inviteAuthorization struct {
	Credential cctx.VerifiedCredential `json:"credential"`
	RoleID     string                  `json:"role_id"`
	RoleType   app.RoleType            `json:"role_type"`
	Scope      permissions.Scope       `json:"scope"`
}

func (h *Client) AuthorizeInvite(ctx context.Context, invite *app.OrgInvite) error {
	if _, err := RoleBindingScope(invite.OrgID, &invite.Scope); err != nil {
		return stderr.NewInvalidRequest(err)
	}
	if invite.RoleType == "" {
		invite.RoleType = app.RoleTypeOrgReadOnly
	}
	role, err := h.ResolveDelegableRoleAtScope(ctx, invite.Scope, invite.RoleType, app.RoleContextTeam)
	if err != nil {
		return err
	}
	credential, ok := cctx.VerifiedCredentialFromContext(ctx)
	if !ok || len(credential.Ceiling) == 0 {
		return fmt.Errorf("verified delegation credential required")
	}
	proof, err := json.Marshal(inviteAuthorization{Credential: credential, RoleID: role.ID, RoleType: role.RoleType, Scope: invite.Scope})
	if err != nil {
		return err
	}
	invite.Authorization = proof
	return nil
}

func (h *Client) AcceptInvite(ctx context.Context, invite *app.OrgInvite, acct *app.Account) error {
	denied := func() error {
		return stderr.ErrAuthorization{Err: fmt.Errorf("invite acceptance requires a current authorized invitation and recipient session"), Description: "ask an org administrator to resend the invitation and sign in again"}
	}
	recipient, ok := cctx.VerifiedCredentialFromContext(ctx)
	if !ok || !customerCredentialAllowed(recipient) || recipient.Profile != cctx.ProfileHuman || acct == nil || recipient.AccountID != acct.ID || invite == nil || invite.ID == "" {
		return denied()
	}
	proof := bytes.Clone(invite.Authorization)
	var authorization inviteAuthorization
	strictErrors, err := strictjson.UnmarshalStrict(proof, &authorization)
	if err != nil || len(strictErrors) != 0 || authorization.RoleID == "" || authorization.RoleType != invite.RoleType || authorization.Scope != invite.Scope {
		return denied()
	}
	if _, err := RoleBindingScope(invite.OrgID, &authorization.Scope); err != nil {
		return denied()
	}
	sourceCtx := cctx.SetVerifiedCredentialContext(ctx, authorization.Credential)
	sourceCtx = cctx.SetAccountContext(sourceCtx, &app.Account{ID: authorization.Credential.AccountID})
	decision, err := h.MutateCustomer(sourceCtx, MutationRequest{
		Permission:       permissions.OrgWrite,
		Target:           permissions.Scope{Kind: permissions.KindOrg, OrgID: invite.OrgID, ID: invite.OrgID},
		SubjectAccountID: recipient.AccountID,
	}, func(tx *gorm.DB, _ permissions.ResolvedTarget) error {
		client := New(Params{DB: tx})
		role, err := client.ResolveDelegableRoleAtScope(tx.Statement.Context, authorization.Scope, authorization.RoleType, app.RoleContextTeam)
		if err != nil {
			return err
		}
		if role.ID != authorization.RoleID {
			return denied()
		}
		if err := RequireUserManagedInOrg(tx, recipient.AccountID, invite.OrgID); err != nil {
			return err
		}
		var account app.Account
		if err := tx.Where(app.Account{ID: recipient.AccountID}).Take(&account).Error; err != nil {
			return err
		}
		var token app.Token
		if err := tx.Select("id", "account_id", "org_id", "token_type", "issued_at", "expires_at", "source_type", "source_id", "authorization").
			Clauses(clause.Locking{Strength: "SHARE"}).Where(app.Token{ID: recipient.TokenID, AccountID: recipient.AccountID}).Take(&token).Error; err != nil {
			return err
		}
		if account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0 {
			return denied()
		}
		if (token.TokenType != app.TokenTypeNuon && token.TokenType != app.TokenTypeAuth && token.TokenType != app.TokenTypeAuth0) ||
			token.SourceType != "" || token.SourceID != "" || !time.Now().Before(token.ExpiresAt) || token.IssuedAt.After(time.Now()) ||
			(token.OrgID != "" && token.OrgID != invite.OrgID) {
			return denied()
		}
		if _, reason := customerTokenCeiling(recipient, account, token, invite.OrgID, time.Now()); reason != ReasonAllowed {
			return denied()
		}
		var current app.OrgInvite
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.OrgInvite{ID: invite.ID, OrgID: invite.OrgID}).Take(&current).Error; err != nil {
			return err
		}
		if current.Status != app.OrgInviteStatusPending || current.Email != account.Email || current.RoleType != authorization.RoleType || current.Scope != authorization.Scope || !bytes.Equal(current.Authorization, proof) {
			return denied()
		}
		var memberships int64
		if err := tx.Model(&app.AccountRole{}).Where(app.AccountRole{AccountID: account.ID, OrgID: generics.NewNullString(current.OrgID)}).Count(&memberships).Error; err != nil {
			return err
		}
		if memberships == 0 {
			if err := client.AddAccountRoleAtScope(tx.Statement.Context, account.ID, role.ID, authorization.Scope); err != nil {
				return err
			}
		}
		res := tx.Model(&app.OrgInvite{}).Where(app.OrgInvite{ID: current.ID, OrgID: current.OrgID, Status: app.OrgInviteStatusPending}).Update("status", app.OrgInviteStatusAccepted)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return denied()
		}
		current.Status = app.OrgInviteStatusAccepted
		*invite = current
		return nil
	})
	if err != nil {
		return err
	}
	if !decision.Allowed {
		return denied()
	}

	if h.analyticsClient != nil {
		h.analyticsClient.Track(ctx, events.InviteAccepted, map[string]interface{}{
			"invite_id": invite.ID,
			"email":     invite.Email,
			"org_id":    invite.OrgID,
			"role_type": invite.RoleType,
		})
	}

	return nil
}
