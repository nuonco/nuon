package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/pterm/pterm"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
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

	syncer := sync.New(s.api, appID, cfg)
	msg, cmpbuildsScheduled, err := syncer.Sync(ctx)
	if err != nil {
		return err
	} else {
		if msg != "" {
			ui.PrintLn(msg)
		}
	}

	ui.PrintSuccess("successfully synced " + cfgFile)

	if len(cmpbuildsScheduled) == 0 {
		return nil
	}

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	multi := pterm.DefaultMultiPrinter

	spinnersByComponentID := make(map[string]*pterm.SpinnerPrinter)
	for _, cmpID := range cmpbuildsScheduled {
		spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("building component " + cmpID)
		spinnersByComponentID[cmpID] = spinner
	}

	multi.Start()

	// NOTE: on updates components are already active and new component_builds records wait to be created.
	// So we need to wait for the new component_builds to be created before we start to poll.
	time.Sleep(time.Second * 5)

	for {
		select {
		case <-pollTimeout.Done():
			err = fmt.Errorf("timeout waiting for components to build")
			ui.PrintError(err)
			for cmpID, spinner := range spinnersByComponentID {
				spinner.Fail("timeout waiting for component " + cmpID + " to build")
			}
			multi.Stop()
			return err
		default:
		}

		for cmpID := range spinnersByComponentID {
			cmpBuild, err := s.api.GetComponentLatestBuild(ctx, cmpID)
			if err != nil {
				if nuon.IsServerError(err) {
					spinnersByComponentID[cmpID].Fail("error building component " + cmpID)
					delete(spinnersByComponentID, cmpID)
					continue
				}
				// in case we didn't wait long enough for an initial build record, ignore and loop again
				if nuon.IsNotFound(err) {
					continue
				}
				// TODO: avoid panic if we error on network issues. We should introduce a retryer at the sdk level.
				// for now, this loop is inherently retrying.
				if cmpBuild == nil {
					continue
				}
			}
			if cmpBuild.Status == componentBuildStatusError {
				spinnersByComponentID[cmpID].Fail("error building component " + cmpID)
				delete(spinnersByComponentID, cmpID)
				continue
			}

			if cmpBuild.Status == componentBuildStatusActive {
				spinnersByComponentID[cmpID].Success("finished building component " + cmpID)
				delete(spinnersByComponentID, cmpID)
				continue
			}
		}

		if len(spinnersByComponentID) == 0 {
			multi.Stop()
			return nil
		}

		time.Sleep(defaultSyncSleep)
	}
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
