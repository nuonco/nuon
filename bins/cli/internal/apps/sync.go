package apps

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/config/validate"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	defaultSyncTimeout           time.Duration = time.Minute * 12
	defaultSyncSleep             time.Duration = time.Second * 20
	componentBuildStatusError                  = "error"
	componentBuildStatusBuilding               = "building"
	componentBuildStatusActive                 = "active"
	componentStatusQueued                      = "queued"
)

func (s *Service) sync(ctx context.Context, cfgFile, appID string) error {
	cfg, err := parse.Parse(parse.ParseConfig{
		Context:     config.ConfigContextSource,
		Filename:    cfgFile,
		BackendType: config.BackendTypeLocal,
		Template:    true,
		V:           validator.New(),
	})
	if err != nil {
		return err
	}

	ui.PrintLn(fmt.Sprintf("validating file \"%s\"", cfgFile))
	err = validate.Validate(ctx, s.v, cfg)
	if err != nil {
		if config.IsWarningErr(err) {
			ui.PrintError(err)
		} else {
			return err
		}
	}

	syncer := sync.New(s.api, appID, cfg)
	err = syncer.Sync(ctx)
	if err != nil {
		return err
	}

	ui.PrintSuccess("successfully synced " + cfgFile)

	s.notifyOrphanedComponents(syncer.OrphanedComponents())
	s.notifyOrphanedActions(syncer.OrphanedActions())

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
	return
}

func (s *Service) Sync(ctx context.Context, all bool, file string) error {
	var (
		cfgFiles []parse.File
		err      error
	)

	if all {
		cfgFiles, err = parse.FindConfigFiles(".")
		if err != nil {
			err = errs.WithUserFacing(err, "error parsing toml files")
			return ui.PrintError(err)
		}
	}
	if file != "" {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return ui.PrintError(&ui.CLIUserError{
				Msg: "specified file doesn't exist",
			})
		}

		appName, err := parse.AppNameFromFilename(file)
		if err != nil {
			err = errs.WithUserFacing(err, "error parsing app name from file")
			return ui.PrintError(err)
		}

		cfgFiles = []parse.File{
			{
				Path:    file,
				AppName: appName,
			},
		}
	}

	if len(cfgFiles) < 1 {
		ui.PrintError(&ui.CLIUserError{
			Msg: "must set -c, --file, or --all and make sure at least one nuon.<app-name>.toml file exists",
		})
		return err
	}

	for _, cfgFile := range cfgFiles {
		appID, err := lookup.AppID(ctx, s.api, cfgFile.AppName)
		if err != nil {
			err = errs.WithUserFacing(err, "error looking up app id")
			return ui.PrintError(err)
		}

		if err := s.sync(ctx, cfgFile.Path, appID); err != nil {
			return ui.PrintError(err)
		}
	}

	return nil
}
