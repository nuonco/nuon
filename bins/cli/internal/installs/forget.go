package installs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Forget(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	res, err := s.api.ForgetInstall(ctx, installID)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	if asJSON {
		type response struct {
			ID        string `json:"id"`
			Forgotten bool   `json:"forgotten"`
		}
		r := response{ID: installID, Forgotten: res}
		ui.PrintJSON(r)
		return nil
	}

	ui.PrintLn(fmt.Sprintf("forget install %s request sent.", installID))

	return nil
}
