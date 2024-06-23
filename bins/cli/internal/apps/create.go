package apps

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

const (
	statusError                  string      = "error"
	statusActive                 string      = "active"
	defaultConfigFilePermissions fs.FileMode = 0o644
)

func (s *Service) Create(ctx context.Context, appName string, appTemplate string, noTemplate, asJSON bool) {
	view := ui.NewCreateView("app", asJSON)
	view.Start()
	view.Update("creating app")
	app, err := s.api.CreateApp(ctx, &models.ServiceCreateAppRequest{
		Name: &appName,
	})
	if err != nil {
		view.Fail(err)
		return
	}

	view.Update("waiting for app to be completed")
	for {
		currentApp, err := s.api.GetApp(ctx, app.ID)
		switch {
		case err != nil:
			view.Fail(err)
		case currentApp.Status == statusError:
			view.Fail(fmt.Errorf("failed to create app: %s", currentApp.StatusDescription))
			return
		case currentApp.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created	app %s", currentApp.ID))
			goto success
		default:
			view.Update(fmt.Sprintf("%s app", currentApp.Status))
		}

		time.Sleep(5 * time.Second)
	}

success:
	if noTemplate {
		return
	}

	// create template
	view.Update("generating app config template")
	tmpl, err := s.api.GetAppConfigTemplate(ctx, app.ID, models.ServiceAppConfigTemplateType(appTemplate))
	if err != nil {
		view.Fail(err)
		return
	}

	view.Update("writing template config to file")
	err = os.WriteFile(tmpl.Filename, []byte(tmpl.Content), defaultConfigFilePermissions)
	if err != nil {
		view.Fail(err)
		return
	}

	view.Update("successfully wrote config template file at " + tmpl.Filename + "\n")
}
