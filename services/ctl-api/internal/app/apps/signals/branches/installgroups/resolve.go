// Package installgroups resolves an install group to the installs it currently
// targets, either from its pinned install IDs or by evaluating its label selector.
package installgroups

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

type Resolved struct {
	InstallIDs []string
	GroupName  string
}

func Resolve(ctx workflow.Context, installGroupID, appBranchID string) (*Resolved, error) {
	logger := workflow.GetLogger(ctx)

	group, err := activities.AwaitGetInstallGroupByID(ctx, installGroupID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install group: %w", err)
	}

	if group.LabelSelector == nil {
		logger.Info("resolved install group",
			"install_group_id", group.ID,
			"install_group_name", group.Name,
			"install_count", len(group.InstallIDs),
		)
		return &Resolved{InstallIDs: group.InstallIDs, GroupName: group.Name}, nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, appBranchID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branch for label resolution: %w", err)
	}

	resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
		AppID:    branch.AppID,
		GroupID:  group.ID,
		Selector: group.LabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to resolve install group labels: %w", err)
	}

	logger.Info("resolved install group",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(resolved.InstallIDs),
		"resolved_via", "label_selector",
	)

	return &Resolved{InstallIDs: resolved.InstallIDs, GroupName: group.Name}, nil
}
