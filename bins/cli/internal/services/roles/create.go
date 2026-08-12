package roles

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateRole(ctx context.Context, title, description string, contexts, permissions []string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	entries, err := ParsePermissionEntries(permissions)
	if err != nil {
		return view.Error(&ui.CLIUserError{Msg: err.Error()})
	}

	role, err := s.api.CreateRole(ctx, &models.ServiceCreateRoleRequest{
		Title:       &title,
		Description: description,
		Contexts:    contexts,
		Permissions: entries,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(role)
		return nil
	}

	view.Render(roleDetail(role))
	return nil
}
