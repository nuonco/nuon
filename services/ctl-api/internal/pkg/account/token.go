package account

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"
	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (c *Client) CreateToken(ctx context.Context, subjectOrEmail string, dur time.Duration) (*app.Token, error) {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to get account")
	}

	token := app.Token{
		CreatedByID: acct.ID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   time.Now().Add(dur),
		IssuedAt:    time.Now(),
		Issuer:      "nuon",
		AccountID:   acct.ID,
	}

	if err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var liveAccount struct{ ID string }
		if err := tx.Model(&app.Account{}).Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: acct.ID}).Take(&liveAccount).Error; err != nil {
			return pkgerrors.Wrap(err, "unable to lock account for token issuance")
		}
		return tx.Create(&token).Error
	}); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to create token")
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

func (c *Client) RevokeToken(ctx context.Context, tokenID string) error {
	_, err := c.revokeTokens(ctx, c.db.Where(app.Token{ID: tokenID}))
	return err
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
