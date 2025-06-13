package sync

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/hasher"
)

func (s *sync) createHelmChartComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	// NOTE(jm): this logic should be updated to be handled _before_ the config gets here.
	obj := comp.HelmChart

	configRequest := &models.ServiceCreateHelmComponentConfigRequest{
		AppConfigID:              s.appConfigID,
		Dependencies:             comp.Dependencies,
		ChartName:                generics.ToPtr(obj.ChartName),
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Values:                   map[string]string{},
		ValuesFiles:              make([]string, 0),
		Namespace:                obj.Namespace,
		StorageDriver:            obj.StorageDriver,
		TakeOwnership:            obj.TakeOwnership,
	}

	for _, ref := range comp.References {
		configRequest.References = append(configRequest.References, ref.String())
	}

	if obj.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    generics.ToPtr(obj.PublicRepo.Branch),
			Directory: generics.ToPtr(obj.PublicRepo.Directory),
			Repo:      generics.ToPtr(obj.PublicRepo.Repo),
		}
	}
	if obj.ConnectedRepo != nil {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    obj.ConnectedRepo.Branch,
			Directory: generics.ToPtr(obj.ConnectedRepo.Directory),
			// NOTE: GitRef is not required for config sync
			Repo: generics.ToPtr(obj.ConnectedRepo.Repo),
		}
	}
	for _, value := range obj.Values {
		configRequest.Values[value.Name] = value.Value
	}
	for k, v := range obj.ValuesMap {
		configRequest.Values[k] = v
	}

	for _, value := range obj.ValuesFiles {
		configRequest.ValuesFiles = append(configRequest.ValuesFiles, value.Contents)
	}

	newChecksum, err := hasher.HashStruct(comp)
	if err != nil {
		return "", "", err
	}

	// Check if we should skip this build due to checksum match
	shouldSkip, existingConfigID, err := s.shouldSkipBuildDueToChecksum(ctx, compID, newChecksum)
	if err != nil {
		return "", "", err
	}

	if shouldSkip {
		return existingConfigID, newChecksum, nil
	}

	configRequest.Checksum = newChecksum
	cfg, err := s.apiClient.CreateHelmComponentConfig(ctx, compID, configRequest)
	if err != nil {
		fmt.Println("here", err)
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, newChecksum, nil
}
