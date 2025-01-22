package installs

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) UpdateInput(ctx context.Context, installID, name, value string) error {
	installInput, err := s.api.UpdateInstallInput(ctx, installID, &models.ServiceUpdateInstallInputRequest{
		Name:  &name,
		Value: &value,
	})
	if err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintJSON(installInput)
	return nil
}
