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
	"github.com/pterm/pterm"
)

const (
	defaultSyncTimeout time.Duration = time.Minute * 5
	defaultSyncSleep   time.Duration = time.Second * 5
)

func (s *Service) sync(ctx context.Context, cfgFile, appID string) error {
	view := ui.NewCreateView(fmt.Sprintf("updated configs for %s", cfgFile), false)
	view.Start()

	byts, err := os.ReadFile(cfgFile)
	if err != nil {
		view.Fail(err)
		return err
	}

	tfJSON, err := parse.ToTerraformJSON(parse.ParseConfig{
		Context:     config.ConfigContextSource,
		Bytes:       byts,
		BackendType: config.BackendTypeLocal,
		Template:    true,
		V:           validator.New(),
	})
	if err != nil {
		view.Fail(err)
		return err
	}

	validateOutput, err := s.execTerraformValidate(ctx, appID, tfJSON)
	if err != nil {
		return err
	}

	if len(validateOutput.Diagnostics) > 0 {
		err = fmt.Errorf("configuration is invalid, %d errors found", len(validateOutput.Diagnostics))
		view.Fail(err)

		data := [][]string{
			{"RESOURCE", "SUMMARY", "ERROR"},
		}
		for _, diag := range validateOutput.Diagnostics {
			data = append(data, []string{
				*diag.Snippet.Context,
				diag.Summary,
				diag.Detail,
			})
		}

		pterm.DefaultTable.
			WithData(data).
			WithHasHeader().
			WithHeaderRowSeparator("-").
			Render()

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
}

func (s *Service) Sync(ctx context.Context, all bool, file string, asJSON bool) error {
	var (
		cfgFiles []parse.File
		err      error
	)

	if all {
		cfgFiles, err = parse.FindConfigFiles(".")
		if err != nil {
			ui.PrintError(err)
			return err
		}
	}
	if file != "" {
		appName, err := parse.AppNameFromFilename(file)
		if err != nil {
			ui.PrintError(err)
			return err
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
			ui.PrintError(err)
			return err
		}

		if err := s.sync(ctx, cfgFile.Path, appID); err != nil {
			return err
		}
	}
	return nil
}
