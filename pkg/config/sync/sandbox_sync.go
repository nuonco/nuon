package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
)

func (s *sync) syncAppSandbox(ctx context.Context, resource string) error {
	req := s.getAppSandboxRequest()
	cfg, err := s.apiClient.CreateAppSandboxConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.state.SandboxConfigID = cfg.ID
	return nil
}

func (s *sync) getAppSandboxRequest() *models.ServiceCreateAppSandboxConfigRequest {
	sandboxInputs := make(map[string]string)
	for _, v := range s.cfg.Sandbox.Vars {
		sandboxInputs[v.Name] = v.Value
	}
	for k, v := range s.cfg.Sandbox.VarMap {
		sandboxInputs[k] = v
	}

	req := &models.ServiceCreateAppSandboxConfigRequest{
		SandboxInputs:           sandboxInputs,
		TerraformVersion:        &s.cfg.Sandbox.TerraformVersion,
		AwsDelegationIamRoleArn: s.cfg.Sandbox.AWSDelegationIAMRoleARN,
	}

	if s.cfg.Sandbox.ConnectedRepo != nil {
		req.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSSandboxConfigRequest{
			Repo:      &s.cfg.Sandbox.ConnectedRepo.Repo,
			Branch:    s.cfg.Sandbox.ConnectedRepo.Branch,
			Directory: &s.cfg.Sandbox.ConnectedRepo.Directory,
		}
	}

	if s.cfg.Sandbox.PublicRepo != nil {
		req.PublicGitVcsConfig = &models.ServicePublicGitVCSSandboxConfigRequest{
			Repo:      &s.cfg.Sandbox.PublicRepo.Repo,
			Branch:    &s.cfg.Sandbox.PublicRepo.Branch,
			Directory: &s.cfg.Sandbox.PublicRepo.Directory,
		}
	}

	return req
}
