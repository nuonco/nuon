package sync

import (
	"context"
	"time"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

// rewrite the following trigger types for backwards compability
var rewriteTriggerTypes map[string]string = map[string]string{
	"pre-component-deploy":  "pre-deploy-component",
	"post-component-deploy": "post-deploy-component",
}

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
		AppConfigID:       generics.ToPtr(s.state.CfgID),
		Timeout:           timeout.Nanoseconds(),
		Dependencies:      action.Dependencies,
		BreakGlassRoleArn: action.BreakGlassRoleARN,
	}

	for _, ref := range action.References {
		request.References = append(request.References, ref.String())
	}

	for _, trigger := range action.Triggers {
		// TODO(jm): remove this once we finish rolling out all new trigger types
		typ := trigger.Type
		if _, ok := rewriteTriggerTypes[typ]; ok {
			typ = rewriteTriggerTypes[typ]
		}

		request.Triggers = append(request.Triggers, &models.ServiceCreateActionWorkflowConfigTriggerRequest{
			Type:          models.NewAppActionWorkflowTriggerType(models.AppActionWorkflowTriggerType(typ)),
			CronSchedule:  trigger.CronSchedule,
			ComponentName: trigger.ComponentName,
			Index:         trigger.Index,
		})
	}

	for _, step := range action.Steps {
		reqStep := &models.ServiceCreateActionWorkflowConfigStepRequest{
			Name:           generics.ToPtr(step.Name),
			EnvVars:        step.EnvVarMap,
			Command:        step.Command,
			InlineContents: step.InlineContents,
		}

		for _, ref := range step.References {
			reqStep.References = append(reqStep.References, ref.String())
		}

		if step.ConnectedRepo != nil {
			reqStep.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSActionWorkflowConfigRequest{
				Repo:      generics.ToPtr(step.ConnectedRepo.Repo),
				Branch:    step.ConnectedRepo.Branch,
				Directory: generics.ToPtr(step.ConnectedRepo.Directory),
			}
		}
		if step.PublicRepo != nil {
			reqStep.PublicGitVcsConfig = &models.ServicePublicGitVCSActionWorkflowConfigRequest{
				Repo:      generics.ToPtr(step.PublicRepo.Repo),
				Branch:    generics.ToPtr(step.PublicRepo.Branch),
				Directory: generics.ToPtr(step.PublicRepo.Directory),
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

	s.state.Actions = append(s.state.Actions, actionState{
		Name: action.Name,
		ID:   actionWorkflow.ID,
	})

	return actionWorkflow.ID, savedConfig.ID, nil
}
