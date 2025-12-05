package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) CreateAppBranch(ctx context.Context, orgID, appID, name string, connectedGithubVCSConfigID string) (*app.AppBranch, error) {
	branch := app.AppBranch{
		OrgID:                      orgID,
		AppID:                      appID,
		Name:                       name,
		ConnectedGithubVCSConfigID: connectedGithubVCSConfigID,
	}

	res := h.db.WithContext(ctx).Create(&branch)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app branch: %w", res.Error)
	}

	return &branch, nil
}
