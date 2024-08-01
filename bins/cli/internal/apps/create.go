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
		if strings.Contains(err.Error(), "duplicated key") {
			err = errs.WithUserFacing(err, fmt.Sprintf("An application already exists with the name %q", appName))
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

	topLevelTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeTopDashLevel), view)
	if err != nil {
		return view.Fail(err)
	}

	installerTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeInstaller), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	runnerTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeRunner), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	sandboxTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeSandbox), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	inputsTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeInputs), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	terraformTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeTerraform), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	terraformInfraTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeTerraformInfra), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	helmTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeHelm), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	dockerBuildTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeDockerDashBuild), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	containerImageTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeContainerDashImage), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	jobTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeJob), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	ecrContainerTmpl, err := s.writeFile(ctx, app.ID, models.ServiceAppConfigTemplateType(models.ServiceAppConfigTemplateTypeEcrDashContainerDashImage), view)
	if err != nil {
		view.Fail(err)
		return err
	}

	view.Update("successfully wrote config template files at\n" + 
				topLevelTmpl.Filename + "\n" +
				installerTmpl.Filename	+ "\n" +
				runnerTmpl.Filename + "\n" +
				sandboxTmpl.Filename + "\n" +
				inputsTmpl.Filename + "\n" +
				terraformTmpl.Filename + "\n" +
				terraformInfraTmpl.Filename + "\n" +
				helmTmpl.Filename + "\n" +
				dockerBuildTmpl.Filename + "\n" +
				containerImageTmpl.Filename + "\n" +
				jobTmpl.Filename + "\n" +
				ecrContainerTmpl.Filename + "\n",
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
