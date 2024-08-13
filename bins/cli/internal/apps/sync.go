package apps

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	defaultSyncTimeout time.Duration = time.Minute * 12
	defaultSyncSleep   time.Duration = time.Second * 20
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

	syncer := sync.New(s.api, appID, cfg)
	if msg, err := syncer.Sync(ctx); err != nil {
		return err
	} else {
		if msg != "" {
			ui.PrintLn(msg)
		}
	}

	ui.PrintSuccess("successfully synced " + cfgFile)

	componentIds := syncer.GetComponentStateIds()
	if len(componentIds) == 0 {
		return nil
	}

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	ui.PrintLn("waiting for components to build")
	for {
		select {
		case <-pollTimeout.Done():
			err = errs.WithUserFacing(err, "timeout waiting for components to build")
			ui.PrintError(err)
			return err
		default:
		}

		finished, err := s.componensBuildsFinished(ctx, appID, componentIds)
		if finished {
			ui.PrintSuccess("component builds completed")
			return nil
		}

		if err != nil {
			err = errs.WithUserFacing(err, "error waiting for components to build")
			ui.PrintError(err)
			return err
		}

		time.Sleep(defaultSyncSleep)
	}
}

func (s *Service) componensBuildsFinished(ctx context.Context, appID string, componentIds []string) (bool, error) {
	components, err := s.api.GetAppComponents(ctx, appID)
	if err != nil {
		return false, err
	}

	for _, comp := range components {
		if !slices.Contains(componentIds, comp.ID) {
			continue
		}
		if comp.Status == statusError {
			return false, errs.NewUserFacing("component build encountered an error: %s", comp.StatusDescription)
		}
		if comp.Status == statusQueued {
			return false, nil
		}
	}

	return true, nil
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
			Msg: "must set -all or -file, and make sure at least one nuon.<app-name>.toml file exists",
		})
		return err
	}

	for _, cfgFile := range cfgFiles {
		ui.PrintLn(fmt.Sprintf("validating file \"%s\"", cfgFile.Path))
		if err := s.validate(ctx, cfgFile, false); err != nil {
			return ui.PrintError(err)
		}
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
