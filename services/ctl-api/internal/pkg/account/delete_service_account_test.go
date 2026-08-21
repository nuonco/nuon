package account

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func deleteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Written out rather than AutoMigrated: these models carry Postgres-only
	// char_length check constraints that sqlite cannot create.
	for _, ddl := range []string{
		`CREATE TABLE accounts (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0,
			email TEXT,
			subject TEXT,
			name TEXT,
			account_type TEXT,
			user_journeys TEXT
		)`,
		`CREATE TABLE account_roles (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at INTEGER DEFAULT 0,
			org_id TEXT,
			role_id TEXT,
			account_id TEXT
		)`,
		`CREATE TABLE tokens (
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
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}

	return db
}

func seedAccount(t *testing.T, db *gorm.DB, id, subject string, acctType app.AccountType) {
	t.Helper()

	require.NoError(t, db.Exec(
		`INSERT INTO accounts (id, email, subject, account_type, created_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, 0)`,
		id, ServiceAccountEmail(subject), subject, acctType, time.Now(),
	).Error)
}

func seedRoleBinding(t *testing.T, db *gorm.DB, id, accountID string) {
	t.Helper()

	require.NoError(t, db.Exec(
		`INSERT INTO account_roles (id, account_id, role_id, created_at, deleted_at)
		 VALUES (?, ?, ?, ?, 0)`,
		id, accountID, "role-org-admin", time.Now(),
	).Error)
}

func seedToken(t *testing.T, db *gorm.DB, id, accountID string) {
	t.Helper()

	require.NoError(t, db.Exec(
		`INSERT INTO tokens (id, account_id, token, expires_at, issued_at, created_at, issuer, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
		id, accountID, "tok-"+id, time.Now().Add(time.Hour), time.Now(), time.Now(), "nuon",
	).Error)
}

// A leftover role binding or token is the whole reason this exists: the account is
// tied to its stack only by a naming convention, so nothing else will ever collect
// it, and what survives is an org-admin credential with no owner.
func TestDeleteAccountRecords(t *testing.T) {
	db := deleteTestDB(t)

	seedAccount(t, db, "acct-1", "stack-1", app.AccountTypeService)
	seedRoleBinding(t, db, "ar-1", "acct-1")
	seedRoleBinding(t, db, "ar-2", "acct-1")
	seedToken(t, db, "tok-1", "acct-1")

	// A second account proves the deletes are scoped rather than table-wide.
	seedAccount(t, db, "acct-2", "stack-2", app.AccountTypeService)
	seedRoleBinding(t, db, "ar-3", "acct-2")
	seedToken(t, db, "tok-2", "acct-2")

	require.NoError(t, deleteAccountRecords(db, "acct-1"))

	// Role bindings are hard-deleted: the many2many declares OnDelete:CASCADE, which
	// is a foreign-key constraint that a soft delete would never fire.
	var roleRows int64
	require.NoError(t, db.Unscoped().Model(&app.AccountRole{}).
		Where("account_id = ?", "acct-1").Count(&roleRows).Error)
	assert.Zero(t, roleRows, "role bindings must be hard-deleted, not soft-deleted")

	// Tokens and the account are soft-deleted, which is enough: the auth middleware
	// resolves a token through FindAccount, and that cannot see a soft-deleted row.
	var tokens []app.Token
	require.NoError(t, db.Where("account_id = ?", "acct-1").Find(&tokens).Error)
	assert.Empty(t, tokens, "tokens must be invisible to a normal query")

	var acct app.Account
	err := db.Where("id = ?", "acct-1").First(&acct).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "the account must be invisible to a normal query")

	var stillThere int64
	require.NoError(t, db.Unscoped().Model(&app.AccountRole{}).
		Where("account_id = ?", "acct-2").Count(&stillThere).Error)
	assert.EqualValues(t, 1, stillThere, "another account's bindings must survive")

	var otherTokens []app.Token
	require.NoError(t, db.Where("account_id = ?", "acct-2").Find(&otherTokens).Error)
	assert.Len(t, otherTokens, 1, "another account's tokens must survive")
}

// Delete workflows retry, and a stack that never had a service account must not be
// able to block its own install from being torn down.
func TestDeleteServiceAccountIsIdempotent(t *testing.T) {
	db := deleteTestDB(t)
	c := &Client{db: db}

	require.NoError(t, c.DeleteServiceAccount(t.Context(), "stack-never-existed"))

	seedAccount(t, db, "acct-1", "stack-1", app.AccountTypeService)
	seedToken(t, db, "tok-1", "acct-1")

	require.NoError(t, c.DeleteServiceAccount(t.Context(), "stack-1"))
	require.NoError(t, c.DeleteServiceAccount(t.Context(), "stack-1"),
		"a second delete must be a no-op, not an error")
}

// FindAccount matches on email, subject, or ID, so a caller passing something
// unexpected could otherwise reach a real user through a path meant for machine
// identities.
func TestDeleteServiceAccountRejectsNonServiceAccount(t *testing.T) {
	db := deleteTestDB(t)
	c := &Client{db: db}

	seedAccount(t, db, "acct-human", "stack-1", app.AccountTypeAuth0)

	err := c.DeleteServiceAccount(t.Context(), "stack-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a service account")

	var acct app.Account
	require.NoError(t, db.Where("id = ?", "acct-human").First(&acct).Error,
		"the account must survive a rejected delete")
}
