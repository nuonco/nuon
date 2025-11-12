package apps

import (
	"context"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/errs"
)

func (s *Service) Rename(ctx context.Context, appID string, name string, rename, asJSON bool) error {
	view := ui.NewCreateView("app", asJSON)
	view.Start()

	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return err
	}

	view.Update("fetching app")
	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		return view.Fail(err)
	}
	if app.Name == name {
		return view.Fail(errors.New("Must provide a different name."))
	}

	view.Update("updating app")
	_, err = s.api.UpdateApp(ctx, appID, &models.ServiceUpdateAppRequest{
		Name: name,
	})
	if err != nil {
		return view.Fail(err)
	}

	origFp := parse.FilenameFromAppName(app.Name)
	newFp := parse.FilenameFromAppName(name)
	_, err = os.Stat(origFp)
	if err != nil {
		return view.Fail(errs.WithUserFacing(err, "no config file found"))
	}

	_, err = os.Stat(newFp)
	if err == nil {
		return view.Fail(errs.NewUserFacing("%s", "config file already exists at "+newFp))
	}

	if rename {
		view.Update("renaming config file")
		err := os.Rename(origFp, newFp)
		if err != nil {
			return view.Fail(errs.WithUserFacing(err, "failed to rename config file"))
		}
	}

	return nil
}
