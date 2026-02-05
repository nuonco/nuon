package account

import (
	"context"
	"errors"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"

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

	res := c.db.WithContext(ctx).
		Create(&token)
	if res.Error != nil {
		return nil, pkgerrors.Wrap(res.Error, "unable to create token")
	}

	return &token, nil
}

func (c *Client) InvalidateTokens(ctx context.Context, subjectOrEmail string) error {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return pkgerrors.Wrap(err, "unable to get account")
	}

	res := c.db.WithContext(ctx).
		Where(app.Token{
			AccountID: acct.ID,
		}).
		Delete(&app.Token{})
	if res.Error != nil {
		return pkgerrors.Wrap(res.Error, "unable to delete tokens")
	}

	return nil
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

	res = c.db.WithContext(ctx).
		Where("account_id = ? AND created_at < ?", acct.ID, latestToken.CreatedAt).
		Delete(&app.Token{})
	if res.Error != nil {
		return 0, pkgerrors.Wrap(res.Error, "unable to delete old tokens")
	}

	return res.RowsAffected, nil
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
