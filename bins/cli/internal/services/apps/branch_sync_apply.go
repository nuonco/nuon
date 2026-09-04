package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) applyBranchPlan(ctx context.Context, appID string, plan []branchPlanItem, result *BranchSyncResult) error {
	applyOrder := []branchOp{branchOpCreate, branchOpUpdate, branchOpDelete}
	indexByName := make(map[string]int, len(result.Branches))
	for i, item := range result.Branches {
		indexByName[item.Name] = i
	}

	mark := func(item branchPlanItem, status, branchID, configID, errMsg string) {
		idx, ok := indexByName[item.Name]
		if !ok {
			return
		}
		result.Branches[idx].Status = status
		if branchID != "" {
			result.Branches[idx].BranchID = branchID
		}
		if configID != "" {
			result.Branches[idx].ConfigID = configID
		}
		if errMsg != "" {
			result.Branches[idx].Error = errMsg
			result.Summary.Failed++
		}
	}

	for _, op := range applyOrder {
		for _, item := range plan {
			if item.Op != op {
				continue
			}
			switch op {
			case branchOpCreate:
				branch, cfg, err := s.createSyncedBranch(ctx, appID, item.Local)
				if err != nil {
					mark(item, "failed", "", "", err.Error())
					return fmt.Errorf("create branch %s: %w", item.Name, err)
				}
				configID := ""
				if cfg != nil {
					configID = cfg.ID
				}
				mark(item, "created", branch.ID, configID, "")
			case branchOpUpdate:
				cfg, err := s.updateSyncedBranch(ctx, appID, item.BranchID, item.Local)
				if err != nil {
					mark(item, "failed", item.BranchID, "", err.Error())
					return fmt.Errorf("update branch %s: %w", item.Name, err)
				}
				configID := ""
				if cfg != nil {
					configID = cfg.ID
				}
				mark(item, "updated", item.BranchID, configID, "")
			case branchOpDelete:
				if err := s.api.DeleteAppBranch(ctx, appID, item.BranchID); err != nil {
					mark(item, "failed", item.BranchID, "", err.Error())
					return fmt.Errorf("delete branch %s: %w", item.Name, err)
				}
				mark(item, "deleted", item.BranchID, "", "")
			}
		}
	}

	for _, item := range plan {
		if item.Op == branchOpUnchanged {
			mark(item, "unchanged", item.BranchID, "", "")
		}
	}
	return nil
}

func (s *Service) createSyncedBranch(ctx context.Context, appID string, cfg *config.AppBranchConfig) (*models.AppAppBranch, *models.AppAppBranchConfig, error) {
	name := cfg.Name
	branch, err := s.api.CreateAppBranch(ctx, appID, &models.ServiceCreateAppBranchRequest{
		Name:      &name,
		ManagedBy: appBranchManagedByConfig,
	})
	if err != nil {
		return nil, nil, err
	}

	configRow, err := s.writeBranchConfig(ctx, appID, branch.ID, cfg)
	if err != nil {
		return branch, nil, err
	}
	return branch, configRow, nil
}

func (s *Service) updateSyncedBranch(ctx context.Context, appID, branchID string, cfg *config.AppBranchConfig) (*models.AppAppBranchConfig, error) {
	return s.writeBranchConfig(ctx, appID, branchID, cfg)
}

func (s *Service) writeBranchConfig(ctx context.Context, appID, branchID string, cfg *config.AppBranchConfig) (*models.AppAppBranchConfig, error) {
	if cfg.ConnectedRepo == nil && cfg.PublicRepo == nil {
		if branchConfigNeedsRepo(cfg) {
			return nil, fmt.Errorf("branch %q sets install groups, preview, or post-deploy runbooks but has no connected_repo or public_repo", cfg.Name)
		}
		return nil, nil
	}

	resolver := newBranchNameResolver(s.api, appID)
	req, err := branchConfigRequest(ctx, resolver, cfg)
	if err != nil {
		return nil, err
	}
	return s.api.CreateAppBranchConfig(ctx, appID, branchID, req)
}

func branchConfigNeedsRepo(cfg *config.AppBranchConfig) bool {
	return len(cfg.InstallGroups) > 0 || cfg.Preview != nil || len(cfg.PostDeployRunbooks) > 0 || cfg.IgnoreChangesRegex != ""
}

func branchConfigRequest(ctx context.Context, resolver *branchNameResolver, cfg *config.AppBranchConfig) (*models.ServiceCreateAppBranchConfigRequest, error) {
	req := &models.ServiceCreateAppBranchConfigRequest{
		IgnoreChangesRegex:   generics.ToPtr(cfg.IgnoreChangesRegex),
		SendStatusesOnIgnore: generics.ToPtr(cfg.SendStatusesOnIgnore),
	}

	if cfg.ConnectedRepo != nil {
		req.ConnectedGithubVcsConfig = &models.HelpersConnectedGithubVCSConfigRequest{
			Repo:      generics.ToPtr(cfg.ConnectedRepo.Repo),
			Directory: generics.ToPtr(cfg.ConnectedRepo.Directory),
			Branch:    cfg.ConnectedRepo.Branch,
		}
	}
	if cfg.PublicRepo != nil {
		req.PublicGitVcsConfig = &models.HelpersPublicGitVCSConfigRequest{
			Repo:      generics.ToPtr(cfg.PublicRepo.Repo),
			Directory: generics.ToPtr(cfg.PublicRepo.Directory),
			Branch:    generics.ToPtr(cfg.PublicRepo.Branch),
		}
	}

	groups, err := installGroupRequests(ctx, resolver, cfg)
	if err != nil {
		return nil, err
	}
	req.InstallGroups = groups

	runbookIDs, err := resolvePostDeployRunbookIDs(ctx, resolver, cfg)
	if err != nil {
		return nil, err
	}
	req.PostDeployRunbookIds = runbookIDs

	preview, err := previewConfigRequest(ctx, resolver, cfg)
	if err != nil {
		return nil, err
	}
	req.PreviewConfig = preview

	return req, nil
}

func installGroupRequests(ctx context.Context, resolver *branchNameResolver, cfg *config.AppBranchConfig) ([]*models.ServiceInstallGroupRequest, error) {
	out := make([]*models.ServiceInstallGroupRequest, 0, len(cfg.InstallGroups))
	for i, group := range cfg.InstallGroups {
		order := int64(group.Order)
		if group.Order == 0 {
			order = int64(i)
		}
		req := &models.ServiceInstallGroupRequest{
			Name:                         generics.ToPtr(group.Name),
			Order:                        &order,
			AutoApproveOnPoliciesPassing: group.AutoApproveOnPoliciesPassing,
		}

		ids := append([]string{}, group.InstallIDs...)
		for _, name := range group.InstallNames {
			id, err := resolver.installID(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("install group %q: %w", group.Name, err)
			}
			ids = appendUnique(ids, id)
		}
		req.InstallIds = ids

		if len(group.LabelSelector) > 0 {
			req.LabelSelector = struct {
				models.GithubComNuoncoNuonPkgLabelsSelector
			}{
				GithubComNuoncoNuonPkgLabelsSelector: models.GithubComNuoncoNuonPkgLabelsSelector{
					MatchLabels: models.GithubComNuoncoNuonPkgLabelsLabels(group.LabelSelector),
				},
			}
		}

		out = append(out, req)
	}
	return out, nil
}

func resolvePostDeployRunbookIDs(ctx context.Context, resolver *branchNameResolver, cfg *config.AppBranchConfig) ([]string, error) {
	if len(cfg.PostDeployRunbooks) == 0 {
		return []string{}, nil
	}
	ids := make([]string, 0, len(cfg.PostDeployRunbooks))
	for _, name := range cfg.PostDeployRunbooks {
		id, err := resolver.runbookID(ctx, name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func previewConfigRequest(ctx context.Context, resolver *branchNameResolver, cfg *config.AppBranchConfig) (*models.AppAppBranchPreviewConfig, error) {
	if cfg.Preview == nil {
		return nil, nil
	}
	p := cfg.Preview
	out := &models.AppAppBranchPreviewConfig{
		Mode: models.AppAppBranchRunPreviewMode(p.Mode),
	}
	if out.Mode == "" {
		out.Mode = models.AppAppBranchRunPreviewModePlanDashOnly
	}
	if p.InstallID != "" {
		out.InstallID = p.InstallID
	}
	if p.InstallName != "" {
		id, err := resolver.installID(ctx, p.InstallName)
		if err != nil {
			return nil, fmt.Errorf("branch %q preview: %w", cfg.Name, err)
		}
		out.InstallID = id
		out.InstallName = p.InstallName
	}
	if len(p.LabelSelector) > 0 {
		out.LabelSelector = &models.GithubComNuoncoNuonPkgLabelsSelector{
			MatchLabels: models.GithubComNuoncoNuonPkgLabelsLabels(p.LabelSelector),
		}
	}
	if p.SetStatuses != nil {
		out.SetStatuses = *p.SetStatuses
	} else {
		out.SetStatuses = true
	}
	if p.Comment != nil {
		out.Comment = *p.Comment
	} else {
		out.Comment = true
	}
	return out, nil
}

func appendUnique(ids []string, id string) []string {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}
