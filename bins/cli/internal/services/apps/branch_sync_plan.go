package apps

import (
	"fmt"
	"sort"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type branchOp string

const (
	branchOpCreate    branchOp = "create"
	branchOpUpdate    branchOp = "update"
	branchOpDelete    branchOp = "delete"
	branchOpUnchanged branchOp = "unchanged"
)

type branchPlanItem struct {
	Name      string
	Op        branchOp
	BranchID  string
	ManagedBy string
	Local     *config.AppBranchConfig
	Remote    *config.AppBranchConfig
	Diff      *diff.Diff
}

func buildBranchSyncPlan(
	local []*config.AppBranchConfig,
	remotes []*models.AppAppBranch,
	remoteCfg map[string]*config.AppBranchConfig,
	directory bool,
) ([]branchPlanItem, error) {
	localByName := make(map[string]*config.AppBranchConfig, len(local))
	localNames := make([]string, 0, len(local))
	for _, cfg := range local {
		if _, ok := localByName[cfg.Name]; ok {
			return nil, &ui.CLIUserError{Msg: fmt.Sprintf("duplicate branch name %q", cfg.Name)}
		}
		localByName[cfg.Name] = cfg
		localNames = append(localNames, cfg.Name)
	}
	sort.Strings(localNames)

	remoteByName := make(map[string]*models.AppAppBranch, len(remotes))
	for _, remote := range remotes {
		remoteByName[remote.Name] = remote
	}

	var plan []branchPlanItem
	for _, name := range localNames {
		cfg := localByName[name]
		remote, exists := remoteByName[name]
		if !exists {
			plan = append(plan, branchPlanItem{
				Name:  name,
				Op:    branchOpCreate,
				Local: cfg,
				Diff:  cfg.Diff(nil),
			})
			continue
		}

		managedBy := remote.ManagedBy
		if managedBy == "" {
			managedBy = appBranchManagedByManually
		}
		if managedBy != appBranchManagedByConfig {
			return nil, &ui.CLIUserError{
				Msg: fmt.Sprintf("branch %q is managed manually and cannot be synced from config; inspect it with `nuon branches get -b %s`, or remove the local file", name, name),
			}
		}

		item := branchPlanItem{
			Name:      name,
			BranchID:  remote.ID,
			ManagedBy: managedBy,
			Local:     cfg,
			Remote:    remoteCfg[name],
			Diff:      cfg.Diff(remoteCfg[name]),
		}
		if item.Diff != nil && item.Diff.Summary().HasChanged {
			item.Op = branchOpUpdate
		} else {
			item.Op = branchOpUnchanged
			item.Diff = nil
		}
		plan = append(plan, item)
	}

	if directory {
		var remoteNames []string
		for name := range remoteByName {
			remoteNames = append(remoteNames, name)
		}
		sort.Strings(remoteNames)
		for _, name := range remoteNames {
			if _, ok := localByName[name]; ok {
				continue
			}
			remote := remoteByName[name]
			if remote.ManagedBy != appBranchManagedByConfig {
				continue
			}
			plan = append(plan, branchPlanItem{
				Name:      name,
				Op:        branchOpDelete,
				BranchID:  remote.ID,
				ManagedBy: remote.ManagedBy,
				Remote:    remoteCfg[name],
			})
		}
	}

	return plan, nil
}
