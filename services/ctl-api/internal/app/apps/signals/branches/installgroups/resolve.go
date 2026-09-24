// Package installgroups resolves an install group to the installs it currently
// targets, either from its pinned install IDs or by evaluating its label selector.
package installgroups

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

type Resolved struct {
	InstallIDs []string
	GroupName  string
}

func ResolvePreviewTarget(
	ctx workflow.Context,
	appBranchID string,
	installID string,
	selector *labels.Selector,
) (*Resolved, error) {
	hasInstallID := installID != ""
	hasSelector := selector != nil && len(selector.MatchLabels) > 0
	if hasInstallID == hasSelector {
		return nil, fmt.Errorf("preview requires exactly one of install_id or label_selector")
	}
	if hasInstallID {
		resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
			InstallIDs:  []string{installID},
			AppBranchID: appBranchID,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to resolve preview install: %w", err)
		}
		return &Resolved{InstallIDs: resolved.InstallIDs, GroupName: "preview"}, nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, appBranchID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branch for preview label resolution: %w", err)
	}
	resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
		AppID:       branch.AppID,
		Selector:    selector,
		AppBranchID: appBranchID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to resolve preview labels: %w", err)
	}

	return &Resolved{InstallIDs: resolved.InstallIDs, GroupName: "preview"}, nil
}

func Resolve(ctx workflow.Context, installGroupID, appBranchID string) (*Resolved, error) {
	logger := workflow.GetLogger(ctx)

	group, err := activities.AwaitGetInstallGroupByID(ctx, installGroupID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install group: %w", err)
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, appBranchID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branch for label resolution: %w", err)
	}

	// Explicit IDs go through the same activity as selectors and all-installs so
	// they are filtered to the installs the branch owns. A group listing an
	// install that has since moved to another branch resolves without it.
	resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
		AppID:       branch.AppID,
		GroupID:     group.ID,
		InstallIDs:  group.InstallIDs,
		Selector:    group.LabelSelector,
		AllInstalls: group.AllInstalls,
		AppBranchID: appBranchID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to resolve install group: %w", err)
	}

	resolvedVia := "install_ids"
	switch {
	case group.AllInstalls:
		resolvedVia = "all_installs"
	case group.LabelSelector != nil:
		resolvedVia = "label_selector"
	}

	logger.Info("resolved install group",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(resolved.InstallIDs),
		"resolved_via", resolvedVia,
	)

	return &Resolved{InstallIDs: resolved.InstallIDs, GroupName: group.Name}, nil
}
