package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type InstallRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	AppID string `json:"app_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @by-field appBranchID
func (a *Activities) getInstallsForAppBranch(ctx context.Context, appBranchID string) ([]InstallRef, error) {
	var branch app.AppBranch
	if err := a.db.WithContext(ctx).First(&branch, "id = ?", appBranchID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", err)
	}

	var installs []app.Install
	if err := a.db.WithContext(ctx).
		Where(app.Install{AppID: branch.AppID}).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to get installs: %w", err)
	}

	refs := make([]InstallRef, len(installs))
	for i, inst := range installs {
		refs[i] = InstallRef{
			ID:    inst.ID,
			Name:  inst.Name,
			AppID: inst.AppID,
		}
	}

	return refs, nil
}
