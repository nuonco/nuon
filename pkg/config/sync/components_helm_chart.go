package sync

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) createHelmChartComponentConfig(ctx context.Context, resource, compID string, comp *config.Component) (string, string, error) {
	// NOTE(jm): this logic should be updated to be handled _before_ the config gets here.
	obj := comp.HelmChart

	configRequest := &models.ServiceCreateHelmComponentConfigRequest{
		ChartName:                generics.ToPtr(obj.ChartName),
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Values:                   map[string]string{},
		ValuesFiles:              make([]string, 0),
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

	requestChecksum, err := s.getChecksum(configRequest)
	if err != nil {
		return "", "", err
	}

	cmpBuild, err := s.apiClient.GetComponentLatestBuild(ctx, compID)
	if err != nil && !nuon.IsNotFound(err) {
		return "", "", err
	}

	doChecksumCompare := true
	if cmpBuild != nil && cmpBuild.Status == "error" {
		doChecksumCompare = false
	}

	if doChecksumCompare {
		prevComponentState := s.getComponentStateById(compID)
		if prevComponentState != nil && prevComponentState.Checksum == requestChecksum {
			return prevComponentState.ConfigID, requestChecksum, nil
		}
	}

	// NOTE: we don't want to make a checksum with the app config id since that can change
	configRequest.AppConfigID = s.appConfigID
	cfg, err := s.apiClient.CreateHelmComponentConfig(ctx, compID, configRequest)
	if err != nil {
		return "", "", err
	}

	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)

	return cfg.ID, requestChecksum, nil
}
