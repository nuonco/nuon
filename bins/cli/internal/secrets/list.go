package secrets

import (
	"context"
	"strings"
	"time"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	secrets, err := s.list(ctx, appID)
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

	view.Render(data)
	return nil
}

func (s *Service) list(ctx context.Context, appID string) ([]*models.AppAppSecret, error) {
	if !s.cfg.PaginationEnabled {
		releases, _, err := s.api.GetAppSecrets(ctx, appID, &models.GetAppSecretsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return releases, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppAppSecret, bool, error) {
		cmps, hasMore, err := s.api.GetAppSecrets(ctx, appID, &models.GetAppSecretsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return cmps, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
