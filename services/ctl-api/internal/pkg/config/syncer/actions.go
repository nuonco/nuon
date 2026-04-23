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
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

// ensureAction creates an action workflow if it doesn't exist, using the shared helpers
// for full initialization (install action workflows).
func (s *syncer) ensureAction(ctx context.Context, action *config.ActionConfig) error {
	_, err := s.getAction(ctx, action.Name)
	if err == nil {
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to check if action %s exists", action.Name),
			Err:         err,
		}
	}

	_, err = s.actionsHelpers.CreateAction(ctx, &actionshelpers.CreateActionParams{
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

// syncAction updates an action workflow and creates its config.
func (s *syncer) syncAction(ctx context.Context, action *config.ActionConfig) error {
	actionWorkflow, err := s.getAction(ctx, action.Name)
	if err != nil {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get action %s", action.Name),
			Err:         err,
		}
	}

	// Sync labels
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

// getAction finds an action workflow by name.
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
