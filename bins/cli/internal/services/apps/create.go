package apps

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/errs"
)

const (
	statusError                  string      = "error"
	statusActive                 string      = "active"
	statusQueued                 string      = "queued"
	defaultConfigFilePermissions fs.FileMode = 0o644
)

func (s *Service) Create(ctx context.Context, appName string, asJSON, noSelect bool) error {
	var view *ui.CreateView
	if !asJSON {
		view = ui.NewCreateView("app", asJSON, s.cfg.Interactive)
		view.Start()
		view.Update("creating app")
	}

	app, err := s.api.CreateApp(ctx, &models.ServiceCreateAppRequest{
		Name: &appName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicated key") {
			err = errs.WithUserFacing(err, "%s", fmt.Sprintf("An application already exists with the name %q", appName))
		}
		if asJSON {
			return ui.PrintError(err)
		}
		return view.Fail(err)
	}

	if !asJSON {
		view.Update("waiting for app to be completed")
	}
	for {
		currentApp, err := s.api.GetApp(ctx, app.ID)
		switch {
		case err != nil:
			if asJSON {
				return ui.PrintError(err)
			}
			return view.Fail(err)
		case currentApp.Status == statusError:
			err := fmt.Errorf("failed to create app: %s", currentApp.StatusDescription)
			if asJSON {
				return ui.PrintError(err)
			}
			return view.Fail(err)
		case currentApp.Status == statusActive:
			goto success
		default:
			if !asJSON {
				view.Update(fmt.Sprintf("%s app", currentApp.Status))
			}
		}

		time.Sleep(5 * time.Second)
	}

success:
	if !noSelect {
		if err := s.setAppID(ctx, app.ID); err != nil && !asJSON {
			view.Fail(errs.NewUserFacing("failed to set new app as current: %s", err))
		}
	}

	if asJSON {
		ui.PrintJSON(map[string]string{
			"id":      app.ID,
			"name":    appName,
			"status":  "created",
			"message": fmt.Sprintf("successfully created app %s", app.ID),
		})
		return nil
	}

	view.Success(app.ID)
	if !noSelect {
		s.printAppSetMsg(appName, app.ID)
	}
	return nil
}

func (s *Service) writeFile(ctx context.Context, appID string, templateType models.ServiceAppConfigTemplateType, view *ui.CreateView) (*models.ServiceAppConfigTemplate, error) {
	view.Update("generating app config template " + string(templateType))
	tmpl, err := s.api.GetAppConfigTemplate(ctx, appID, templateType)
	if err != nil {
		return nil, err
	}

	view.Update("writing template " + string(templateType) + " config to file")
	err = os.WriteFile(tmpl.Filename, []byte(tmpl.Content), defaultConfigFilePermissions)
	if err != nil {
		return tmpl, err
	}

	return tmpl, nil
}
