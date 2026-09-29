package syncer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/syncerr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
)

func (s *syncer) ensureAction(ctx context.Context, action *config.ActionConfig) error {
	_, err := s.getAction(ctx, action.Name)
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to check if action %s exists", action.Name),
			Err:         err,
		}
	}

	_, err = s.actionsHelpers.CreateActionWithDB(ctx, s.db, &actionshelpers.CreateActionParams{
		AppID:  s.appID,
		OrgID:  s.orgID,
		Name:   action.Name,
		Labels: action.Labels,
	})
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create action %s", action.Name),
			Err:         err,
		}
	}

	return nil
}

func (s *syncer) syncAction(ctx context.Context, action *config.ActionConfig) error {
	actionWorkflow, err := s.getAction(ctx, action.Name)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get action %s", action.Name),
			Err:         err,
		}
	}

	labelRes := s.db.WithContext(ctx).
		Model(&actionWorkflow).
		Select("labels").
		Updates(app.ActionWorkflow{Labeled: labels.Labeled{Labels: labels.Labels(action.Labels)}})
	if labelRes.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to update labels for action workflow %s", action.Name),
			Err:         labelRes.Error,
		}
	}

	timeout := 5 * time.Minute
	if action.Timeout != "" {
		parsedTimeout, err := time.ParseDuration(action.Timeout)
		if err != nil {
			return sync.SyncErr{
				Resource:    fmt.Sprintf("action-%s", action.Name),
				Description: "invalid timeout duration",
				Err:         err,
			}
		}
		timeout = parsedTimeout
	}

	var parentApp app.App
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		First(&parentApp, "id = ?", s.appID)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to get app for action VCS config",
			Err:         res.Error,
		}
	}

	triggers := make([]app.ActionWorkflowTriggerConfig, 0, len(action.Triggers))
	for _, trigger := range action.Triggers {
		if err := validation.ValidateCronSchedule(trigger.CronSchedule); err != nil {
			return sync.SyncErr{
				Resource:    fmt.Sprintf("action-%s", action.Name),
				Description: err.Error(),
				Err:         err,
			}
		}

		var componentID generics.NullString
		if trigger.ComponentName != "" {
			resolved := false
			for _, comp := range s.state.Components {
				if comp.Name == trigger.ComponentName {
					componentID = generics.NewNullString(comp.ID)
					resolved = true
					break
				}
			}
			if !resolved {
				var comp app.Component
				if err := s.db.WithContext(ctx).
					Where("app_id = ? AND name = ? AND deleted_at = 0", s.appID, trigger.ComponentName).
					First(&comp).Error; err == nil {
					componentID = generics.NewNullString(comp.ID)
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return sync.SyncInternalErr{Description: fmt.Sprintf("unable to resolve trigger component %q", trigger.ComponentName), Err: err}
				}
			}
		}

		triggers = append(triggers, app.ActionWorkflowTriggerConfig{
			AppID:        s.appID,
			AppConfigID:  s.appConfigID,
			Index:        int(trigger.Index),
			Type:         app.ActionWorkflowTriggerType(trigger.Type),
			CronSchedule: trigger.CronSchedule,
			ComponentID:  componentID,
		})
	}

	vcsHelper := s.vcsHelpers
	steps := make([]app.ActionWorkflowStepConfig, 0, len(action.Steps))

	for idx, step := range action.Steps {
		var githubVCSConfig *app.ConnectedGithubVCSConfig
		var publicGitConfig *app.PublicGitVCSConfig
		var err error

		if step.ConnectedRepo != nil {
			githubVCSConfig, err = vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
				Repo:      step.ConnectedRepo.Repo,
				Branch:    step.ConnectedRepo.Branch,
				Directory: step.ConnectedRepo.Directory,
			}, parentApp.Org)
			if err != nil {
				return syncerr.From(fmt.Sprintf("action-%s", action.Name), fmt.Sprintf("unable to create connected github vcs config for step %s", step.Name), err)
			}
		}

		if step.PublicRepo != nil {
			publicGitConfig, err = vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
				Repo:      step.PublicRepo.Repo,
				Branch:    step.PublicRepo.Branch,
				Directory: step.PublicRepo.Directory,
			})
			if err != nil {
				return syncerr.From(fmt.Sprintf("action-%s", action.Name), fmt.Sprintf("unable to create public git vcs config for step %s", step.Name), err)
			}
		}

		references := make([]string, 0)
		for _, ref := range step.References {
			references = append(references, ref.String())
		}

		envVars := pgtype.Hstore{}
		for k, v := range step.EnvVarMap {
			envVars[k] = &v
		}

		steps = append(steps, app.ActionWorkflowStepConfig{
			AppID:                    s.appID,
			AppConfigID:              s.appConfigID,
			Idx:                      idx,
			Name:                     step.Name,
			EnvVars:                  envVars,
			Command:                  step.Command,
			InlineContents:           step.InlineContents,
			References:               pq.StringArray(references),
			ConnectedGithubVCSConfig: githubVCSConfig,
			PublicGitVCSConfig:       publicGitConfig,
		})
	}

	actionReferences := make([]string, 0)
	for _, ref := range action.References {
		actionReferences = append(actionReferences, ref.String())
	}

	depIDs := []string{}
	if len(action.Dependencies) > 0 {
		depIDs, err = s.componentHelpers.GetComponentIDsWithDB(ctx, s.db, s.appID, action.Dependencies)
		if err != nil {
			return syncerr.From(fmt.Sprintf("action-%s", action.Name), "unable to resolve dependencies", err)
		}
	}

	built := build.ActionWorkflowConfig(build.ActionWorkflowInput{
		AppID:                 s.appID,
		AppConfigID:           s.appConfigID,
		OrgID:                 s.orgID,
		ActionWorkflowID:      actionWorkflow.ID,
		Timeout:               timeout,
		DependencyIDs:         depIDs,
		References:            actionReferences,
		BreakGlassRole:        action.BreakGlassRole,
		Role:                  action.Role,
		EnableKubeConfig:      action.EnableKubeConfig,
		KubernetesContextName: action.KubernetesContext,
		Image:                 action.Image,
	})
	built.Triggers = triggers
	built.Steps = steps
	awc := *built

	var existing app.ActionWorkflowConfig
	if err := s.db.WithContext(ctx).
		Where("app_config_id = ? AND action_workflow_id = ? AND deleted_at = 0", s.appConfigID, actionWorkflow.ID).
		First(&existing).Error; err == nil {
		awc = existing
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return sync.SyncInternalErr{Description: fmt.Sprintf("unable to look up action workflow config for %s", action.Name), Err: err}
	} else {
		res = s.db.WithContext(ctx).Create(&awc)
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to create action workflow config for %s", action.Name),
				Err:         res.Error,
			}
		}
	}

	s.state.Actions = append(s.state.Actions, sync.ActionState{
		Name: action.Name,
		ID:   actionWorkflow.ID,
	})

	return nil
}

func (s *syncer) getAction(ctx context.Context, name string) (*app.ActionWorkflow, error) {
	var aw app.ActionWorkflow
	res := s.db.WithContext(ctx).
		Where("app_id = ? AND name = ?", s.appID, name).
		First(&aw)

	if res.Error != nil {
		return nil, res.Error
	}

	return &aw, nil
}
