package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		resp, err := s.api.DeleteInstall(ctx, installID)
		if err != nil {
			return ui.PrintJSONError(err)
		}
		type response struct {
			ID         string `json:"id"`
			Deleted    bool   `json:"deleted"`
			WorkflowID string `json:"workflow_id,omitempty"`
		}
		r := response{ID: installID, Deleted: resp != nil}
		if resp != nil {
			r.WorkflowID = resp.WorkflowID
		}
		ui.PrintJSON(r)
		return nil
	}

	view := ui.NewDeleteView("install", installID, s.cfg.Interactive)
	view.Start()
	_, err = s.api.DeleteInstall(ctx, installID)
	if err != nil {
		return view.Fail(err)
	}

	// unset install_id if it is the currentInstallID
	currentInstallID := s.GetInstallID()

	if installID == currentInstallID {
		if err := s.unsetInstallID(ctx); err != nil {
			return view.Fail(err)
		}
	}

	view.SuccessQueued()
	return nil
}
