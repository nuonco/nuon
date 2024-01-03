package installs

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SetCurrent(ctx context.Context, installID string, asJSON bool) {
	view := ui.NewGetView()
	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		userErr, isUserError := nuon.ToUserError(err)
		if isUserError && userErr.Error == s.notFoundErr(installID).Error() {
			s.printInstallNotFoundMsg(installID)
		} else {
			view.Error(err)
		}

		return
	}

	if err := s.setInstallInConfig(ctx, installID); err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(install)
	} else {
		s.printInstallSetMsg(install.Name, install.ID)
	}
}
