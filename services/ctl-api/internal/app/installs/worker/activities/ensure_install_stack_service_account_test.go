package activities

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func stackTokenTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Written out rather than AutoMigrated: Token carries a Postgres-only
	// char_length check constraint that sqlite cannot create.
	require.NoError(t, db.Exec(`
		CREATE TABLE tokens (
			id TEXT PRIMARY KEY,
			created_by_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0,
			account_id TEXT,
			org_id TEXT,
			name TEXT,
			role TEXT,
			token TEXT,
			token_type TEXT,
			expires_at DATETIME,
			issued_at DATETIME,
			issuer TEXT
		)`).Error)

	return db
}

func insertStackToken(t *testing.T, db *gorm.DB, id, accountID string, expiresAt time.Time, deleted bool) {
	t.Helper()

	var deletedAt int64
	if deleted {
		deletedAt = time.Now().Unix()
	}

	// Created via Exec because Token.BeforeCreate unconditionally overwrites ID.
	require.NoError(t, db.Exec(
		`INSERT INTO tokens (id, account_id, token, expires_at, issued_at, created_at, issuer, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, accountID, "tok-"+id, expiresAt, time.Now(), time.Now(), "nuon", deletedAt,
	).Error)
}

// hasLiveStackToken is the entire mint decision. A false re-mints, so a wrong answer
// either strands a stack without a credential or issues a second one on every
// reconcile — and the customer's Terraform holds whichever came first.
func TestHasLiveStackToken(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	t.Run("no token at all", func(t *testing.T) {
		db := stackTokenTestDB(t)

		live, err := hasLiveStackToken(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.False(t, live, "a stack with no token must mint one")
	})

	t.Run("live token", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-live", "acct-1", future, false)

		live, err := hasLiveStackToken(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.True(t, live, "an unexpired token must not be re-minted")
	})

	t.Run("expired token", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-old", "acct-1", past, false)

		live, err := hasLiveStackToken(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.False(t, live, "an expired token is not a credential")
	})

	t.Run("revoked token", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-gone", "acct-1", future, true)

		live, err := hasLiveStackToken(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.False(t, live, "a soft-deleted token must be filtered out by gorm")
	})

	t.Run("another account's token does not count", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-other", "acct-2", future, false)

		live, err := hasLiveStackToken(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.False(t, live)
	})

	// The partial-failure case that motivated keying on the token rather than on the
	// account: minting failed after the account was created, so the account exists
	// and the token does not. Keying on account existence would skip the retry and
	// strand the stack permanently.
	t.Run("expired and revoked tokens together still re-mint", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-old", "acct-1", past, false)
		insertStackToken(t, db, "tok-gone", "acct-1", future, true)

		live, err := hasLiveStackToken(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.False(t, live)
	})
}
