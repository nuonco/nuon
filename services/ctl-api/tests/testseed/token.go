package testseed

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func BuildToken(acct *app.Account, expiresAt time.Time) *app.Token {
	return &app.Token{
		CreatedByID: acct.ID,
		AccountID:   acct.ID,
		Token:       generics.GetFakeObj[string](),
		TokenType:   app.TokenTypeStatic,
		ExpiresAt:   expiresAt,
		IssuedAt:    time.Now(),
		Issuer:      "testseed",
	}
}

func (s *Seeder) CreateToken(ctx context.Context, t *testing.T, acct *app.Account, expiresAt time.Time) *app.Token {
	tok := BuildToken(acct, expiresAt)
	require.NoError(t, s.db.WithContext(ctx).Create(tok).Error)
	return tok
}
