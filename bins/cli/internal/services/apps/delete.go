package apps

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		err = s.ensureNoActiveComponents(ctx, appID)
		if err != nil {
			ui.PrintJSONError(err)
			return err
		}

		res, err := s.api.DeleteApp(ctx, appID)
		if err != nil {
			ui.PrintJSONError(err)
			return err
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		r := response{ID: appID, Deleted: res}
		ui.PrintJSON(r)
		return nil
	}

	view := ui.NewDeleteView("app", appID)
	view.Start()

	view.Update("removing all components from app config")
	err = s.ensureNoActiveComponents(ctx, appID)
	if err != nil {
		return view.Fail(err)
	}

	view.Update("deleting app")

	_, err = s.api.DeleteApp(ctx, appID)
	if err != nil {
		return view.Fail(err)
	}

	// unset appID if it is the currentAppID
	currentAppID := s.getAppID()
	if appID == currentAppID {
		if err := s.unsetAppID(ctx); err != nil {
			return view.Fail(err)
		}
	}

	view.SuccessQueued()
	return nil
}

func (s *Service) ensureNoActiveComponents(ctx context.Context, appID string) error {
	appCfg, err := s.api.GetAppLatestConfig(ctx, appID)
	if err != nil {
		return err
	}

	if len(appCfg.ComponentIds) > 0 {
		newCfg, err := s.api.CreateAppConfig(ctx, appID, &models.ServiceCreateAppConfigRequest{
			Readme:     appCfg.Readme,
			CliVersion: appCfg.CliVersion,
		})
		if err != nil {
			return err
		}

		_, err = s.api.UpdateAppConfig(ctx, appID, newCfg.ID, &models.ServiceUpdateAppConfigRequest{
			ComponentIds:      make([]string, 0),
			State:             appCfg.State,
			Status:            models.AppAppConfigStatusActive,
			StatusDescription: "removing all components before app deletion",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
