package orgs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func toOrgOptions(orgs []*models.AppOrg) []bubbles.OrgOption {
	opts := make([]bubbles.OrgOption, len(orgs))
	for i, org := range orgs {
		opts[i] = bubbles.OrgOption{
			ID:           org.ID,
			Name:         org.Name,
			IsEvaluation: false,
		}
	}
	return opts
}

func (s *Service) Select(ctx context.Context, orgID string, offset, limit int, asJSON bool) error {
	view := ui.NewGetView()

	if orgID != "" {
		s.SetCurrent(ctx, orgID, asJSON)
	} else {
		orgs, _, err := s.list(ctx, offset, limit, "")
		if err != nil {
			return view.Error(err)
		}

		if len(orgs) == 0 {
			s.printNoOrgsMsg()
			return nil
		}

		allLoadedOrgs := orgs
		orgOptions := toOrgOptions(orgs)

		searchFn := func(q string) ([]bubbles.OrgOption, error) {
			results, _, err := s.list(ctx, 0, limit, q)
			if err != nil {
				return nil, err
			}
			for _, r := range results {
				found := false
				for _, existing := range allLoadedOrgs {
					if existing.ID == r.ID {
						found = true
						break
					}
				}
				if !found {
					allLoadedOrgs = append(allLoadedOrgs, r)
				}
			}
			return toOrgOptions(results), nil
		}

		selectedOrgID, err := bubbles.SelectOrg(orgOptions, searchFn, s.cfg.Interactive)
		if err != nil {
			return view.Error(err)
		}

		if err := s.setOrgID(ctx, selectedOrgID); err != nil {
			return view.Error(err)
		}

		var selectedOrg *models.AppOrg
		for _, org := range allLoadedOrgs {
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
