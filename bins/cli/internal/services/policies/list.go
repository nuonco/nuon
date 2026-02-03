package policies

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) List(ctx context.Context, appID string, offset, limit int, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	configs, hasMore, err := s.api.GetAppPoliciesConfigs(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch policies configs"))
	}

	if asJSON {
		ui.PrintJSON(configs)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"CREATED AT",
			"POLICIES COUNT",
		},
	}
	for _, cfg := range configs {
		policiesCount := 0
		if cfg.Policies != nil {
			policiesCount = len(cfg.Policies)
		}
		data = append(data, []string{
			cfg.ID,
			cfg.CreatedAt,
			fmt.Sprintf("%d", policiesCount),
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}
