package syncer

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

// syncAction creates or updates an action workflow and its config.
// Duplicates logic from services/ctl-api/internal/app/actions/service/create_app_action_workflow.go
// and services/ctl-api/internal/app/actions/service/create_action_workflow_config.go
func (s *syncer) syncAction(ctx context.Context, action *config.ActionConfig) error {
	// Find or create action workflow
	var actionWorkflow app.ActionWorkflow
	res := s.db.WithContext(ctx).
		Where("app_id = ? AND name = ?", s.appID, action.Name).
		First(&actionWorkflow)

	if res.Error != nil {
		if res.Error != gorm.ErrRecordNotFound {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to find action workflow %s", action.Name),
				Err:         res.Error,
			}
		}

		// Create new action workflow
		actionWorkflow = app.ActionWorkflow{
			AppID: s.appID,
			Name:  action.Name,
		}

		res = s.db.WithContext(ctx).Create(&actionWorkflow)
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: fmt.Sprintf("unable to create action workflow %s", action.Name),
				Err:         res.Error,
			}
		}
	}

	// Parse timeout
	timeout := 5 * time.Minute // default
	if action.Timeout != "" {
		parsedTimeout, err := time.ParseDuration(action.Timeout)
		if err != nil {
			return sync.SyncErr{
				Resource:    fmt.Sprintf("action-%s", action.Name),
				Description: "invalid timeout duration",
			}
		}
		timeout = parsedTimeout
	}

	// Get app for VCS config
	var parentApp app.App
	res = s.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		First(&parentApp, "id = ?", s.appID)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to get app for action VCS config",
			Err:         res.Error,
		}
	}

	// Build triggers
	triggers := make([]app.ActionWorkflowTriggerConfig, 0, len(action.Triggers))
	for _, trigger := range action.Triggers {
		triggers = append(triggers, app.ActionWorkflowTriggerConfig{
			Index:        int(trigger.Index),
			Type:         app.ActionWorkflowTriggerType(trigger.Type),
			CronSchedule: trigger.CronSchedule,
			ComponentID:  generics.NewNullString(trigger.ComponentName),
		})
	}

	// Build steps
	vcsHelper := vcshelpers.New(vcshelpers.Params{})
	steps := make([]app.ActionWorkflowStepConfig, 0, len(action.Steps))

	for _, step := range action.Steps {
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
				return sync.SyncInternalErr{
					Description: fmt.Sprintf("unable to create connected github vcs config for action %s step %s", action.Name, step.Name),
					Err:         err,
				}
			}
		}

		if step.PublicRepo != nil {
			publicGitConfig, err = vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
				Repo:      step.PublicRepo.Repo,
				Branch:    step.PublicRepo.Branch,
				Directory: step.PublicRepo.Directory,
			})
			if err != nil {
				return sync.SyncInternalErr{
					Description: fmt.Sprintf("unable to create public git vcs config for action %s step %s", action.Name, step.Name),
					Err:         err,
				}
			}
		}

		// Convert references
		references := make([]string, 0)
		for _, ref := range step.References {
			references = append(references, ref.String())
		}

		// Convert env vars map to pgtype.Hstore
		envVars := pgtype.Hstore{}
		for k, v := range step.EnvVarMap {
			envVars[k] = &v
		}

		steps = append(steps, app.ActionWorkflowStepConfig{
			Name:                     step.Name,
			EnvVars:                  envVars,
			Command:                  step.Command,
			InlineContents:           step.InlineContents,
			References:               pq.StringArray(references),
			ConnectedGithubVCSConfig: githubVCSConfig,
			PublicGitVCSConfig:       publicGitConfig,
		})
	}

	// Convert action references
	actionReferences := make([]string, 0)
	for _, ref := range action.References {
		actionReferences = append(actionReferences, ref.String())
	}

	// Create action workflow config
	awc := app.ActionWorkflowConfig{
		AppConfigID:            s.appConfigID,
		ActionWorkflowID:       actionWorkflow.ID,
		Timeout:                timeout,
		ComponentDependencyIDs: pq.StringArray(action.Dependencies),
		References:             pq.StringArray(actionReferences),
		BreakGlassRoleARN:      generics.NewNullString(action.BreakGlassRole),
		Triggers:               triggers,
		Steps:                  steps,
	}

	res = s.db.WithContext(ctx).Create(&awc)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create action workflow config for %s", action.Name),
			Err:         res.Error,
		}
	}

	// Add to state
	s.state.Actions = append(s.state.Actions, sync.ActionState{
		Name: action.Name,
		ID:   actionWorkflow.ID,
	})

	return nil
}
