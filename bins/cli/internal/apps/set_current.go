package apps

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SetCurrent(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewGetView()
	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		userErr, isUserError := nuon.ToUserError(err)
		if isUserError && userErr.Error == s.notFoundErr(appID).Error() {
			s.printAppNotFoundMsg(appID)
		} else {
			view.Error(err)
		}

		return err
	}

	if err := s.setAppInConfig(ctx, appID); err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(app)
	} else {
		s.printAppSetMsg(app.Name, app.ID)
	}
	return nil
}
