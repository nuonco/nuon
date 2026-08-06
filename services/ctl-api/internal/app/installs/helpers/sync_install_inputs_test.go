package helpers_test

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installsyncer "github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/installs"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// A config sync declares only the inputs it knows about. Replacing the snapshot with
// just those drops everything set out of band — a dashboard update, or customer stack
// outputs — and leaving the row unpinned hides it from the config migration.
func (s *StackOutputInputsTestSuite) TestConfigSyncMergesAndPinsInputs() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	cfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)

	var inputCfg app.AppInputConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppInputConfig{AppConfigID: cfg.ID}).First(&inputCfg).Error)

	s.deps.Seed.CreateInstallInputs(ctx, s.T(), install.ID, inputCfg.ID, map[string]*string{
		"region":       ptrTo("us-west-2"),
		"set_by_hand":  ptrTo("keep-me"),
		"from_a_stack": ptrTo("also-keep-me"),
	})

	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.InstallState{}).
		Create(&app.InstallState{InstallID: install.ID}).Error)

	installCfg := &config.Install{
		Name:           install.Name,
		ApprovalOption: config.InstallApprovalOptionApproveAll,
		InputGroups: []config.InputGroup{
			{Inputs: map[string]string{"region": "eu-central-1"}},
		},
	}

	res, err := installsyncer.SyncInstall(ctx, s.deps.DB, s.deps.Helpers, testApp.ID, installCfg)
	s.Require().NoError(err)
	s.Require().True(res.Changed, "config declared a different region, so the sync must apply")

	var newest app.InstallInputs
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.InstallInputs{InstallID: install.ID}).
		Order("created_at DESC").
		First(&newest).Error)

	s.Equal(inputCfg.ID, newest.AppInputConfigID, "synced inputs must be pinned to the install's config")

	s.Require().NotNil(newest.Values["region"])
	s.Equal("eu-central-1", *newest.Values["region"], "config-declared value must win")
	s.Require().NotNil(newest.Values["set_by_hand"], "out-of-band input was dropped")
	s.Equal("keep-me", *newest.Values["set_by_hand"])
	s.Require().NotNil(newest.Values["from_a_stack"], "customer stack output was dropped")

	var st app.InstallState
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.InstallState{InstallID: install.ID}).
		Order("created_at DESC").
		First(&st).Error)
	s.Contains(st.StalePartials, pkgstate.PartialInputs, "state was not invalidated, so the install keeps the old inputs")
}
