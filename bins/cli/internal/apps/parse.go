package apps

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
)

func (s *Service) Parse(ctx context.Context, file string) {
	view := ui.NewGetView()

	var err error

	appName, err := parse.AppNameFromFilename(file)
	if err != nil {
		view.Error(err)
		return
	}

	if err := s.parse(ctx, parse.File{
		AppName: appName,
		Path:    file,
	}); err != nil {
		view.Error(err)
	}
}

func (s *Service) parse(ctx context.Context, file parse.File) error {
	cfg, err := s.loadConfig(ctx, file.Path)
	if err != nil {
		ui.PrintError(err)
		return err
	}

	byts, err := config.ToJSON(cfg)
	if err != nil {
		ui.PrintError(err)
		return err
	}

	fmt.Println(string(byts))
	return nil
}
