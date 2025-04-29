package apps

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/config/validate"
	"github.com/powertoolsdev/mono/pkg/errs"
)

func (s *Service) SyncDir(ctx context.Context, dir string) error {
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
		Dirname: dir,
		V:       validator.New(),
	})
	if err != nil {
		return ui.PrintError(err)
	}

	if s.cfg.Debug {
		ui.PrintJSON(cfg)
	}

	ui.PrintLn(fmt.Sprintf("validating configs"))
	err = validate.Validate(ctx, s.v, cfg)
	if err != nil {
		if config.IsWarningErr(err) {
			ui.PrintError(err)
		} else {
			return ui.PrintError(err)
		}
	}
	ui.PrintLn(fmt.Sprintf("all configs valid"))

	syncer := sync.New(s.api, appID, cfg)
	err = syncer.Sync(ctx)
	if err != nil {
		return ui.PrintError(err)
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
