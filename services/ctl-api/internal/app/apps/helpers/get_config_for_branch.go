package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

var ErrAppBranchNotFound = errors.New("app branch not found")

func (h *Helpers) GetLatestActiveAppConfigForBranch(ctx context.Context, appID, branchID string) (*app.AppConfig, error) {
	var branch app.AppBranch
	if err := h.db.WithContext(ctx).
		Where(app.AppBranch{ID: branchID, AppID: appID}).
		First(&branch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("%w: %s", ErrAppBranchNotFound, branchID),
				Description: "branch does not belong to this app",
			}
		}
		return nil, errors.Wrap(err, "unable to get app branch")
	}

	var appConfig app.AppConfig
	res := h.db.WithContext(ctx).
		Scopes(ActiveAppConfigs(appID)).
		Where("app_branch_id = ?", branchID).
		Order("created_at DESC").
		First(&appConfig)
	if res.Error == nil {
		return &appConfig, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(res.Error, "unable to get app config")
	}

	return h.GetLatestActiveAppConfigBare(ctx, appID)
}
