package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) EnsureInstallComponents(ctx context.Context, appID string, installIDs []string) error {
	// fetch the parent app's installs and ensure each gets the new component
	parentApp := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Installs").
		Preload("Components").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return fmt.Errorf("unable to get app: %w", res.Error)
	}

	// if install IDs are not passed in, then update all installs
	if len(installIDs) < 1 {
		for _, install := range parentApp.Installs {
			installIDs = append(installIDs, install.ID)
		}
	}

	// create an install component for all known installs
	installCmps := make([]app.InstallComponent, 0)
	for _, installID := range installIDs {
		for _, component := range parentApp.Components {
			installCmps = append(installCmps, app.InstallComponent{
				ComponentID: component.ID,
				InstallID:   installID,
			})
		}
	}

	if len(installCmps) < 1 {
		return nil
	}

	res = h.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&installCmps)
	if res.Error != nil {
		return fmt.Errorf("unable to create install components: %w", res.Error)
	}

	res = h.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(h.TFWorkspacesFromICs(installCmps))
	if res.Error != nil {
		return fmt.Errorf("unable to create terraform workspaces: %w", res.Error)
	}

	return nil
}

func (h *Helpers) TFWorkSpaceFromIC(ic app.InstallComponent) app.TerraformWorkspace {
	return app.TerraformWorkspace{
		OrgID:     ic.OrgID,
		OwnerID:   ic.InstallID,
		OwnerType: "install_components",
	}
}

func (h *Helpers) TFWorkspacesFromICs(ics []app.InstallComponent) []app.TerraformWorkspace {
	workspaces := make([]app.TerraformWorkspace, 0)
	for _, ic := range ics {
		workspaces = append(workspaces, h.TFWorkSpaceFromIC(ic))
	}
	return workspaces
}
