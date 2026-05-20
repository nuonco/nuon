package apisyncer

import (
	"context"
	"fmt"
	"maps"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) syncAppSandbox(ctx context.Context, resource string) error {
	if s.cfg.Sandbox == nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      fmt.Errorf("sandbox config is nil"),
		}
	}
	req := s.getAppSandboxRequest()
	cfg, err := s.apiClient.CreateAppSandboxConfig(ctx, s.appID, req)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.state.SandboxConfigID = cfg.ID
	return nil
}

func (s *syncer) getAppSandboxRequest() *models.ServiceCreateAppSandboxConfigRequest {
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

	if s.cfg.Sandbox.MaxAutoRetries != nil {
		req.MaxAutoRetries = int64(*s.cfg.Sandbox.MaxAutoRetries)
	}
	if s.cfg.Sandbox.SkipNoops != nil {
		req.SkipNoops = *s.cfg.Sandbox.SkipNoops
	}
	if s.cfg.Sandbox.AutoApproveOnPoliciesPassing != nil {
		req.AutoApproveOnPoliciesPassing = *s.cfg.Sandbox.AutoApproveOnPoliciesPassing
	}
	maps.Copy(req.Variables, s.cfg.Sandbox.VarsMap)
	maps.Copy(req.EnvVars, s.cfg.Sandbox.EnvVarMap)

	for _, v := range s.cfg.Sandbox.VariablesFiles {
		req.VariablesFiles = append(req.VariablesFiles, v.Contents)
	}
	for _, ref := range s.cfg.Sandbox.References {
		req.References = append(req.References, ref.String())
	}

	if len(s.cfg.Sandbox.OperationRoles) > 0 {
		req.OperationRoles = make(map[string]string)
		for _, opRole := range s.cfg.Sandbox.OperationRoles {
			req.OperationRoles[string(opRole.Operation)] = opRole.RoleName
		}
	}

	if s.cfg.Sandbox.ConnectedRepo != nil {
		req.ConnectedGithubVcsConfig = &models.HelpersConnectedGithubVCSConfigRequest{
			Repo:      &s.cfg.Sandbox.ConnectedRepo.Repo,
			Branch:    s.cfg.Sandbox.ConnectedRepo.Branch,
			Directory: &s.cfg.Sandbox.ConnectedRepo.Directory,
		}
	}
	if s.cfg.Sandbox.PublicRepo != nil {
		req.PublicGitVcsConfig = &models.HelpersPublicGitVCSConfigRequest{
			Repo:      &s.cfg.Sandbox.PublicRepo.Repo,
			Branch:    &s.cfg.Sandbox.PublicRepo.Branch,
			Directory: &s.cfg.Sandbox.PublicRepo.Directory,
		}
	}

	return req
}
