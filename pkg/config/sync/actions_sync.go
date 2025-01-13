package sync

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) syncAction(ctx context.Context, resource string, action *config.ActionConfig) (string, string, error) {
	isNew := false
	actionWorkflow, err := s.apiClient.GetAppActionWorkflow(ctx, s.appID, action.Name)
	if err != nil {
		if !nuon.IsNotFound(err) {
			return "", "", err
		}

		isNew = true
		actionWorkflow, err = s.apiClient.CreateActionWorkflow(ctx, s.appID, &models.ServiceCreateAppActionWorkflowRequest{
			Name: action.Name,
		})
		if err != nil {
			return "", "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	if !isNew {
		_, err = s.apiClient.UpdateActionWorkflow(ctx, actionWorkflow.ID, &models.ServiceUpdateActionWorkflowRequest{
			Name: action.Name,
		})
		if err != nil {
			return "", "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	timeout := time.Duration(0)
	if action.Timeout != "" {
		timeout, err = time.ParseDuration(action.Timeout)
		if err != nil {
			return "", "", SyncInternalErr{
				Description: "unable to parse timeout",
				Err:         err,
			}
		}
	}

	request := &models.ServiceCreateActionWorkflowConfigRequest{
		AppConfigID: generics.ToPtr(s.state.CfgID),
		Timeout:     timeout.Nanoseconds(),
	}

	for _, trigger := range action.Triggers {
		request.Triggers = append(request.Triggers, &models.ServiceCreateActionWorkflowConfigTriggerRequest{
			Type:         models.NewAppActionWorkflowTriggerType(models.AppActionWorkflowTriggerType(trigger.Type)),
			CronSchedule: trigger.CronSchedule,
		})
	}

	for _, step := range action.Steps {
		reqStep := &models.ServiceCreateActionWorkflowConfigStepRequest{
			Name:    generics.ToPtr(step.Name),
			EnvVars: step.EnvVarMap,
			Command: &step.Command,
		}

		if s.cfg.Sandbox.ConnectedRepo != nil {
			reqStep.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSActionWorkflowConfigRequest{
				Repo:      &s.cfg.Sandbox.ConnectedRepo.Repo,
				Branch:    s.cfg.Sandbox.ConnectedRepo.Branch,
				Directory: &s.cfg.Sandbox.ConnectedRepo.Directory,
			}
		}
		if s.cfg.Sandbox.PublicRepo != nil {
			reqStep.PublicGitVcsConfig = &models.ServicePublicGitVCSActionWorkflowConfigRequest{
				Repo:      &s.cfg.Sandbox.PublicRepo.Repo,
				Branch:    &s.cfg.Sandbox.PublicRepo.Branch,
				Directory: &s.cfg.Sandbox.PublicRepo.Directory,
			}
		}

		request.Steps = append(request.Steps, reqStep)
	}

	// INFO: We always create a new action workflow config per app config
	savedConfig, err := s.apiClient.CreateActionWorkflowConfig(ctx, actionWorkflow.ID, request)
	if err != nil {
		return "", "", SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return actionWorkflow.ID, savedConfig.ID, nil
}
