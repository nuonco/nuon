package apps

import (
	"context"
	"fmt"
	"os"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config/parse"
)

func (s *Service) Rename(ctx context.Context, appID string, name string, rename, asJSON bool) {
	view := ui.NewCreateView("app", asJSON)
	view.Start()

	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view.Update("fetching app")
	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		view.Fail(err)
		return
	}
	if app.Name == name {
		view.Fail(fmt.Errorf("Must provide a different name."))
		return
	}

	view.Update("updating app")
	_, err = s.api.UpdateApp(ctx, appID, &models.ServiceUpdateAppRequest{
		Name: name,
	})
	if err != nil {
		view.Fail(err)
		return
	}

	origFp := parse.FilenameFromAppName(app.Name)
	newFp := parse.FilenameFromAppName(name)
	_, err = os.Stat(origFp)
	if err != nil {
		view.Update("no config file found")
		return
	}

	_, err = os.Stat(newFp)
	if err == nil {
		view.Update("config file already exists at " + newFp)
		return
	}

	if err != nil && rename {
		view.Update("renaming config file")
		err := os.Rename(origFp, newFp)
		if err != nil {
			view.Fail(err)
			return
		}
	}

	return
}
