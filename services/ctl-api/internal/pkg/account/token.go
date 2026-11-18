package account

import (
	"cmp"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type IssueTokenRequest struct {
	SourceType app.TokenSourceType
	SourceID   string
	OrgID      string

	// TokenType, Issuer and CreatedByID default to a nuon token issued by
	// "nuon" and created by the account itself.
	TokenType   app.TokenType
	Issuer      string
	CreatedByID string

	// Invalidate revokes the account's existing credentials in the same
	// transaction that issues the new token.
	Invalidate bool
}

func (c *Client) CreateToken(ctx context.Context, subjectOrEmail string, dur time.Duration) (*app.Token, error) {
	return c.IssueToken(ctx, subjectOrEmail, dur, IssueTokenRequest{})
}

func (c *Client) IssueToken(ctx context.Context, subjectOrEmail string, dur time.Duration, req IssueTokenRequest) (*app.Token, error) {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to get account")
	}

	now := time.Now()
	token := app.Token{
		CreatedByID: cmp.Or(req.CreatedByID, acct.ID),
		TokenType:   cmp.Or(req.TokenType, app.TokenTypeNuon),
		ExpiresAt:   now.Add(dur),
		IssuedAt:    now,
		Issuer:      cmp.Or(req.Issuer, "nuon"),
		AccountID:   acct.ID,
		OrgID:       req.OrgID,
		SourceType:  req.SourceType,
		SourceID:    req.SourceID,
	}
	if !req.Invalidate {
		return c.insertToken(ctx, &token)
	}
	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.OrgID != "" {
			if err := lockOrg(tx, req.OrgID); err != nil {
				return err
			}
		}
		// Revocation locks the account FOR UPDATE, so take it before issuance's KEY SHARE.
		if err := revokeAccountCredentials(tx, acct.ID); err != nil {
			return err
		}
		_, err := c.withDB(tx).insertToken(ctx, &token)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (c *Client) InvalidateTokens(ctx context.Context, subjectOrEmail string) error {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return pkgerrors.Wrap(err, "unable to get account")
	}

	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return revokeAccountCredentials(tx, acct.ID)
	})
}

// Restores rather than recreates: (deleted_at, email, subject) is unique and
// deleted_at has second precision, so a second soft-deleted row can collide.
func (c *Client) IssueLogStreamToken(ctx context.Context, orgID, logStreamID string, dur time.Duration) (*app.Token, error) {
	if orgID == "" || logStreamID == "" {
		return nil, pkgerrors.New("org and log stream IDs are required")
	}
	var token *app.Token
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		req := ManagedServiceAccountRequest{
			OrgID: orgID, OwnerType: plugins.TableName(tx, app.LogStream{}), OwnerID: logStreamID,
			Purpose: app.ManagedServiceAccountPurposeLogStream, InstanceKey: "default", Subject: logStreamID,
		}
		if err := lockManagedAccountOwner(tx, req); err != nil {
			return err
		}
		client := c.withDB(tx)
		if _, err := client.restoreServiceAccount(ctx, req); err != nil {
			return err
		}
		binding, err := client.EnsureManagedServiceAccount(ctx, req)
		if err != nil {
			return err
		}
		if err := client.authzClient.EnsureLogStreamRole(ctx, orgID, logStreamID, binding.AccountID); err != nil {
			return err
		}
		token, err = client.CreateToken(ctx, binding.AccountID, dur)
		return err
	})
	return token, err
}

func (c *Client) InvalidateOldTokens(ctx context.Context, subjectOrEmail string) (int64, error) {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "unable to get account")
	}

	var latestToken app.Token
	res := c.db.WithContext(ctx).
		Where(app.Token{AccountID: acct.ID}).
		Order("created_at DESC").
		First(&latestToken)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, pkgerrors.Wrap(res.Error, "unable to find latest token")
	}

	return c.revokeTokens(ctx, c.db.Where("account_id = ? AND created_at < ?", acct.ID, latestToken.CreatedAt))
}

func (c *Client) RevokeTokenCreatedBefore(ctx context.Context, subjectOrEmail, tokenID string, cutoff time.Time) (int64, error) {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "unable to get account")
	}

	newestID, err := c.newestTokenID(ctx, acct.ID)
	if err != nil || newestID == "" || newestID == tokenID {
		return 0, err
	}

	return c.revokeTokens(ctx, c.db.
		Where(app.Token{ID: tokenID, AccountID: acct.ID}).
		Where("created_at < ?", cutoff))
}

func (c *Client) RevokeTokensExcept(ctx context.Context, subjectOrEmail string, keepIDs []string, cutoff time.Time) (int64, error) {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "unable to get account")
	}

	newestID, err := c.newestTokenID(ctx, acct.ID)
	if err != nil || newestID == "" {
		return 0, err
	}

	keep := append([]string{newestID}, keepIDs...)
	return c.revokeTokens(ctx, c.db.
		Where(app.Token{AccountID: acct.ID}).
		Where("created_at < ?", cutoff).
		Where("id <> ALL(?)", pq.Array(keep)))
}

func (c *Client) newestTokenID(ctx context.Context, accountID string) (string, error) {
	var newest app.Token
	res := c.db.WithContext(ctx).
		Where(app.Token{AccountID: accountID}).
		Order("created_at DESC").
		First(&newest)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", pkgerrors.Wrap(res.Error, "unable to find newest token")
	}

	return newest.ID, nil
}

func (c *Client) RevokeToken(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return gorm.ErrRecordNotFound
	}
	_, err := c.revokeTokens(ctx, c.db.Where(app.Token{ID: tokenID}))
	return err
}

func (c *Client) DeleteStaticToken(ctx context.Context, orgID, tokenID string) error {
	if tokenID == "" {
		return gorm.ErrRecordNotFound
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockOrg(tx, orgID); err != nil {
			return err
		}
		var token app.Token
		if err := tx.Where(app.Token{ID: tokenID, OrgID: orgID, TokenType: app.TokenTypeStatic}).Take(&token).Error; err != nil {
			return err
		}
		var acct app.Account
		if err := tx.Where(app.Account{ID: token.AccountID}).Take(&acct).Error; err != nil {
			return err
		}
		dedicated := acct.AccountType == app.AccountTypeService && strings.HasPrefix(acct.Subject, DedicatedTokenSubjectPrefix(orgID))
		// Deleting the account locks it before its tokens, so lock it first,
		// in the same order as account-wide revocation.
		if dedicated {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Account{ID: acct.ID}).Take(&acct).Error; err != nil {
				return err
			}
		}
		if err := c.withDB(tx).RevokeToken(ctx, tokenID); err != nil {
			return err
		}
		if dedicated {
			return deleteServiceAccount(tx, acct.ID)
		}
		return nil
	})
}

func (c *Client) revokeTokens(ctx context.Context, where *gorm.DB) (int64, error) {
	var revoked int64
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []string
		if res := tx.Unscoped().Model(&app.Token{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			Order("id").
			Pluck("id", &ids); res.Error != nil {
			return pkgerrors.Wrap(res.Error, "unable to look up tokens")
		}
		if len(ids) == 0 {
			return nil
		}

		res := tx.Where("id = ANY(?)", pq.Array(ids)).Delete(&app.Token{})
		if res.Error != nil {
			return pkgerrors.Wrap(res.Error, "unable to delete tokens")
		}
		revoked = res.RowsAffected

		return RevokeDerivedCredentials(tx, app.TokenSourceTypeToken, ids)
	})
	return revoked, err
}

func (c *Client) RevokeOrgTokens(ctx context.Context, orgID, accountID string) error {
	if orgID == "" || accountID == "" {
		return gorm.ErrRecordNotFound
	}
	// Revoking every org token of an account waits for in-flight issuance on that
	// account, which holds KEY SHARE, so a token committed meanwhile is revoked too.
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: accountID}).Find(&app.Account{}).Error; err != nil {
			return err
		}
		client := c.withDB(tx)
		_, err := client.revokeTokens(ctx, client.db.Where(app.Token{OrgID: orgID, AccountID: accountID}))
		return err
	})
}

func (c *Client) RevokeExpiredTokens(ctx context.Context, accountIDs []string, before time.Time) (int64, error) {
	// Only live tokens: revokeTokens also sees deleted rows, and cleanup runs
	// repeatedly over every runner account.
	return c.revokeTokens(ctx, c.db.Where("account_id IN ? AND expires_at < ? AND deleted_at = 0", accountIDs, before))
}

func (c *Client) EndSession(ctx context.Context, value string) error {
	if value == "" {
		return nil
	}
	return c.db.WithContext(ctx).Where(app.Token{Token: value}).Delete(&app.Token{}).Error
}

func (c *Client) ExtendToken(ctx context.Context, subjectOrEmail string, dur time.Duration) error {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return pkgerrors.Wrap(err, "unable to get account")
	}

	var token app.Token
	res := c.db.WithContext(ctx).
		Where(app.Token{
			AccountID: acct.ID,
		}).
		Order("expires_at desc").
		Limit(1).
		First(&token)
	if res.Error != nil {
		return pkgerrors.Wrap(res.Error, "unable to extend token")
	}

	// update the token expiry
	var updatedToken app.Token
	res = c.db.WithContext(ctx).
		Model(&updatedToken).
		Where(&app.Token{
			ID: token.ID,
		}).
		Updates(app.Token{
			ExpiresAt: token.ExpiresAt.Add(dur),
		})
	if res.Error != nil {
		return pkgerrors.Wrap(res.Error, "unable to update token")
	}

	return nil
}
