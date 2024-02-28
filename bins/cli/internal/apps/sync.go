package apps

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	cfgFilePrefix      string        = "nuon."
	defaultFormat      string        = "toml"
	defaultSyncTimeout time.Duration = time.Minute * 5
	defaultSyncSleep   time.Duration = time.Second * 5
)

func (s *Service) findConfigFiles(format string) ([]string, error) {
	cfgFiles := make([]string, 0)
	if err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(path, cfgFilePrefix) && strings.HasSuffix(path, format) {
			cfgFiles = append(cfgFiles, path)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to look for current config files: %w", err)
	}

	return cfgFiles, nil
}

func (s *Service) formatToAPIFormat(format string) (models.AppAppConfigFmt, error) {
	switch format {
	case "toml":
		return models.AppAppConfigFmtToml, nil
	case "json":
		return models.AppAppConfigFmtJSON, nil
	case "yaml":
		return models.AppAppConfigFmtYaml, nil

	default:
		return "", fmt.Errorf("%s is not a support config format", format)
	}

}

func (s *Service) appIDFromFile(ctx context.Context, file, format string) (string, error) {
	pieces := strings.SplitN(file, ".", 3)
	if len(pieces) != 3 {
		return "", &ui.CLIUserError{
			Msg: fmt.Sprintf("invalid config file must be of the format `nuon.<app-name>.%s`", format),
		}
	}
	appID := pieces[1]

	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return "", fmt.Errorf("app does not exist: %s", pieces[1])
	}

	return appID, nil
}

func (s *Service) sync(ctx context.Context, cfgFile string) error {
	view := ui.NewCreateView(fmt.Sprintf("syncing %s", cfgFile), false)
	view.Start()

	byts, err := os.ReadFile(cfgFile)
	if err != nil {
		view.Fail(err)
		return err
	}

	appID, err := s.appIDFromFile(ctx, cfgFile, defaultFormat)
	if err != nil {
		view.Fail(err)
		return err
	}

	apiFmt, err := s.formatToAPIFormat(defaultFormat)
	if err != nil {
		view.Fail(err)
		return err
	}

	cfg, err := s.api.CreateAppConfig(ctx, appID, &models.ServiceCreateAppConfigRequest{
		Content: generics.ToPtr(string(byts)),
		Format:  apiFmt.Pointer(),
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
		cfgFiles []string
		err      error
	)

	if all {
		cfgFiles, err = s.findConfigFiles(defaultFormat)
		if err != nil {
			ui.PrintError(err)
			return
		}
	}
	if file != "" {
		cfgFiles = []string{file}
	}
	if len(cfgFiles) < 1 {
		ui.PrintError(&ui.CLIUserError{
			Msg: fmt.Sprintf("must set -all or -file, and make sure at least one nuon.<app-name>.%s file exists", defaultFormat),
		})
		return
	}

	for _, cfgFile := range cfgFiles {
		if err := s.sync(ctx, cfgFile); err != nil {
			break
		}
	}
}
