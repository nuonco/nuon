package secrets

import (
	"context"
	"strings"
	"time"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	secrets, err := s.api.GetAppSecrets(ctx, appID)
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
			secret.CreatedBy.Email,
			createdAt.Format(time.Stamp),
		})
	}

	view.Render(data)
	return nil
}
