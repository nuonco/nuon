package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CustomerTokenRequest struct {
	AccountID  string
	OrgID      string
	Name       string
	Role       app.RoleType
	Scope      permissions.Scope
	Duration   time.Duration
	Invalidate bool
}

func (c *Client) IssueCustomerToken(ctx context.Context, req CustomerTokenRequest) (*app.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if req.AccountID == "" || req.Role == "" || req.Duration <= 0 || req.Duration > 365*24*time.Hour {
		return nil, stderr.NewInvalidRequest(fmt.Errorf("token requires an account, a role and a positive duration of at most one year"))
	}
	scope, err := authz.RoleBindingScope(req.OrgID, &req.Scope)
	if err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	var token app.Token
	client := authz.New(authz.Params{DB: c.db})
	decision, err := client.MutateCustomer(ctx, authz.MutationRequest{
		Permission: permissions.OrgWrite, Target: permissions.Scope{Kind: permissions.KindOrg, OrgID: req.OrgID, ID: req.OrgID},
		SubjectAccountID: req.AccountID,
	}, func(tx *gorm.DB, _ permissions.ResolvedTarget) error {
		ctx := tx.Statement.Context
		credential, _ := cctx.VerifiedCredentialFromContext(ctx)
		var source app.Token
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where(app.Token{ID: credential.TokenID, AccountID: credential.AccountID}).Take(&source).Error; err != nil {
			return err
		}
		if source.TokenType == app.TokenTypeFederated {
			return stderr.ErrAuthorization{Err: fmt.Errorf("federated credentials cannot create API tokens"), Description: "federated credentials cannot create API tokens"}
		}
		if err := authz.RequireUserManaged(tx, req.AccountID); err != nil {
			return err
		}
		var subject app.Account
		if err := tx.Select("id", "account_type", "email").Where(app.Account{ID: req.AccountID}).Take(&subject).Error; err != nil {
			return err
		}
		if subject.AccountType != app.AccountTypeService {
			return stderr.ErrNotFound{Err: fmt.Errorf("service account not found"), Description: "service account not found"}
		}
		authzClient := authz.New(authz.Params{DB: tx})
		grants, err := authzClient.LoadScopedGrants(ctx, subject.ID, req.OrgID)
		if err != nil {
			return stderr.ErrAuthorization{Err: fmt.Errorf("invalid customer token subject: %w", err), Description: "service account has no valid permissions in this org"}
		}
		if grants.API != permissions.APICtl {
			return stderr.ErrAuthorization{Err: fmt.Errorf("invalid customer token subject grants"), Description: "access denied"}
		}
		target, err := authz.New(authz.Params{DB: tx.Clauses(clause.Locking{Strength: "SHARE"})}).ResolveTarget(ctx, scope)
		if err != nil {
			return stderr.ErrNotFound{Err: err, Description: "token scope not found"}
		}
		role, err := authzClient.ResolveDelegableRoleAtScope(ctx, scope, req.Role, app.RoleContextAPIToken)
		if err != nil {
			return err
		}
		policy, err := permissions.DecodePolicy(role.Version, role.Policies[0].Document)
		if err != nil {
			return err
		}
		ceiling, err := policy.Compile(req.OrgID, scope)
		if err != nil {
			return err
		}
		for _, grant := range ceiling {
			if !permissions.ContainsGrant(grants.Grants, target, grant.Permissions) {
				return stderr.ErrAuthorization{Err: fmt.Errorf("token exceeds subject authority"), Description: "token exceeds the account's permissions"}
			}
		}
		now := time.Now()
		expiresAt := now.Add(req.Duration)
		interactive := credential.Profile == cctx.ProfileHuman && (source.TokenType == app.TokenTypeNuon || source.TokenType == app.TokenTypeAuth || source.TokenType == app.TokenTypeAuth0)
		if !interactive && expiresAt.After(source.ExpiresAt) {
			expiresAt = source.ExpiresAt
		}
		if !expiresAt.After(now) {
			return stderr.ErrAuthorization{Err: fmt.Errorf("source credential expired"), Description: "access denied"}
		}
		if req.Invalidate {
			if req.AccountID == credential.AccountID {
				return stderr.NewInvalidRequest(fmt.Errorf("cannot invalidate the credential authorizing issuance"))
			}
			if err := revokeAccountCredentials(tx, req.AccountID); err != nil {
				return err
			}
			if err := tx.Select("id").Where(app.Token{ID: credential.TokenID, AccountID: credential.AccountID}).Take(&app.Token{}).Error; err != nil {
				return stderr.NewInvalidRequest(fmt.Errorf("replacement would revoke its authorizing credential"))
			}
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = subject.Email
		}
		token = app.Token{
			CreatedByID: credential.AccountID, AccountID: subject.ID, OrgID: req.OrgID, Name: name, Role: string(req.Role),
			Token: domains.NewUserTokenID(), TokenType: app.TokenTypeStatic, IssuedAt: now, ExpiresAt: expiresAt, Issuer: credential.AccountID,
			SourceType: app.TokenSourceTypeToken, SourceID: credential.TokenID,
		}
		return PersistToken(tx, &token, permissions.CredentialLimits{Profile: permissions.ProfileServiceAccount, API: permissions.APICtl, Ceiling: ceiling})
	})
	if err != nil {
		return nil, err
	}
	if !decision.Allowed {
		return nil, stderr.ErrAuthorization{Err: fmt.Errorf("token issuance denied: %s", decision.Reason), Description: "access denied"}
	}
	return &token, nil
}

func PersistToken(tx *gorm.DB, token *app.Token, limits permissions.CredentialLimits) error {
	data, err := limits.MarshalJSON()
	if err != nil {
		return err
	}
	token.Authorization = data
	return tx.Create(token).Error
}
