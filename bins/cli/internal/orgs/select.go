package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
)

func (s *Service) Select(ctx context.Context, orgID string, asJSON bool) error {
	view := ui.NewGetView()

	if orgID != "" {
		s.SetCurrent(ctx, orgID, asJSON)
	} else {
		orgs, _, err := s.list(ctx, 0, 50)
		if err != nil {
			return view.Error(err)
		}

		if len(orgs) == 0 {
			s.printNoOrgsMsg()
			return nil
		}

		// Convert orgs to selector options
		orgOptions := make([]bubbles.OrgOption, len(orgs))
		for i, org := range orgs {
			// TODO: Detect evaluation orgs based on user journey data
			orgOptions[i] = bubbles.OrgOption{
				ID:           org.ID,
				Name:         org.Name,
				IsEvaluation: false, // Will be updated when user journey detection is added
			}
		}

		// Show org selector
		selectedOrgID, err := bubbles.SelectOrg(orgOptions)
		if err != nil {
			return view.Error(err)
		}

		if err := s.setOrgID(ctx, selectedOrgID); err != nil {
			return view.Error(err)
		}

		// Find selected org for display
		var selectedOrg *models.AppOrg
		for _, org := range orgs {
			if org.ID == selectedOrgID {
				selectedOrg = org
				break
			}
		}

		if selectedOrg != nil {
			s.printOrgSetMsg(selectedOrg.Name, selectedOrg.ID)
		}
	}
	return nil
}
