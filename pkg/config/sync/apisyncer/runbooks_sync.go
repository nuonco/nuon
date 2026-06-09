package apisyncer

import (
	"context"
	"fmt"
	"time"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
)

func (s *syncer) syncRunbook(ctx context.Context, resource string, runbook *config.RunbookConfig) (string, string, error) {
	isNew := false
	savedRunbook, err := s.apiClient.GetAppRunbook(ctx, s.appID, runbook.Name)
	if err != nil {
		if !nuon.IsNotFound(err) {
			return "", "", err
		}

		isNew = true
		savedRunbook, err = s.apiClient.CreateRunbook(ctx, s.appID, &models.ServiceCreateRunbookRequest{
			Name:        generics.ToPtr(runbook.Name),
			Description: runbook.Description,
			Labels:      runbook.Labels,
		})
		if err != nil {
			return "", "", sync.SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	if !isNew {
		_, err = s.apiClient.UpdateRunbook(ctx, savedRunbook.ID, &models.ServiceUpdateRunbookRequest{
			Name:        runbook.Name,
			Description: runbook.Description,
			Labels:      runbook.Labels,
		})
		if err != nil {
			return "", "", sync.SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	request := &models.ServiceCreateRunbookConfigRequest{
		AppConfigID: s.state.CfgID,
		Readme:      runbook.Readme,
	}

	for idx, step := range runbook.Steps {
		timeout := time.Duration(0)
		if step.Timeout != "" {
			timeout, _ = time.ParseDuration(step.Timeout)
		}

		request.Steps = append(request.Steps, &models.ServiceCreateRunbookStepConfigRequest{
			Name:                 generics.ToPtr(step.Name),
			Type:                 generics.ToPtr(string(step.Type)),
			Idx:                  int64(idx),
			ComponentName:        step.ComponentName,
			DeployDependents:     step.DeployDependents,
			TearDownDependents:   step.TearDownDependents,
			SkipComponentDeploys: step.SkipComponentDeploys,
			ActionName:           step.ActionName,
			Command:              step.Command,
			InlineContents:       step.InlineContents,
			EnvVars:              step.EnvVarMap,
			Timeout:              timeout.Nanoseconds(),
			Role:                 step.Role,
		})
	}

	for _, input := range runbook.Inputs {
		var defaultVal string
		if input.Default != nil {
			defaultVal = fmt.Sprintf("%v", input.Default)
		}

		request.Inputs = append(request.Inputs, &models.ServiceCreateRunbookInputRequest{
			Name:        generics.ToPtr(input.Name),
			DisplayName: input.DisplayName,
			Description: input.Description,
			Default:     defaultVal,
			Required:    input.Required,
			Sensitive:   input.Sensitive,
			Type:        input.Type,
		})
	}

	savedConfig, err := s.apiClient.CreateRunbookConfig(ctx, savedRunbook.ID, request)
	if err != nil {
		return "", "", sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	s.appendRunbookState(sync.RunbookState{
		Name: runbook.Name,
		ID:   savedRunbook.ID,
	})

	return savedRunbook.ID, savedConfig.ID, nil
}
