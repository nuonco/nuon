package variables

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) Create(ctx context.Context, appID, name, value string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	var view *ui.CreateView
	if !asJSON {
		view = ui.NewCreateView("variable", asJSON, s.cfg.Interactive)
		view.Start()
	}

	secret, err := s.api.CreateAppSecret(ctx, appID, &models.ServiceCreateAppSecretRequest{
		Name:  &name,
		Value: &value,
	})
	if err != nil {
		if asJSON {
			return ui.PrintError(err)
		}
		return view.Fail(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]string{
			"id":      secret.ID,
			"app_id":  appID,
			"name":    name,
			"status":  "created",
			"message": fmt.Sprintf("successfully created variable (%s)", secret.ID),
		})
		return nil
	}

	view.Success(secret.ID)
	return nil
}
