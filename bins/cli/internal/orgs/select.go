package orgs

import (
	"context"
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/pterm/pterm"
)

func (s *Service) Select(ctx context.Context, orgID string, asJSON bool) {
	view := ui.NewGetView()

	if orgID != "" {
		s.SetCurrent(ctx, orgID, asJSON)
	} else {
		orgs, err := s.api.GetOrgs(ctx)
		if err != nil {
			view.Error(err)
			return
		}

		if len(orgs) == 0 {
			s.printNoOrgsMsg()
			return
		}

		// select options
		var options []string
		for _, org := range orgs {
			options = append(options, fmt.Sprintf("%s: %s", org.Name, org.ID))
		}

		// select org prompt
		selectedOrg, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
		org := strings.Split(selectedOrg, ":")

		if err := s.setOrgInConfig(ctx, strings.ReplaceAll(org[1], " ", "")); err != nil {
			view.Error(err)
			return
		}

		s.printOrgSetMsg(org[0], org[1])
	}
}
