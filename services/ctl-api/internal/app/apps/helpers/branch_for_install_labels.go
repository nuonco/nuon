package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func branchSelectsLabels(groups []app.AppBranchInstallGroup, lbls labels.Labels) bool {
	return false
}

// BranchForInstallLabels returns the app branch whose install groups select
// lbls, or nil when no branch matches.
func (h *Helpers) BranchForInstallLabels(ctx context.Context, appID string, lbls labels.Labels) (*app.AppBranch, error) {
	return BranchForInstallLabelsWithDB(ctx, h.db, appID, lbls)
}

func BranchForInstallLabelsWithDB(ctx context.Context, db *gorm.DB, appID string, lbls labels.Labels) (*app.AppBranch, error) {
	return nil, nil
}
