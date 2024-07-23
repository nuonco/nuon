package apps

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	statusError                  string      = "error"
	statusActive                 string      = "active"
	defaultConfigFilePermissions fs.FileMode = 0o644
)

func (s *Service) Create(ctx context.Context, appName string, appTemplate string, noTemplate, asJSON bool) error {
	view := ui.NewCreateView("app", asJSON)
	view.Start()
	view.Update("creating app")
	app, err := s.api.CreateApp(ctx, &models.ServiceCreateAppRequest{
		Name: &appName,
	})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicated key"):
			err = errs.UserFacingError(err, fmt.Sprintf("An application already exists with the name %q", appName))
		default:
			err = errors.Wrap(err, "error creating org")
		}
		view.Fail(err)
		return err
	}

	view.Update("waiting for app to be completed")
	for {
		currentApp, err := s.api.GetApp(ctx, app.ID)
		switch {
		case err != nil:
			view.Fail(err)
			return err
		case currentApp.Status == statusError:
			err := fmt.Errorf("failed to create app: %s", currentApp.StatusDescription)
			view.Fail(err)
			return err
		case currentApp.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created app %s", currentApp.ID))
			goto success
		default:
			view.Update(fmt.Sprintf("%s app", currentApp.Status))
		}

		time.Sleep(5 * time.Second)
	}

success:
	if noTemplate {
		return nil
	}

	// create template
	view.Update("generating app config template")
	tmpl, err := s.api.GetAppConfigTemplate(ctx, app.ID, models.ServiceAppConfigTemplateType(appTemplate))
	if err != nil {
		view.Fail(err)
		return err
	}

	view.Update("writing template config to file")
	err = os.WriteFile(tmpl.Filename, []byte(tmpl.Content), defaultConfigFilePermissions)
	if err != nil {
		view.Fail(err)
		return err
	}

	view.Update("successfully wrote config template file at " + tmpl.Filename + "\n")
	return nil
}
