package sync

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
)

func (s *sync) syncAppSandbox(ctx context.Context, resource string) error {
	if s.cfg.Sandbox == nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      fmt.Errorf("sandbox config is nil"),
		}
	}
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
	req := &models.ServiceCreateAppSandboxConfigRequest{
		AppConfigID:      s.appConfigID,
		TerraformVersion: &s.cfg.Sandbox.TerraformVersion,
		Variables:        map[string]string{},
		EnvVars:          map[string]string{},
		VariablesFiles:   make([]string, 0),
		References:       make([]string, 0),
	}

	if s.cfg.Sandbox.DriftSchedule != nil {
		req.DriftSchedule = *s.cfg.Sandbox.DriftSchedule
	}

	for k, v := range s.cfg.Sandbox.VarsMap {
		req.Variables[k] = v
	}
	for k, v := range s.cfg.Sandbox.EnvVarMap {
		req.EnvVars[k] = v
	}
	for _, v := range s.cfg.Sandbox.VariablesFiles {
		req.VariablesFiles = append(req.VariablesFiles, v.Contents)
	}
	for _, ref := range s.cfg.Sandbox.References {
		req.References = append(req.References, ref.String())
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
