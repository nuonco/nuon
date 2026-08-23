package testseed

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// BuildToken creates an app.Token for the given account with fake defaults. The ID
// is deliberately left empty: Token.BeforeCreate assigns a real one, and the tokens
// table's char_length check constraint rejects anything hand-rolled.
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

// CreateToken builds and persists a token for the account, expiring at expiresAt.
// Pass a past expiry to seed an expired token. For a revoked token, create it and
// then soft-delete it with db.Delete — that exercises the same path production
// revocation does.
func (s *Seeder) CreateToken(ctx context.Context, t *testing.T, acct *app.Account, expiresAt time.Time) *app.Token {
	tok := BuildToken(acct, expiresAt)
	require.NoError(t, s.db.WithContext(ctx).Create(tok).Error)
	return tok
}
