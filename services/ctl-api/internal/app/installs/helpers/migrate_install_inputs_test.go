package helpers_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/suite"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	temporal "github.com/nuonco/nuon/pkg/temporal/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migrateInputsDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type MigrateInstallInputsTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps migrateInputsDeps
}

func TestMigrateInstallInputsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(MigrateInstallInputsTestSuite))
}

type migrateInputsWorkflowRun struct{}

func (r *migrateInputsWorkflowRun) GetID() string    { return "test-workflow-id" }
func (r *migrateInputsWorkflowRun) GetRunID() string { return "test-run-id" }
func (r *migrateInputsWorkflowRun) Get(context.Context, any) error {
	return nil
}
func (r *migrateInputsWorkflowRun) GetWithOptions(context.Context, any, tclient.WorkflowRunGetOptions) error {
	return nil
}

func (s *MigrateInstallInputsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	ctrl := gomock.NewController(s.T())
	mockTC := temporal.NewMockClient(ctrl)
	mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&migrateInputsWorkflowRun{}, nil).AnyTimes()

	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T:     s.T(),
		Mocks: &tests.TestMocks{MockTC: mockTC},
	}), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *MigrateInstallInputsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// seedAppConfigWithInputs creates a further app config carrying an input config
// whose single input is named "region" (what the seeder builds).
func (s *MigrateInstallInputsTestSuite) seedAppConfigWithInputs(ctx context.Context, appID string) *app.AppConfig {
	cfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), appID)
	s.deps.Seed.CreateAppInputConfig(ctx, s.T(), appID, cfg.ID)
	return cfg
}

func (s *MigrateInstallInputsTestSuite) latestInstallInputs(ctx context.Context, installID string) app.InstallInputs {
	var got app.InstallInputs
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.InstallInputs{InstallID: installID}).
		Order("created_at DESC").
		First(&got).Error)
	return got
}

func ptrTo[T any](v T) *T { return &v }

// The migration must carry values onto the new config. Dropping them leaves
// templates like {{.nuon.inputs.inputs.region}} resolving to nil at plan time.
func (s *MigrateInstallInputsTestSuite) TestMigratesValuesTiedToOutgoingConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID,
		map[string]*string{"region": ptrTo("us-west-2")})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Require().NotNil(got.Values["region"])
	s.Equal("us-west-2", *got.Values["region"])
}

// Regression for #857: the lookup uses Find, which reports "no rows" through
// RowsAffected rather than gorm.ErrRecordNotFound. An install whose inputs are
// tied to some other config version took the miss path and had its values
// silently replaced with an empty set, which then persisted through every later
// migration.
func (s *MigrateInstallInputsTestSuite) TestMigratesValuesNotTiedToOutgoingConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	// Values live against an unrelated input config, so the primary lookup misses.
	strayCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	var strayInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: strayCfg.ID}).First(&strayInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, strayInputCfg.ID,
		map[string]*string{"region": ptrTo("eu-central-1")})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Require().NotNil(got.Values["region"], "fallback must carry the install's latest values forward")
	s.Equal("eu-central-1", *got.Values["region"])
}

// Inputs dropped from the new config must not be carried forward.
func (s *MigrateInstallInputsTestSuite) TestDropsValuesRemovedFromNewConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID, map[string]*string{
		"region": ptrTo("us-west-2"),
		"gone":   ptrTo("stale"),
	})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Equal("us-west-2", *got.Values["region"])
	s.NotContains(got.Values, "gone")
}

// The install's inputs must pin to the input config of the app config the
// install itself is pinned to. Pinning to the app's newest input config instead
// lets the two diverge as soon as a newer app config exists, which is what makes
// the migration lookup miss.
func (s *MigrateInstallInputsTestSuite) TestCreateInstallPinsInputsToItsOwnAppConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	activeCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)

	// A newer, non-active config with its own input config: the app's "latest"
	// input config now belongs to a config no install is pinned to.
	newerCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	// The active filter reads status_v2, so both columns have to move.
	pending := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusPending))
	s.Require().NoError(s.deps.DB.WithContext(ctx).Model(&app.AppConfig{}).
		Where("id = ?", newerCfg.ID).
		Updates(map[string]any{
			"status":    app.AppConfigStatusPending,
			"status_v2": pending,
		}).Error)
	s.deps.Seed.CreateAppInputConfig(ctx, s.T(), testApp.ID, newerCfg.ID)

	install, err := s.deps.Helpers.CreateInstall(ctx, testApp.ID, &installhelpers.CreateInstallParams{
		Name:        "pin-check",
		SandboxMode: true,
		AWSAccount:  &installhelpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
		Inputs:      map[string]*string{"region": ptrTo("us-west-2")},
	})
	s.Require().NoError(err)

	var pinnedInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: install.AppConfigID}).First(&pinnedInputCfg).Error)

	got := s.latestInstallInputs(ctx, install.ID)
	s.Equal(pinnedInputCfg.ID, got.AppInputConfigID,
		"inputs must pin to the install's own app config, not the app's newest")
	s.Require().NotNil(got.Values["region"])
	s.Equal("us-west-2", *got.Values["region"])
	s.Equal(activeCfg.ID, install.AppConfigID)
}

// Aug 4 incident: a read keyed on the outgoing config cannot see an update already
// pinned to the incoming one, so it copied a stale snapshot forward.
func (s *MigrateInstallInputsTestSuite) TestMigrationDoesNotRevertUpdateOnNewConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID,
		map[string]*string{"region": ptrTo("us-west-2")})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	var newInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: newCfg.ID}).First(&newInputCfg).Error)

	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, newInputCfg.ID, map[string]*string{
		"region": ptrTo("us-west-2"),
		"gate":   ptrTo("true"),
	})

	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Require().NotNil(got.Values["gate"], "migration must not revert an update already on the new config")
	s.Equal("true", *got.Values["gate"])
	s.Equal("us-west-2", *got.Values["region"])
}

func (s *MigrateInstallInputsTestSuite) TestMigrationBeforeUpdateStillCarriesValues() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID,
		map[string]*string{"region": ptrTo("eu-west-1")})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)
	s.Require().NoError(s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
		map[string]string{install.ID: oldCfg.ID}, newCfg.ID))

	got := s.latestInstallInputs(ctx, install.ID)
	s.Require().NotNil(got.Values["region"])
	s.Equal("eu-west-1", *got.Values["region"])
}

// Without the lock the migration's append discards an update that landed after its read.
func (s *MigrateInstallInputsTestSuite) TestConcurrentUpdateIsNotLost() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var oldInputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: oldCfg.ID}).First(&oldInputCfg).Error)
	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, oldInputCfg.ID,
		map[string]*string{"region": ptrTo("us-west-2"), "legacy": ptrTo("drop-me")})

	newCfg := s.seedAppConfigWithInputs(ctx, testApp.ID)

	held := make(chan struct{})
	release := make(chan struct{})
	updateDone := make(chan error, 1)
	migrateDone := make(chan error, 1)

	go func() {
		updateDone <- s.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := installhelpers.LockInstallInputs(ctx, tx, install.ID); err != nil {
				return err
			}
			close(held)
			<-release

			return tx.WithContext(ctx).Create(&app.InstallInputs{
				InstallID:        install.ID,
				AppInputConfigID: oldInputCfg.ID,
				Values: pgtype.Hstore{
					"region": ptrTo("eu-central-1"),
					"legacy": ptrTo("drop-me"),
				},
			}).Error
		})
	}()

	<-held
	go func() {
		migrateDone <- s.deps.Helpers.MigrateInstallInputsToNewConfig(ctx, s.deps.DB,
			map[string]string{install.ID: oldCfg.ID}, newCfg.ID)
	}()

	select {
	case err := <-migrateDone:
		s.Require().Failf("migration ran unserialized", "finished while an inputs writer held the lock: %v", err)
	case <-time.After(time.Second):
	}

	close(release)
	s.Require().NoError(<-updateDone)
	s.Require().NoError(<-migrateDone)

	got := s.latestInstallInputs(ctx, install.ID)
	s.Require().NotNil(got.Values["region"])
	s.Equal("eu-central-1", *got.Values["region"])
	s.NotContains(got.Values, "legacy")
}
