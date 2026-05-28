package apisyncer

import (
	"context"
	"time"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"

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
		savedRunbook, err = s.apiClient.CreateRunbook(ctx, s.appID, &nuon.CreateRunbookRequest{
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

	if !isNew {
		_, err = s.apiClient.UpdateRunbook(ctx, savedRunbook.ID, &nuon.UpdateRunbookRequest{
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

	request := &nuon.CreateRunbookConfigRequest{
		AppConfigID: generics.ToPtr(s.state.CfgID),
		Readme:      runbook.Readme,
	}

	for idx, step := range runbook.Steps {
		timeout := time.Duration(0)
		if step.Timeout != "" {
			timeout, _ = time.ParseDuration(step.Timeout)
		}

		request.Steps = append(request.Steps, &nuon.CreateRunbookStepConfigRequest{
			Name:               step.Name,
			Type:               string(step.Type),
			Idx:                int64(idx),
			ComponentName:      step.ComponentName,
			DeployDependencies: step.DeployDependencies,
			ActionName:         step.ActionName,
			Command:            step.Command,
			InlineContents:     step.InlineContents,
			EnvVars:            step.EnvVarMap,
			Timeout:            timeout.Nanoseconds(),
			Role:               step.Role,
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
