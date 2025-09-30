package account

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Client) CreateToken(ctx context.Context, subjectOrEmail string, dur time.Duration) (*app.Token, error) {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get account")
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
		return nil, errors.Wrap(res.Error, "unable to create token")
	}

	return &token, nil
}

func (c *Client) InvalidateTokens(ctx context.Context, subjectOrEmail string) error {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return errors.Wrap(err, "unable to get account")
	}

	res := c.db.WithContext(ctx).
		Where(app.Token{
			AccountID: acct.ID,
		}).
		Delete(&app.Token{})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete tokens")
	}

	return nil
}

func (c *Client) ExtendToken(ctx context.Context, subjectOrEmail string, dur time.Duration) error {
	acct, err := c.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return errors.Wrap(err, "unable to get account")
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
		return errors.Wrap(res.Error, "unable to extend token")
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
		return errors.Wrap(res.Error, "unable to update token")
	}

	return nil
}
