package service

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

// liveStackTokenExpiry drives what the TF Module tab tells the customer: whether they
// still have a working credential, and when it dies. A wrong answer here either tells
// someone their stack is authenticated when it is not, or prompts them to mint a
// duplicate token they did not need.
func TestLiveStackTokenExpiry(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	farFuture := time.Now().Add(365 * 24 * time.Hour).UTC().Truncate(time.Second)
	past := time.Now().Add(-24 * time.Hour)

	t.Run("no token at all", func(t *testing.T) {
		db := stackTokenTestDB(t)

		expiry, err := liveStackTokenExpiry(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.True(t, expiry.IsZero(), "no token means no expiry to report")
	})

	t.Run("live token reports its expiry", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-live", "acct-1", future, false)

		expiry, err := liveStackTokenExpiry(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.WithinDuration(t, future, expiry, time.Second)
	})

	t.Run("expired token", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-old", "acct-1", past, false)

		expiry, err := liveStackTokenExpiry(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.True(t, expiry.IsZero(), "an expired token is not a credential")
	})

	t.Run("revoked token", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-gone", "acct-1", future, true)

		expiry, err := liveStackTokenExpiry(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.True(t, expiry.IsZero(), "a soft-deleted token must be filtered out by gorm")
	})

	t.Run("another account's token does not count", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-other", "acct-2", future, false)

		expiry, err := liveStackTokenExpiry(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.True(t, expiry.IsZero())
	})

	// The reason this orders by expiry and not by created_at. The create modal lets
	// the caller pick a duration, so the newest token can be the shortest-lived one;
	// reporting its expiry would warn a customer holding a year-long credential that
	// they expire tomorrow.
	t.Run("reports the longest-lived token, not the newest", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-year", "acct-1", farFuture, false)
		insertStackToken(t, db, "tok-day", "acct-1", future, false)

		expiry, err := liveStackTokenExpiry(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.WithinDuration(t, farFuture, expiry, time.Second)
	})

	t.Run("expired and revoked tokens together report nothing live", func(t *testing.T) {
		db := stackTokenTestDB(t)
		insertStackToken(t, db, "tok-old", "acct-1", past, false)
		insertStackToken(t, db, "tok-gone", "acct-1", future, true)

		expiry, err := liveStackTokenExpiry(ctx, db, "acct-1")
		require.NoError(t, err)
		assert.True(t, expiry.IsZero())
	})
}
