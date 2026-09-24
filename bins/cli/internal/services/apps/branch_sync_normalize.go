package apps

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func canonicalizeLocalBranch(ctx context.Context, resolver *branchNameResolver, in *config.AppBranchConfig) (*config.AppBranchConfig, error) {
	out := cloneAppBranchConfig(in)
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
	out.Run = normalizeRunConfig(out.Run)
	out.InstallGroups = withDefaultInstallGroup(out)
	return out, nil
}

// withDefaultInstallGroup mirrors the server, which seeds a single default group
// on any config written without one. Without this a branch that declares no
// install_groups compares unequal to the group the server stored and every sync
// reports drift no update can settle. A branch with no repo has no config
// written at all, so it keeps an empty list.
func withDefaultInstallGroup(cfg *config.AppBranchConfig) []config.AppBranchInstallGroupConfig {
	if len(cfg.InstallGroups) > 0 {
		return cfg.InstallGroups
	}
	if cfg.ConnectedRepo == nil && cfg.PublicRepo == nil {
		return cfg.InstallGroups
	}
	return []config.AppBranchInstallGroupConfig{{
		Name:    defaultInstallGroupName,
		Order:   0,
		Default: true,
	}}
}

func normalizeRemoteBranch(ctx context.Context, resolver *branchNameResolver, name string, latest *models.AppAppBranchConfig) (*config.AppBranchConfig, error) {
	out := &config.AppBranchConfig{Name: name, Run: normalizeRunConfig(nil)}
	if latest == nil {
		return out, nil
	}

	if latest.RunConfig != nil {
		out.Run = normalizeRunConfig(&config.AppBranchRunConfig{
			Mode:        string(latest.RunConfig.Mode),
			TagPrefix:   latest.RunConfig.TagPrefix,
			GithubLabel: latest.RunConfig.GithubLabel,
		})
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
			Default:                      group.Default,
			AutoApproveOnPoliciesPassing: group.AutoApproveOnPoliciesPassing,
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
		preview.IgnoreDrafts = generics.ToPtr(p.IgnoreDrafts)
		preview.React = generics.ToPtr(p.React)
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
	if p.IgnoreDrafts == nil {
		p.IgnoreDrafts = generics.ToPtr(true)
	}
	if p.React == nil {
		p.React = generics.ToPtr(true)
	}
}

// normalizeRunConfig returns a copy with the server's mode aliases and default
// applied. An absent run config is written as push, so it compares equal to one.
func normalizeRunConfig(in *config.AppBranchRunConfig) *config.AppBranchRunConfig {
	out := &config.AppBranchRunConfig{}
	if in != nil {
		*out = *in
	}
	switch out.Mode {
	case "", "all":
		out.Mode = string(models.AppAppBranchRunModePush)
	case "on_tag_prefix":
		out.Mode = string(models.AppAppBranchRunModeOnTag)
	}
	return out
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
		if in.Preview.IgnoreDrafts != nil {
			preview.IgnoreDrafts = generics.ToPtr(*in.Preview.IgnoreDrafts)
		}
		if in.Preview.React != nil {
			preview.React = generics.ToPtr(*in.Preview.React)
		}
		out.Preview = &preview
	}
	if in.InstallGroups != nil {
		out.InstallGroups = make([]config.AppBranchInstallGroupConfig, len(in.InstallGroups))
		for i, group := range in.InstallGroups {
			g := group
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
