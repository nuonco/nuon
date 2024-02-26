package apps

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	cfgFilePrefix string = "nuon."
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
		return "", fmt.Errorf("unable to lookup app: %w", err)
	}

	return appID, nil
}

func (s *Service) Sync(ctx context.Context, all bool, file, format string, asJSON bool) {
	var (
		cfgFiles []string
		err      error
	)

	if all {
		cfgFiles, err = s.findConfigFiles(format)
		if err != nil {
			ui.PrintError(err)
			return
		}
	} else if file != "" {
		cfgFiles = []string{file}
	}
	if len(cfgFiles) < 1 {
		ui.PrintError(&ui.CLIUserError{
			Msg: fmt.Sprintf("must set -all or -file, and make sure at least one nuon.<app-name>.%s file exists", format),
		})
		return
	}

	view := ui.NewGetView()
	rows := [][]string{
		{
			"file",
			"app-id",
			"status",
		},
	}
	for _, cfgFile := range cfgFiles {
		byts, err := os.ReadFile(cfgFile)
		if err != nil {
			ui.PrintError(err)
			return
		}

		appID, err := s.appIDFromFile(ctx, cfgFile, format)
		if err != nil {
			ui.PrintError(err)
			return
		}

		apiFmt, err := s.formatToAPIFormat(format)
		if err != nil {
			ui.PrintError(err)
			return
		}

		cfg, err := s.api.CreateAppConfig(ctx, appID, &models.ServiceCreateAppConfigRequest{
			Content: generics.ToPtr(string(byts)),
			Format:  apiFmt.Pointer(),
		})
		if err != nil {
			ui.PrintError(err)
			return
		}
		rows = append(rows, []string{
			cfgFile,
			appID,
			string(cfg.Status),
		})
	}

	view.Render(rows)
}
