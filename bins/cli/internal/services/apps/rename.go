package apps

import (
	"context"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/errs"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) Rename(ctx context.Context, appID string, name string, rename, asJSON bool) error {
	var view *ui.SpinnerView
	if !asJSON {
		view = ui.NewSpinnerView(asJSON, s.cfg.Interactive)
		view.Start("renaming app")
	}

	fail := func(err error) error {
		if asJSON {
			return ui.PrintError(err)
		}
		view.Fail(err)
		return err
	}

	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	if !asJSON {
		view.Update("fetching app")
	}
	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		return fail(err)
	}
	if app.Name == name {
		return fail(errors.New("Must provide a different name."))
	}

	if !asJSON {
		view.Update("updating app")
	}
	_, err = s.api.UpdateApp(ctx, appID, &models.ServiceUpdateAppRequest{
		Name: name,
	})
	if err != nil {
		return fail(err)
	}

	origFp := parse.FilenameFromAppName(app.Name)
	newFp := parse.FilenameFromAppName(name)
	_, err = os.Stat(origFp)
	if err != nil {
		return fail(errs.WithUserFacing(err, "no config file found"))
	}

	_, err = os.Stat(newFp)
	if err == nil {
		return fail(errs.NewUserFacing("%s", "config file already exists at "+newFp))
	}

	if rename {
		if !asJSON {
			view.Update("renaming config file")
		}
		if err := os.Rename(origFp, newFp); err != nil {
			return fail(errs.WithUserFacing(err, "failed to rename config file"))
		}
	}

	if asJSON {
		ui.PrintJSON(map[string]string{
			"id":      appID,
			"name":    name,
			"status":  "renamed",
			"message": fmt.Sprintf("renamed app to %s", name),
		})
		return nil
	}

	view.Success(fmt.Sprintf("renamed app to %s", name))
	return nil
}
