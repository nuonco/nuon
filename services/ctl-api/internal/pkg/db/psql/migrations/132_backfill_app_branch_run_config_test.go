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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migration132TestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestMigration132Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(migration132TestSuite))
}

func (s *migration132TestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	cfg, err := tests.LoadDBConfig()
	require.NoError(s.T(), err)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)
	s.seeder = testseed.New(testseed.Params{DB: s.db})
	s.SetDB(s.db)
}

func (s *migration132TestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *migration132TestSuite) seedBranch(ctx context.Context, name string) *app.AppBranch {
	testApp := s.seeder.CreateApp(ctx, s.T())
	branch := &app.AppBranch{AppID: testApp.ID, Name: name}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(branch).Error)
	return branch
}

func (s *migration132TestSuite) seedConfig(ctx context.Context, branchID string, runConfig *app.AppBranchRunConfig) *app.AppBranchConfig {
	cfg := &app.AppBranchConfig{AppBranchID: branchID, RunConfig: runConfig}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(cfg).Error)
	return cfg
}

func (s *migration132TestSuite) runConfigJSON(ctx context.Context, id string) string {
	var raw *string
	require.NoError(s.T(), s.db.WithContext(ctx).Raw(
		`SELECT run_config::text FROM app_branch_configs WHERE id = ?`, id,
	).Scan(&raw).Error)
	if raw == nil {
		return ""
	}
	return *raw
}

func (s *migration132TestSuite) TestBackfillsNullAndPreservesExplicitConfigs() {
	ctx, _ := s.seeder.EnsureAccount(context.Background(), s.T())
	ctx, _ = s.seeder.EnsureOrg(ctx, s.T())

	nullBranch := s.seedBranch(ctx, "null-run")
	sqlNull := s.seedConfig(ctx, nullBranch.ID, nil)
	require.NoError(s.T(), s.db.WithContext(ctx).Exec(
		`UPDATE app_branch_configs SET run_config = NULL WHERE id = ?`, sqlNull.ID,
	).Error)
	jsonNull := s.seedConfig(ctx, nullBranch.ID, nil)
	require.NoError(s.T(), s.db.WithContext(ctx).Exec(
		`UPDATE app_branch_configs SET run_config = 'null'::jsonb WHERE id = ?`, jsonNull.ID,
	).Error)
	historical := s.seedConfig(ctx, nullBranch.ID, nil)
	require.NoError(s.T(), s.db.WithContext(ctx).Exec(
		`UPDATE app_branch_configs SET run_config = NULL, deleted_at = 1 WHERE id = ?`, historical.ID,
	).Error)

	explicitBranch := s.seedBranch(ctx, "explicit-run")
	alreadyPush := s.seedConfig(ctx, explicitBranch.ID, &app.AppBranchRunConfig{Mode: app.AppBranchRunModePush})
	tagPrefix := s.seedConfig(ctx, explicitBranch.ID, &app.AppBranchRunConfig{
		Mode:      app.AppBranchRunModeTagPrefix,
		TagPrefix: "release/",
	})
	label := s.seedConfig(ctx, explicitBranch.ID, &app.AppBranchRunConfig{
		Mode:        app.AppBranchRunModeGithubLabel,
		GithubLabel: "deploy-cadence-daily",
	})
	manual := s.seedConfig(ctx, explicitBranch.ID, &app.AppBranchRunConfig{Mode: app.AppBranchRunModeManualOnly})

	migrations := psqlmigrations.New(psqlmigrations.Params{L: zap.NewNop()})
	require.NoError(s.T(), migrations.Migration132BackfillAppBranchRunConfig(ctx, s.db))
	require.NoError(s.T(), migrations.Migration132BackfillAppBranchRunConfig(ctx, s.db))

	require.JSONEq(s.T(), `{"mode":"push"}`, s.runConfigJSON(ctx, sqlNull.ID))
	require.JSONEq(s.T(), `{"mode":"push"}`, s.runConfigJSON(ctx, jsonNull.ID))
	require.JSONEq(s.T(), `{"mode":"push"}`, s.runConfigJSON(ctx, historical.ID))

	require.JSONEq(s.T(), `{"mode":"push"}`, s.runConfigJSON(ctx, alreadyPush.ID))
	require.JSONEq(s.T(), `{"mode":"on_tag_prefix","tag_prefix":"release/"}`, s.runConfigJSON(ctx, tagPrefix.ID))
	require.JSONEq(s.T(), `{"mode":"on_github_label","github_label":"deploy-cadence-daily"}`, s.runConfigJSON(ctx, label.ID))
	require.JSONEq(s.T(), `{"mode":"manual_only"}`, s.runConfigJSON(ctx, manual.ID))
}
