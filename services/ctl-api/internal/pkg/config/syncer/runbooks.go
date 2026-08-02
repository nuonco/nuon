package syncer

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ensureRunbook creates a runbook if it doesn't exist, using the shared helpers
// for full initialization (install runbooks).
func (s *syncer) ensureRunbook(ctx context.Context, runbook *config.RunbookConfig) error {
	existing, err := s.getRunbook(ctx, runbook.Name)
	if err == nil {
		res := s.db.WithContext(ctx).
			Model(&existing).
			Select("description", "labels").
			Updates(app.Runbook{
				Description: runbook.Description,
				Labeled:     labels.Labeled{Labels: labels.Labels(runbook.Labels)},
			})
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to update runbook %s", runbook.Name),
				Err:         res.Error,
			}
		}

		if err := s.runbooksHelpers.EnsureInstallRunbooks(ctx, s.appID, nil); err != nil {
			return sync.SyncInternalErr{Description: fmt.Sprintf("unable to ensure install runbooks for %s", runbook.Name), Err: err}
		}
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to check if runbook %s exists", runbook.Name),
			Err:         err,
		}
	}

	rbk := app.Runbook{
		AppID:       s.appID,
		Name:        runbook.Name,
		Description: runbook.Description,
	}
	if len(runbook.Labels) > 0 {
		rbk.Labels = labels.Labels(runbook.Labels)
	}
	res := s.db.WithContext(ctx).Create(&rbk)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create runbook %s", runbook.Name),
			Err:         res.Error,
		}
	}

	return nil
}

// syncRunbook creates a runbook config for the current app config.
func (s *syncer) syncRunbook(ctx context.Context, runbook *config.RunbookConfig) error {
	rbk, err := s.getRunbook(ctx, runbook.Name)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get runbook %s", runbook.Name),
			Err:         err,
		}
	}

	steps := make([]app.RunbookStepConfig, 0, len(runbook.Steps))
	for idx, step := range runbook.Steps {
		var trigger app.Trigger
		if step.Type == config.RunbookStepTypeWaitForEvent {
			var org app.Org
			if err := s.db.WithContext(ctx).Select("features").Where(app.Org{ID: s.orgID}).First(&org).Error; err != nil || !org.Features[string(app.OrgFeatureTriggers)] {
				return sync.SyncErr{Resource: fmt.Sprintf("runbook-%s", runbook.Name), Description: "triggers feature is not enabled"}
			}
			if err := s.db.WithContext(ctx).Where(app.Trigger{OrgID: s.orgID, Name: step.Trigger}).First(&trigger).Error; err != nil {
				return sync.SyncErr{Resource: fmt.Sprintf("runbook-%s", runbook.Name), Description: fmt.Sprintf("unable to find trigger %q", step.Trigger)}
			}
		}
		timeout := time.Duration(0)
		if step.Timeout != "" {
			parsedTimeout, err := time.ParseDuration(step.Timeout)
			if err != nil {
				return sync.SyncErr{
					Resource:    fmt.Sprintf("runbook-%s", runbook.Name),
					Description: fmt.Sprintf("invalid timeout duration for step %s", step.Name),
				}
			}
			timeout = parsedTimeout
		}

		envVars := pgtype.Hstore{}
		for k, v := range step.EnvVarMap {
			envVars[k] = &v
		}

		stepCfg := app.RunbookStepConfig{
			Idx:                  idx,
			Name:                 step.Name,
			Type:                 app.RunbookStepType(step.Type),
			PlanOnly:             step.PlanOnly,
			ComponentName:        step.ComponentName,
			DeployDependents:     step.DeployDependents,
			TearDownDependents:   step.TearDownDependents,
			SkipComponentDeploys: step.SkipComponentDeploys,
			Command:              step.Command,
			InlineContents:       step.InlineContents,
			EnvVars:              envVars,
			Timeout:              timeout,
			Role:                 step.Role,
			TriggerID:            trigger.ID,
			TriggerName:          trigger.Name,
			EventTypes:           step.EventTypes,
		}
		for _, filter := range step.Filters {
			stepCfg.Filters = append(stepCfg.Filters, app.TriggerFilter{From: filter.From, Path: filter.Path, Op: app.TriggerFilterType(filter.Op), Value: filter.Value})
		}

		// Resolve action_name to ActionWorkflowID
		if step.ActionName != "" {
			var aw app.ActionWorkflow
			if err := s.db.WithContext(ctx).
				Where(app.ActionWorkflow{AppID: s.appID, Name: step.ActionName}).
				First(&aw).Error; err != nil {
				return sync.SyncErr{
					Resource:    fmt.Sprintf("runbook-%s", runbook.Name),
					Description: fmt.Sprintf("unable to find action %q for step %s", step.ActionName, step.Name),
				}
			}
			stepCfg.ActionWorkflowID = generics.NewNullString(aw.ID)
		}

		steps = append(steps, stepCfg)
	}

	inputs := make([]app.RunbookInput, 0, len(runbook.Inputs))
	for idx, input := range runbook.Inputs {
		var defaultVal string
		if input.Default != nil {
			defaultVal = fmt.Sprintf("%v", input.Default)
		}
		inputType := generics.ValOrDefault(input.Type, "string")

		inputs = append(inputs, app.RunbookInput{
			Idx:         idx,
			Name:        input.Name,
			DisplayName: input.DisplayName,
			Description: input.Description,
			Default:     defaultVal,
			Required:    input.Required,
			Sensitive:   input.Sensitive,
			Type:        app.RunbookInputType(inputType),
		})
	}

	rbc := app.RunbookConfig{
		AppConfigID: s.appConfigID,
		RunbookID:   rbk.ID,
		AppID:       s.appID,
		Readme:      runbook.Readme,
		Steps:       steps,
		Inputs:      inputs,
	}

	res := s.db.WithContext(ctx).Create(&rbc)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create runbook config for %s", runbook.Name),
			Err:         res.Error,
		}
	}

	s.state.Runbooks = append(s.state.Runbooks, sync.RunbookState{
		Name: runbook.Name,
		ID:   rbk.ID,
	})

	return nil
}

// getRunbook finds a runbook by name.
func (s *syncer) getRunbook(ctx context.Context, name string) (*app.Runbook, error) {
	var rbk app.Runbook
	res := s.db.WithContext(ctx).
		Where(app.Runbook{AppID: s.appID, Name: name}).
		First(&rbk)

	if res.Error != nil {
		return nil, res.Error
	}

	return &rbk, nil
}
