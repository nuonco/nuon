package migrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type migration129TestSuite struct {
	tests.BaseDBTestSuite

	db *gorm.DB
}

func TestMigration129Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(migration129TestSuite))
}

// SetupSuite runs against a database of its own. This suite rebuilds pre-migration schema, which
// means dropping indexes and rewriting whole tables - safe only when nothing else is using them,
// and `go test ./...` runs package binaries concurrently against the shared test database.
func (s *migration129TestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	cfg, err := tests.LoadDBConfig()
	require.NoError(s.T(), err)
	cfg.DBName = cfg.DBName + "_mig129"
	require.NoError(s.T(), tests.CreateAndMigrateDatabase(cfg))

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)
	s.SetDB(s.db)
}

func (s *migration129TestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

// restoreLegacySchema puts the two tables back the way they were before this migration so the
// migration can be exercised against the state it will actually meet in a long-lived database.
//
// Existing rows have to go first: the legacy indexes forbid exactly the rows the post-migration
// schema allows, so recreating them over rows left by an earlier test fails on a duplicate key.
func (s *migration129TestSuite) restoreLegacySchema(ctx context.Context) {
	require.NoError(s.T(), s.db.WithContext(ctx).Exec(`
		DELETE FROM account_identities;
		DELETE FROM identity_providers;
		ALTER TABLE account_identities ALTER COLUMN identity_provider_id DROP NOT NULL;
		DROP INDEX IF EXISTS idx_account_identity_account_idp;
		DROP INDEX IF EXISTS idx_account_identity_idp_sub;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_account_identity_account_provider
			ON account_identities (account_id, provider_type);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_account_identity_provider_sub
			ON account_identities (provider_type, sub);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_type
			ON identity_providers (deleted_at, org_id);
	`).Error)
}

func (s *migration129TestSuite) seedAccount(ctx context.Context) *app.Account {
	acct := &app.Account{
		ID:          domains.NewAccountID(),
		Subject:     domains.NewAccountID(),
		Email:       fmt.Sprintf("%s@test.nuon.co", domains.NewAccountID()),
		AccountType: app.AccountTypeAuth,
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(acct).Error)
	return acct
}

func (s *migration129TestSuite) seedLegacyIdentity(
	ctx context.Context,
	accountID string,
	providerType app.ProviderType,
	sub string,
	identityProviderID *string,
) string {
	id := domains.NewAccountIdentityID()
	require.NoError(s.T(), s.db.WithContext(ctx).Exec(`
		INSERT INTO account_identities (id, created_at, updated_at, account_id, identity_provider_id, provider_type, sub)
		VALUES (?, now(), now(), ?, ?, ?, ?)
	`, id, accountID, identityProviderID, providerType, sub).Error)
	return id
}

func (s *migration129TestSuite) identityProviderID(ctx context.Context, id string) string {
	var out string
	require.NoError(s.T(), s.db.WithContext(ctx).
		Raw(`SELECT identity_provider_id FROM account_identities WHERE id = ?`, id).
		Scan(&out).Error)
	return out
}

func (s *migration129TestSuite) indexExists(ctx context.Context, name string) bool {
	var count int64
	require.NoError(s.T(), s.db.WithContext(ctx).
		Raw(`SELECT count(*) FROM pg_indexes WHERE indexname = ?`, name).
		Scan(&count).Error)
	return count > 0
}

func (s *migration129TestSuite) TestBackfillsEnvIdentitiesAndDropsLegacyIndexes() {
	ctx := context.Background()
	s.restoreLegacySchema(ctx)

	dbProvider := &app.IdentityProvider{
		ID:           domains.NewIdentityProviderID(),
		ProviderType: app.ProviderTypeGoogle,
		Enabled:      true,
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(dbProvider).Error)

	acct := s.seedAccount(ctx)
	envIdentity := s.seedLegacyIdentity(ctx, acct.ID, app.ProviderTypeOIDC, "auth0|"+acct.ID, nil)
	dbIdentity := s.seedLegacyIdentity(ctx, acct.ID, app.ProviderTypeGoogle, "google|"+acct.ID, &dbProvider.ID)

	migrations := psqlmigrations.New(psqlmigrations.Params{L: zap.NewNop()})
	require.NoError(s.T(), migrations.Migration129IdentityProvidersMultiPerType(ctx, s.db))

	require.Equal(s.T(), app.EnvIdentityProviderID(app.ProviderTypeOIDC), s.identityProviderID(ctx, envIdentity),
		"env-provider identities must land on the sentinel the auth service looks them up by")
	require.Equal(s.T(), dbProvider.ID, s.identityProviderID(ctx, dbIdentity))

	require.False(s.T(), s.indexExists(ctx, "idx_account_identity_account_provider"))
	require.False(s.T(), s.indexExists(ctx, "idx_account_identity_provider_sub"))
	require.False(s.T(), s.indexExists(ctx, "idx_provider_type"))
}

// The sentinel is namespaced by provider type precisely so the backfill cannot collide on the
// unique indexes that replace (account_id, provider_type) and (provider_type, sub).
func (s *migration129TestSuite) TestBackfillSurvivesAnAccountWithTwoEnvIdentities() {
	ctx := context.Background()
	s.restoreLegacySchema(ctx)

	acct := s.seedAccount(ctx)
	oidcIdentity := s.seedLegacyIdentity(ctx, acct.ID, app.ProviderTypeOIDC, "shared-sub", nil)
	githubIdentity := s.seedLegacyIdentity(ctx, acct.ID, app.ProviderTypeGitHub, "shared-sub", nil)

	migrations := psqlmigrations.New(psqlmigrations.Params{L: zap.NewNop()})
	require.NoError(s.T(), migrations.Migration129IdentityProvidersMultiPerType(ctx, s.db))

	require.Equal(s.T(), app.EnvIdentityProviderID(app.ProviderTypeOIDC), s.identityProviderID(ctx, oidcIdentity))
	require.Equal(s.T(), app.EnvIdentityProviderID(app.ProviderTypeGitHub), s.identityProviderID(ctx, githubIdentity))
}

func (s *migration129TestSuite) TestAllowsMultipleProvidersOfTheSameType() {
	ctx := context.Background()
	s.restoreLegacySchema(ctx)

	migrations := psqlmigrations.New(psqlmigrations.Params{L: zap.NewNop()})
	require.NoError(s.T(), migrations.Migration129IdentityProvidersMultiPerType(ctx, s.db))
	// idempotent: a redeploy re-runs nothing but must not fail if it does
	require.NoError(s.T(), migrations.Migration129IdentityProvidersMultiPerType(ctx, s.db))

	for range 2 {
		require.NoError(s.T(), s.db.WithContext(ctx).Create(&app.IdentityProvider{
			ID:           domains.NewIdentityProviderID(),
			ProviderType: app.ProviderTypeOIDC,
			Enabled:      true,
		}).Error)
	}
}
