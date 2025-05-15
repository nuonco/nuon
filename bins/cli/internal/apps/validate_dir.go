package apps

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/validate"
	"github.com/powertoolsdev/mono/pkg/errs"
)

func (s *Service) ValidateDir(ctx context.Context, dir string) error {
	ui.PrintLn("syncing directory from " + dir)

	appName, err := parse.AppNameFromDirName(dir)
	if err != nil {
		err = errs.WithUserFacing(err, "error parsing app name from file")
		return ui.PrintError(err)
	}

	_, err = lookup.AppID(ctx, s.api, appName)
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

	return nil
}
