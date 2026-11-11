package helpers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func branchSelectsLabels(groups []app.AppBranchInstallGroup, lbls labels.Labels) bool {
	for i := range groups {
		selector := groups[i].LabelSelector
		if selector == nil || len(selector.MatchLabels) == 0 {
			continue
		}
		if selector.Matches(lbls) {
			return true
		}
	}
	return false
}

// BranchForInstallLabels returns the app branch whose install groups select
// lbls, or nil when no branch matches.
func (h *Helpers) BranchForInstallLabels(ctx context.Context, appID string, lbls labels.Labels) (*app.AppBranch, error) {
	return BranchForInstallLabelsWithDB(ctx, h.db, appID, lbls)
}

func BranchForInstallLabelsWithDB(ctx context.Context, db *gorm.DB, appID string, lbls labels.Labels) (*app.AppBranch, error) {
	if len(lbls) == 0 {
		return nil, nil
	}

	var branches []app.AppBranch
	if err := db.WithContext(ctx).
		Where(app.AppBranch{AppID: appID}).
		Order("name").
		Find(&branches).Error; err != nil {
		return nil, fmt.Errorf("unable to list app branches for app %s: %w", appID, err)
	}

	var matched []app.AppBranch
	for i := range branches {
		groups, err := LatestConfigInstallGroupsWithDB(ctx, db, branches[i].ID)
		if err != nil {
			return nil, err
		}
		if branchSelectsLabels(groups, lbls) {
			matched = append(matched, branches[i])
		}
	}

	switch len(matched) {
	case 0:
		return nil, nil
	case 1:
		return &matched[0], nil
	default:
		names := make([]string, len(matched))
		for i := range matched {
			names[i] = matched[i].Name
		}
		sort.Strings(names)
		joined := strings.Join(names, ", ")
		return nil, stderr.ErrUser{
			Err: fmt.Errorf("install labels match app branches %s", joined),
			Description: fmt.Sprintf(
				"The install's labels match more than one app branch (%s). Set app_branch on the install to choose one.",
				joined,
			),
		}
	}
}
