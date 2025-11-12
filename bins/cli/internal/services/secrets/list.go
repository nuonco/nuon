package secrets

import (
	"context"
	"strings"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, offset, limit int, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	secrets, hasMore, err := s.list(ctx, appID, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(secrets)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"VALUE",
			"CREATED-BY",
			"CREATED-AT",
		},
	}
	for _, secret := range secrets {
		createdAt, err := time.Parse(time.RFC3339Nano, secret.CreatedAt)
		if err != nil {
			return view.Error(err)
		}

		data = append(data, []string{
			secret.ID,
			secret.Name,
			strings.Repeat("*", int(secret.Length)),
			createdAt.Format(time.Stamp),
		})
	}

	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) list(ctx context.Context, appID string, offset, limit int) ([]*models.AppAppSecret, bool, error) {
	cmps, hasMore, err := s.api.GetAppSecrets(ctx, appID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, false, err
	}
	return cmps, hasMore, nil
}
