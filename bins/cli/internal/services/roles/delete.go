package roles

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) DeleteRole(ctx context.Context, roleID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	if err := s.api.DeleteRole(ctx, roleID); err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]string{"id": roleID, "status": "deleted"})
		return nil
	}

	ui.PrintLn(fmt.Sprintf("deleted role %s", roleID))
	return nil
}
