package apps

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
)

const (
	defaultSyncTimeout time.Duration = time.Minute * 5
	defaultSyncSleep   time.Duration = time.Second * 5
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
	if err := syncer.Sync(ctx); err != nil {
		return err
	}

	ui.PrintSuccess("successfully synced " + cfgFile)
	return nil
}

func (s *Service) Sync(ctx context.Context, all bool, file string) error {
	var (
		cfgFiles []parse.File
		err      error
	)

	if all {
		cfgFiles, err = parse.FindConfigFiles(".")
		if err != nil {
			return ui.PrintError(err)
		}
	}
	if file != "" {
		appName, err := parse.AppNameFromFilename(file)
		if err != nil {
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
		appID, err := lookup.AppID(ctx, s.api, cfgFile.AppName)
		if err != nil {
			return ui.PrintError(err)
		}

		if err := s.sync(ctx, cfgFile.Path, appID); err != nil {
			return ui.PrintError(err)
		}
	}

	return nil
}
