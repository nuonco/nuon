package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/config/validate"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	defaultSyncTimeout        time.Duration = time.Minute * 20
	defaultSyncSleep          time.Duration = time.Second * 20
	componentBuildStatusError               = "error"

	componentBuildStatusBuilding = "building"
	componentBuildStatusActive   = "active"
	componentStatusQueued        = "queued"
)

func (s *Service) DeprecatedSyncDir(ctx context.Context, dir string, version string) error {
	deprecatedWarning := config.ErrConfig{
		Description: "nuon apps sync-dir is deprecated, please use nuon apps sync instead",
		Warning:     true,
		Err:         fmt.Errorf("deprecated command nuon sync-dir"),
	}
	ui.PrintError(deprecatedWarning)
	return s.SyncDir(ctx, dir, version)
}

func (s *Service) SyncDir(ctx context.Context, dir string, version string) error {
	ui.PrintLn("syncing directory from " + dir)
	appName, err := parse.AppNameFromDirName(dir)
	if err != nil {
		err = errs.WithUserFacing(err, "error parsing app name from file")
		return ui.PrintError(err)
	}

	appID, err := lookup.AppID(ctx, s.api, appName)
	if err != nil {
		err = errs.WithUserFacing(err, "error looking up app id")
		return ui.PrintError(err)
	}

	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname:       dir,
		V:             validator.New(),
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return ui.PrintError(err)
	}

	if s.cfg.Debug {
		ui.PrintJSON(cfg)
	}

	ui.PrintLn("validating configs")
	err = validate.Validate(ctx, s.v, cfg)
	if err != nil {
		if config.IsWarningErr(err) {
			ui.PrintError(err)
		} else {
			return ui.PrintError(err)
		}
	}
	ui.PrintLn("all configs valid")

	// TODO(onprem): remove this after a few releases
	if len(cfg.Installs) > 0 {
		ui.PrintWarning("deprecated: skipped syncing installs from app config. to sync these installs, switch to 'nuon installs sync' command.")
	}

	syncer := sync.New(s.api, appID, version, cfg)
	err = syncer.Sync(ctx)
	if err != nil {
		return ui.PrintError(err)
	}

	if err := s.api.UpdateAppConfigInstalls(ctx, appID, syncer.GetAppConfigID(), &models.ServiceUpdateAppConfigInstallsRequest{
		UpdateAll: true,
	}); err != nil {
		return err
	}

	ui.PrintSuccess("successfully synced " + dir)
	s.notifyOrphanedComponents(syncer.OrphanedComponents())
	s.notifyOrphanedActions(syncer.OrphanedActions())

	cmpsScheduled := syncer.GetComponentsScheduled()
	if len(cmpsScheduled) == 0 {
		return nil
	}

	if err := s.pollComponentBuilds(ctx, cmpsScheduled); err != nil {
		return errors.Wrap(err, "unable to poll builds")
	}

	return nil
}

func (s *Service) notifyOrphanedComponents(cmps map[string]string) {
	if len(cmps) == 0 {
		return
	}

	msg := "Existing component(s) are no longer defined in the config:\n"

	for name, id := range cmps {
		msg += fmt.Sprintf("Component: Name=%s | ID=%s\n", name, id)
	}

	ui.PrintLn(msg)
}

func (s *Service) notifyOrphanedActions(actions map[string]string) {
	if len(actions) == 0 {
		return
	}

	msg := "Existing action(s) are no longer defined in the config:\n"

	for name, id := range actions {
		msg += fmt.Sprintf("Action: Name=%s | ID=%s\n", name, id)
	}

	ui.PrintLn(msg)
}
