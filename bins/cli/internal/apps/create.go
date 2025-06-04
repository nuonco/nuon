package apps

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	statusError                  string      = "error"
	statusActive                 string      = "active"
	statusQueued                 string      = "queued"
	defaultConfigFilePermissions fs.FileMode = 0o644
)

func (s *Service) Create(ctx context.Context, appName string, appTemplate string, noTemplate, asJSON, noSelect bool) error {
	view := ui.NewCreateView("app", asJSON)
	view.Start()
	view.Update("creating app")
	app, err := s.api.CreateApp(ctx, &models.ServiceCreateAppRequest{
		Name: &appName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicated key") {
			err = errs.WithUserFacing(err, "%s", fmt.Sprintf("An application already exists with the name %q", appName))
		}
		return view.Fail(err)
	}

	view.Update("waiting for app to be completed")
	for {
		currentApp, err := s.api.GetApp(ctx, app.ID)
		switch {
		case err != nil:
			return view.Fail(err)
		case currentApp.Status == statusError:
			return view.Fail(fmt.Errorf("failed to create app: %s", currentApp.StatusDescription))
		case currentApp.Status == statusActive:
			view.Success(currentApp.ID)
			goto success
		default:
			view.Update(fmt.Sprintf("%s app", currentApp.Status))
		}

		time.Sleep(5 * time.Second)
	}

success:
	if !noSelect {
		if err := s.setAppID(ctx, app.ID); err == nil {
			s.printAppSetMsg(appName, app.ID)
		} else {
			view.Fail(errs.NewUserFacing("failed to set new app as current: %s", err))
		}
	}

	if noTemplate {
		return nil
	}

	flatTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeFlat), view)
	if err != nil {
		return view.Fail(err)
	}

	view.Update("successfully wrote config template files at\n" +
		flatTmpl.Filename + "\n",
	)
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
