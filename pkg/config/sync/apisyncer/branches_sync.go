package apisyncer

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) syncAppBranches(ctx context.Context, resource string) error {
	branches := s.getAllBranches()
	if len(branches) == 0 {
		return nil
	}

	existingBranches, err := s.apiClient.GetAppBranches(ctx, s.appID)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      fmt.Errorf("unable to list app branches: %w", err),
		}
	}

	existingByName := make(map[string]*models.AppAppBranch, len(existingBranches))
	for _, b := range existingBranches {
		existingByName[b.Name] = b
	}

	for _, branchCfg := range branches {
		if err := s.syncSingleBranch(ctx, resource, branchCfg, existingByName); err != nil {
			return err
		}
	}

	return nil
}

func (s *syncer) syncSingleBranch(ctx context.Context, resource string, branchCfg *config.AppBranchConfig, existingByName map[string]*models.AppAppBranch) error {
	existing, found := existingByName[branchCfg.Name]

	var branchID string
	if !found {
		branch, err := s.apiClient.CreateAppBranch(ctx, s.appID, &models.ServiceCreateAppBranchRequest{
			Name: generics.ToPtr(branchCfg.Name),
		})
		if err != nil {
			return sync.SyncAPIErr{
				Resource: resource,
				Err:      fmt.Errorf("unable to create app branch %q: %w", branchCfg.Name, err),
			}
		}
		branchID = branch.ID
	} else {
		branchID = existing.ID
	}

	if branchCfg.ConnectedRepo == nil && branchCfg.PublicRepo == nil {
		return nil
	}

	var nameToID map[string]string
	for _, group := range branchCfg.InstallGroups {
		if len(group.InstallNames) > 0 {
			var err error
			nameToID, err = s.resolveInstallNames(ctx)
			if err != nil {
				return sync.SyncAPIErr{
					Resource: resource,
					Err:      fmt.Errorf("unable to list installs for name resolution: %w", err),
				}
			}
			break
		}
	}

	configReq := &models.ServiceCreateAppBranchConfigRequest{}

	if branchCfg.ConnectedRepo != nil {
		configReq.ConnectedGithubVcsConfig = &models.HelpersConnectedGithubVCSConfigRequest{
			Repo:      generics.ToPtr(branchCfg.ConnectedRepo.Repo),
			Branch:    branchCfg.ConnectedRepo.Branch,
			Directory: generics.ToPtr(branchCfg.ConnectedRepo.Directory),
		}
	}

	if branchCfg.PublicRepo != nil {
		configReq.PublicGitVcsConfig = &models.HelpersPublicGitVCSConfigRequest{
			Repo:      generics.ToPtr(branchCfg.PublicRepo.Repo),
			Branch:    generics.ToPtr(branchCfg.PublicRepo.Branch),
			Directory: generics.ToPtr(branchCfg.PublicRepo.Directory),
		}
	}

	for i, group := range branchCfg.InstallGroups {
		order := int64(group.Order)
		if order == 0 {
			order = int64(i)
		}

		igReq := &models.ServiceInstallGroupRequest{
			Name:           generics.ToPtr(group.Name),
			Order:          &order,
			UseForPreviews: group.UseForPreviews,
			InstallIds:     group.InstallIDs,
		}

		if len(group.InstallNames) > 0 {
			seen := make(map[string]bool, len(igReq.InstallIds))
			for _, id := range igReq.InstallIds {
				seen[id] = true
			}
			var unresolved []string
			for _, name := range group.InstallNames {
				id, ok := nameToID[name]
				if !ok {
					unresolved = append(unresolved, name)
					continue
				}
				if !seen[id] {
					igReq.InstallIds = append(igReq.InstallIds, id)
					seen[id] = true
				}
			}
			if len(unresolved) > 0 {
				return sync.SyncAPIErr{
					Resource: resource,
					Err:      fmt.Errorf("install group %q: unknown install names: %v", group.Name, unresolved),
				}
			}
		}

		if len(group.LabelSelector) > 0 {
			igReq.LabelSelector.GithubComNuoncoNuonPkgLabelsSelector.MatchLabels = models.GithubComNuoncoNuonPkgLabelsLabels(group.LabelSelector)
		}

		configReq.InstallGroups = append(configReq.InstallGroups, igReq)
	}

	_, err := s.apiClient.CreateAppBranchConfig(ctx, s.appID, branchID, configReq)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      fmt.Errorf("unable to create config for app branch %q: %w", branchCfg.Name, err),
		}
	}

	return nil
}

func (s *syncer) resolveInstallNames(ctx context.Context) (map[string]string, error) {
	nameToID := make(map[string]string)
	offset := 0
	for {
		installs, hasNext, err := s.apiClient.GetAppInstalls(ctx, s.appID, &models.GetPaginatedQuery{Offset: offset})
		if err != nil {
			return nil, err
		}
		for _, inst := range installs {
			if inst.Name != "" {
				nameToID[inst.Name] = inst.ID
			}
		}
		if !hasNext {
			break
		}
		offset += len(installs)
	}
	return nameToID, nil
}

func (s *syncer) getAllBranches() []*config.AppBranchConfig {
	var branches []*config.AppBranchConfig

	if s.cfg.Branch != nil {
		branches = append(branches, s.cfg.Branch)
	}

	branches = append(branches, s.cfg.Branches...)

	return branches
}
