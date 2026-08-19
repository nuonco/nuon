package activities

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ResolveInstallGroupInstallsInput struct {
	AppID    string           `json:"app_id"`
	GroupID  string           `json:"group_id"`
	Selector *labels.Selector `json:"selector"`

	// AllInstalls resolves every install on the app that no other branch owns,
	// ignoring Selector. AppBranchID is the owning branch.
	AllInstalls bool   `json:"all_installs,omitempty"`
	AppBranchID string `json:"app_branch_id,omitempty"`
}

type ResolveInstallGroupInstallsOutput struct {
	InstallIDs []string `json:"install_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ResolveInstallGroupInstalls(ctx context.Context, input *ResolveInstallGroupInstallsInput) (*ResolveInstallGroupInstallsOutput, error) {
	if input.AllInstalls {
		ids, err := a.helpers.ResolveAllInstallsForBranch(ctx, input.AppID, input.AppBranchID)
		if err != nil {
			return nil, err
		}

		a.l.Info("resolved install group",
			zap.String("group_id", input.GroupID),
			zap.String("resolved_via", "all_installs"),
			zap.Int("resolved_count", len(ids)),
		)

		return &ResolveInstallGroupInstallsOutput{InstallIDs: ids}, nil
	}

	var installs []app.Install
	if err := a.db.WithContext(ctx).
		Where(app.Install{AppID: input.AppID}).
		Scopes(labels.WithLabels("labels", input.Selector.MatchLabels)).
		Find(&installs).Error; err != nil {
		return nil, err
	}

	ids := make([]string, len(installs))
	for i, inst := range installs {
		ids[i] = inst.ID
	}

	a.l.Info("resolved install group",
		zap.String("group_id", input.GroupID),
		zap.String("resolved_via", "label_selector"),
		zap.Int("resolved_count", len(ids)),
	)

	return &ResolveInstallGroupInstallsOutput{InstallIDs: ids}, nil
}
