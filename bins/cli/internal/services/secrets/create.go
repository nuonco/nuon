package secrets

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Create(ctx context.Context, appID, name, value string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewCreateView("secret", asJSON)
	view.Start()
	view.Update("creating secret")

	secret, err := s.api.CreateAppSecret(ctx, appID, &models.ServiceCreateAppSecretRequest{
		Name:  &name,
		Value: &value,
	})
	if err != nil {
		return view.Fail(err)
	}

	view.Update(fmt.Sprintf("successfully created secret (%s)\n", secret.ID))
	return nil
}
