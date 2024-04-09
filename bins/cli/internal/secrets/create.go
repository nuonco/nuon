package secrets

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Create(ctx context.Context, appID, name, value string, asJSON bool) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewCreateView("secret", asJSON)
	view.Start()
	view.Update("creating secret")

	secret, err := s.api.CreateAppSecret(ctx, appID, &models.ServiceCreateAppSecretRequest{
		Name:  &name,
		Value: &value,
	})
	if err != nil {
		view.Fail(err)
		return
	}

	view.Update(fmt.Sprintf("successfully created secret (%s)\n", secret.ID))
}
