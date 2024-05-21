package apps

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	defaultSyncTimeout time.Duration = time.Minute * 5
	defaultSyncSleep   time.Duration = time.Second * 5
)

func (s *Service) sync(ctx context.Context, cfgFile, appID string) error {
	view := ui.NewCreateView(fmt.Sprintf("syncing %s", cfgFile), false)
	view.Start()

	byts, err := os.ReadFile(cfgFile)
	if err != nil {
		view.Fail(err)
		return err
	}

	tfJSON, err := parse.ToTerraformJSON(parse.ParseConfig{
		Context:     config.ConfigContextSource,
		Bytes:       byts,
		BackendType: config.BackendTypeS3,
		Template:    true,
		V:           validator.New(),
	})
	if err != nil {
		view.Fail(err)
		return err
	}

	cfg, err := s.api.CreateAppConfig(ctx, appID, &models.ServiceCreateAppConfigRequest{
		Content:                generics.ToPtr(string(byts)),
		GeneratedTerraformJSON: string(tfJSON),
	})
	if err != nil {
		view.Fail(err)
		return err
	}

	pollTimeout, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	view.Update("waiting for app to be synced")
	for {
		select {
		case <-pollTimeout.Done():
			err := fmt.Errorf("timeout syncing")
			view.Fail(err)
			return err
		default:
		}

		cfg, err := s.api.GetAppConfig(ctx, appID, cfg.ID)
		if err != nil {
			view.Fail(err)
			return err
		}

		switch cfg.Status {
		case models.AppAppConfigStatusActive:
			view.Success("successfully synced " + cfgFile)
			return nil
		case models.AppAppConfigStatusError:
			view.Fail(fmt.Errorf("failed to sync :%s", cfg.Status))
			return nil
		case models.AppAppConfigStatusOutdated:
			view.Success("config is out dated")
			return nil
		default:
			view.Update(string(cfg.Status))
		}

		time.Sleep(defaultSyncSleep)
	}

	return nil
}

func (s *Service) Sync(ctx context.Context, all bool, file string, asJSON bool) {
	var (
		cfgFiles []parse.File
		err      error
	)

	if all {
		cfgFiles, err = parse.FindConfigFiles(".")
		if err != nil {
			ui.PrintError(err)
			return
		}
	}
	if file != "" {
		appName, err := parse.AppNameFromFilename(file)
		if err != nil {
			ui.PrintError(err)
			return
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
			Msg: fmt.Sprintf("must set -all or -file, and make sure at least one nuon.<app-name>.toml file exists"),
		})
		return
	}

	for _, cfgFile := range cfgFiles {
		appID, err := lookup.AppID(ctx, s.api, cfgFile.AppName)
		if err != nil {
			ui.PrintError(err)
			return
		}

		if err := s.sync(ctx, cfgFile.Path, appID); err != nil {
			break
		}
	}
}
