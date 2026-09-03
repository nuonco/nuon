package activities

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ResolveInstallGroupInstallsInput struct {
	AppID      string           `json:"app_id"`
	GroupID    string           `json:"group_id"`
	InstallIDs []string         `json:"install_ids,omitempty"`
	Selector   *labels.Selector `json:"selector"`

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
	var candidateIDs []string
	resolveByIDs := input.AllInstalls || input.InstallIDs != nil
	if input.AllInstalls {
		ids, err := a.helpers.ResolveAllInstallsForBranch(ctx, input.AppID, input.AppBranchID)
		if err != nil {
			return nil, err
		}
		candidateIDs = ids
	} else if len(input.InstallIDs) > 0 {
		candidateIDs = input.InstallIDs
	}

	var installs []app.Install
	query := a.db.WithContext(ctx).Preload("OperatingModel")
	if resolveByIDs {
		query = query.Where("id IN ?", candidateIDs)
	} else {
		query = query.
			Where(app.Install{AppID: input.AppID}).
			Scopes(labels.WithLabels("labels", input.Selector.MatchLabels))
	}
	if err := query.Find(&installs).Error; err != nil {
		return nil, err
	}

	allowed := make(map[string]bool, len(installs))
	for _, install := range installs {
		if install.AppBranchUpdateEligible() {
			allowed[install.ID] = true
		}
	}
	ids := make([]string, 0, len(installs))
	if resolveByIDs {
		for _, id := range candidateIDs {
			if allowed[id] {
				ids = append(ids, id)
			}
		}
	} else {
		for _, install := range installs {
			if allowed[install.ID] {
				ids = append(ids, install.ID)
			}
		}
	}

	resolvedVia := "label_selector"
	if input.AllInstalls {
		resolvedVia = "all_installs"
	} else if input.InstallIDs != nil {
		resolvedVia = "install_ids"
	}
	a.l.Info("resolved install group",
		zap.String("group_id", input.GroupID),
		zap.String("resolved_via", resolvedVia),
		zap.Int("resolved_count", len(ids)),
	)

	return &ResolveInstallGroupInstallsOutput{InstallIDs: ids}, nil
}
