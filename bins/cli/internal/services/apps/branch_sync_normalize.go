package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func canonicalizeLocalBranch(ctx context.Context, resolver *branchNameResolver, in *config.AppBranchConfig) (*config.AppBranchConfig, error) {
	out := cloneAppBranchConfig(in)
	for i, group := range out.InstallGroups {
		names := append([]string{}, group.InstallNames...)
		for _, id := range group.InstallIDs {
			name, err := resolver.installName(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("install group %q: %w", group.Name, err)
			}
			if name == "" {
				name = id
			}
			names = appendUnique(names, name)
		}
		out.InstallGroups[i].InstallIDs = nil
		out.InstallGroups[i].InstallNames = names
	}
	if out.Preview != nil {
		if out.Preview.InstallID != "" && out.Preview.InstallName == "" {
			name, err := resolver.installName(ctx, out.Preview.InstallID)
			if err != nil {
				return nil, err
			}
			if name != "" {
				out.Preview.InstallName = name
				out.Preview.InstallID = ""
			}
		}
		normalizePreviewDefaults(out.Preview)
	}
	return out, nil
}

func normalizeRemoteBranch(ctx context.Context, resolver *branchNameResolver, name string, latest *models.AppAppBranchConfig) (*config.AppBranchConfig, error) {
	out := &config.AppBranchConfig{Name: name}
	if latest == nil {
		return out, nil
	}

	if latest.ConnectedGithubVcsConfig != nil {
		out.ConnectedRepo = &config.ConnectedRepoConfig{
			Repo:      latest.ConnectedGithubVcsConfig.Repo,
			Directory: latest.ConnectedGithubVcsConfig.Directory,
			Branch:    latest.ConnectedGithubVcsConfig.Branch,
		}
	}
	if latest.PublicGitVcsConfig != nil {
		out.PublicRepo = &config.PublicRepoConfig{
			Repo:      latest.PublicGitVcsConfig.Repo,
			Directory: latest.PublicGitVcsConfig.Directory,
			Branch:    latest.PublicGitVcsConfig.Branch,
		}
	}

	for _, group := range latest.InstallGroups {
		if group == nil {
			continue
		}
		cfg := config.AppBranchInstallGroupConfig{
			Name:                         group.Name,
			Order:                        int(group.Order),
			AutoApproveOnPoliciesPassing: group.AutoApproveOnPoliciesPassing,
		}
		for _, id := range group.InstallIds {
			name, err := resolver.installName(ctx, id)
			if err != nil {
				return nil, err
			}
			if name == "" {
				name = id
			}
			cfg.InstallNames = append(cfg.InstallNames, name)
		}
		if group.LabelSelector != nil && len(group.LabelSelector.MatchLabels) > 0 {
			cfg.LabelSelector = map[string]string(group.LabelSelector.MatchLabels)
		}
		out.InstallGroups = append(out.InstallGroups, cfg)
	}

	for _, id := range latest.PostDeployRunbookIds {
		out.PostDeployRunbooks = append(out.PostDeployRunbooks, resolver.runbookName(ctx, id))
	}

	out.IgnoreChangesRegex = latest.IgnoreChangesRegex
	out.SendStatusesOnIgnore = latest.SendStatusesOnIgnore

	if latest.PreviewConfig != nil {
		p := latest.PreviewConfig
		preview := &config.AppBranchPreviewConfig{
			Mode: string(p.Mode),
		}
		if p.InstallName != "" {
			preview.InstallName = p.InstallName
		} else if p.InstallID != "" {
			name, err := resolver.installName(ctx, p.InstallID)
			if err != nil {
				return nil, err
			}
			if name != "" {
				preview.InstallName = name
			} else {
				preview.InstallID = p.InstallID
			}
		}
		if p.LabelSelector != nil && len(p.LabelSelector.MatchLabels) > 0 {
			preview.LabelSelector = map[string]string(p.LabelSelector.MatchLabels)
		}
		preview.SetStatuses = generics.ToPtr(p.SetStatuses)
		preview.Comment = generics.ToPtr(p.Comment)
		normalizePreviewDefaults(preview)
		out.Preview = preview
	}

	return out, nil
}

func normalizePreviewDefaults(p *config.AppBranchPreviewConfig) {
	if p.Mode == "" {
		p.Mode = "plan-only"
	}
	if p.SetStatuses == nil {
		p.SetStatuses = generics.ToPtr(true)
	}
	if p.Comment == nil {
		p.Comment = generics.ToPtr(true)
	}
}

func cloneAppBranchConfig(in *config.AppBranchConfig) *config.AppBranchConfig {
	if in == nil {
		return &config.AppBranchConfig{}
	}
	out := *in
	if in.ConnectedRepo != nil {
		repo := *in.ConnectedRepo
		out.ConnectedRepo = &repo
	}
	if in.PublicRepo != nil {
		repo := *in.PublicRepo
		out.PublicRepo = &repo
	}
	if in.Preview != nil {
		preview := *in.Preview
		if in.Preview.LabelSelector != nil {
			preview.LabelSelector = copyStringMap(in.Preview.LabelSelector)
		}
		if in.Preview.SetStatuses != nil {
			preview.SetStatuses = generics.ToPtr(*in.Preview.SetStatuses)
		}
		if in.Preview.Comment != nil {
			preview.Comment = generics.ToPtr(*in.Preview.Comment)
		}
		out.Preview = &preview
	}
	if in.InstallGroups != nil {
		out.InstallGroups = make([]config.AppBranchInstallGroupConfig, len(in.InstallGroups))
		for i, group := range in.InstallGroups {
			g := group
			g.InstallIDs = append([]string{}, group.InstallIDs...)
			g.InstallNames = append([]string{}, group.InstallNames...)
			g.LabelSelector = copyStringMap(group.LabelSelector)
			if group.AutoApproveOnPoliciesPassing != nil {
				g.AutoApproveOnPoliciesPassing = generics.ToPtr(*group.AutoApproveOnPoliciesPassing)
			}
			out.InstallGroups[i] = g
		}
	}
	out.PostDeployRunbooks = append([]string{}, in.PostDeployRunbooks...)
	return &out
}

func copyStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
