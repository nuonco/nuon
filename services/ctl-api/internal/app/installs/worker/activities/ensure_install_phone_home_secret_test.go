package activities

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func phoneHomeTestDB(t *testing.T) *gorm.DB {
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

func insertToken(t *testing.T, db *gorm.DB, id, value string, expiresAt time.Time) {
	t.Helper()

	// Created via Exec because Token.BeforeCreate unconditionally overwrites ID.
	require.NoError(t, db.Exec(
		`INSERT INTO tokens (id, token, expires_at, issued_at, issuer, deleted_at) VALUES (?, ?, ?, ?, ?, 0)`,
		id, value, expiresAt, time.Now(), "nuon",
	).Error)
}

// livePhoneHomeTokens decides whether a stack version's recorded token still backs a
// real credential. Everything it reports as missing gets re-minted, which is what
// makes the reconciler self-heal after a partial failure — so each way a token can
// stop being live needs to be covered.
func TestLivePhoneHomeTokens(t *testing.T) {
	db := phoneHomeTestDB(t)
	acts := &Activities{db: db}

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	insertToken(t, db, "tok00000000000000000000live", "live-value", future)
	insertToken(t, db, "tok0000000000000000000expd", "expired-value", past)
	insertToken(t, db, "tok0000000000000000000soft", "soft-deleted", future)
	require.NoError(t, db.Exec(`UPDATE tokens SET deleted_at = ? WHERE id = ?`,
		time.Now().Unix(), "tok0000000000000000000soft").Error)

	versions := []app.InstallStackVersion{
		{ID: "isv1", PhoneHomeTokenID: "tok00000000000000000000live"},
		{ID: "isv2", PhoneHomeTokenID: "tok0000000000000000000expd"},
		{ID: "isv3", PhoneHomeTokenID: "tok0000000000000000000soft"},
		{ID: "isv4", PhoneHomeTokenID: "tok00000000000000000gone00"},
		{ID: "isv5", PhoneHomeTokenID: ""},
	}

	live, err := acts.livePhoneHomeTokens(context.Background(), versions)
	require.NoError(t, err)

	assert.Equal(t, "live-value", live["tok00000000000000000000live"])

	// An expired token is a dead credential: the auth middleware checks ExpiresAt
	// unconditionally, so leaving it in the map would strand the stack.
	assert.NotContains(t, live, "tok0000000000000000000expd", "an expired token must not count as live")
	assert.NotContains(t, live, "tok0000000000000000000soft", "a soft-deleted token must not count as live")
	assert.NotContains(t, live, "tok00000000000000000gone00", "a missing row must not count as live")
	assert.Len(t, live, 1)
}

func TestLivePhoneHomeTokensNoTokens(t *testing.T) {
	acts := &Activities{db: phoneHomeTestDB(t)}

	live, err := acts.livePhoneHomeTokens(context.Background(), []app.InstallStackVersion{
		{ID: "isv1"},
		{ID: "isv2"},
	})
	require.NoError(t, err)
	assert.Empty(t, live)
}

// The 10-year expiry is deliberate: a deployed stack can be updated years after it
// was applied, and the ordinary 90-day runner token expiry would silently strand it.
func TestPhoneHomeTokenTimeoutIsLongLived(t *testing.T) {
	assert.Greater(t, phoneHomeTokenTimeout, 9*365*24*time.Hour,
		"a phone-home token must outlive any plausible deployed stack")
}
