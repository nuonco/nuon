package orgs

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SetCurrent(ctx context.Context, orgID string, asJSON bool) {
	view := ui.NewGetView()
	s.api.SetOrgID(orgID)
	org, err := s.api.GetOrg(ctx)
	if err != nil {
		userErr, isUserError := nuon.ToUserError(err)
		if isUserError && userErr.Error == s.notFoundErr(orgID).Error() {
			s.printOrgNotFoundMsg(orgID)
		} else {
			view.Error(err)
		}

		return
	}

	if err := s.setOrgInConfig(ctx, orgID); err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(org)
	} else {
		s.printOrgSetMsg(org.Name, org.ID)
	}
}
