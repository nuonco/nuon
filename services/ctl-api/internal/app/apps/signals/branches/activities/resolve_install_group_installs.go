package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type ResolveInstallGroupInstallsInput struct {
	AppID    string           `json:"app_id"`
	GroupID  string           `json:"group_id"`
	Selector *labels.Selector `json:"selector"`

	// InstallIDs is the group's explicit membership list.
	InstallIDs []string `json:"install_ids,omitempty"`

	// AllInstalls resolves every install the branch owns, ignoring Selector and
	// InstallIDs.
	AllInstalls bool `json:"all_installs,omitempty"`

	// AppBranchID is the branch the group belongs to. Install groups resolve
	// within the installs that branch owns; AppScoped callers use it only for
	// logging/validation context.
	AppBranchID string `json:"app_branch_id,omitempty"`

	// AppScoped resolves against every install on AppID instead of the
	// installs AppBranchID owns. Preview targets are the only caller: a
	// preview can point at an install on another branch (e.g. main), and the
	// preview-candidate picker offers exactly those installs, so applying the
	// preview must be able to reach them too.
	AppScoped bool `json:"app_scoped,omitempty"`
}

type ResolveInstallGroupInstallsOutput struct {
	InstallIDs []string `json:"install_ids"`
}

// ResolveInstallGroupInstalls turns an install group into the installs it
// deploys to. All three targeting modes start from the installs the branch
// owns, so a group can never reach an install that belongs to another branch,
// whatever its selector or ID list says. AppScoped callers (preview targets)
// are the one exception: they start from every install on the app instead.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ResolveInstallGroupInstalls(ctx context.Context, input *ResolveInstallGroupInstallsInput) (*ResolveInstallGroupInstallsOutput, error) {
	if input.AppBranchID == "" {
		return nil, fmt.Errorf("app_branch_id is required to resolve install group %s", input.GroupID)
	}

	var owned []app.Install
	var err error
	if input.AppScoped {
		if input.AppID == "" {
			return nil, fmt.Errorf("app_id is required to resolve an app-scoped preview target")
		}
		owned, err = a.helpers.AppInstalls(ctx, input.AppID)
	} else {
		owned, err = a.helpers.BranchInstalls(ctx, input.AppBranchID)
	}
	if err != nil {
		return nil, err
	}

	group := &app.AppBranchInstallGroup{
		InstallIDs:    input.InstallIDs,
		LabelSelector: input.Selector,
		AllInstalls:   input.AllInstalls,
	}

	// A config save already rejects an install two groups target, but a config
	// can go stale against installs that moved since, so the run refuses rather
	// than deploying the same install from two groups.
	siblings, err := a.siblingInstallGroups(ctx, input.GroupID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(owned))
	for i := range owned {
		install := &owned[i]
		if !appshelpers.InstallMatchesGroup(group, install) {
			continue
		}
		if err := appshelpers.ValidateInstallSingleGroup(siblings, install); err != nil {
			return nil, err
		}
		ids = append(ids, install.ID)
	}

	resolvedVia := "install_ids"
	switch {
	case input.AllInstalls:
		resolvedVia = "all_installs"
	case input.Selector != nil:
		resolvedVia = "label_selector"
	}

	a.l.Info("resolved install group",
		zap.String("group_id", input.GroupID),
		zap.String("app_branch_id", input.AppBranchID),
		zap.String("resolved_via", resolvedVia),
		zap.Int("branch_install_count", len(owned)),
		zap.Int("resolved_count", len(ids)),
	)

	return &ResolveInstallGroupInstallsOutput{InstallIDs: ids}, nil
}

// siblingInstallGroups returns every group on the same config as groupID,
// including groupID itself. Callers without a group (preview targets) get none.
func (a *Activities) siblingInstallGroups(ctx context.Context, groupID string) ([]app.AppBranchInstallGroup, error) {
	if groupID == "" {
		return nil, nil
	}

	var group app.AppBranchInstallGroup
	if err := a.db.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install group %s: %w", groupID, err)
	}

	var groups []app.AppBranchInstallGroup
	if err := a.db.WithContext(ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: group.AppBranchConfigID}).
		Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("unable to get install groups for config %s: %w", group.AppBranchConfigID, err)
	}

	return groups, nil
}
