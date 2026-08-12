package roles

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) EditRole(ctx context.Context, roleID, title, description string, contexts, permissions []string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	if title == "" && description == "" && contexts == nil && permissions == nil {
		return view.Error(&ui.CLIUserError{
			Msg: "nothing to update: pass at least one of --title, --description, --context, or --permission",
		})
	}

	req := &models.ServiceUpdateRoleRequest{
		Title:       title,
		Description: description,
		Contexts:    contexts,
	}

	if permissions != nil {
		entries, err := ParsePermissionEntries(permissions)
		if err != nil {
			return view.Error(&ui.CLIUserError{Msg: err.Error()})
		}
		req.Permissions = entries
	}

	role, err := s.api.UpdateRole(ctx, roleID, req)
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
