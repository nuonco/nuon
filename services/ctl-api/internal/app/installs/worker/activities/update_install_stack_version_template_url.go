package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateInstallStackVersionTemplateURLRequest struct {
	ID           string `validate:"required"`
	TemplateURL  string `validate:"required"`
	QuickLinkURL string
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateInstallStackVersionTemplateURL(ctx context.Context, req *UpdateInstallStackVersionTemplateURLRequest) error {
	updates := map[string]interface{}{
		"template_url": req.TemplateURL,
	}
	if req.QuickLinkURL != "" {
		updates["quick_link_url"] = req.QuickLinkURL
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{}).
		Where("id = ?", req.ID).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to update template URL: %w", res.Error)
	}
	return nil
}
