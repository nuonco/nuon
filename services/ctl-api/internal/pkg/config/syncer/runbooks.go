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
	_, err := s.getRunbook(ctx, runbook.Name)
	if err == nil {
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

	// Ensure install runbooks for all existing installs
	if err := s.runbooksHelpers.EnsureInstallRunbooks(ctx, s.appID, nil); err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to ensure install runbooks for %s", runbook.Name),
			Err:         err,
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
			Idx:                idx,
			Name:               step.Name,
			Type:               app.RunbookStepType(step.Type),
			ComponentName:      step.ComponentName,
			DeployDependencies: step.DeployDependencies,
			Command:            step.Command,
			InlineContents:     step.InlineContents,
			EnvVars:            envVars,
			Timeout:            timeout,
			Role:               step.Role,
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

	rbc := app.RunbookConfig{
		AppConfigID: s.appConfigID,
		RunbookID:   rbk.ID,
		AppID:       s.appID,
		Readme:      runbook.Readme,
		Steps:       steps,
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
