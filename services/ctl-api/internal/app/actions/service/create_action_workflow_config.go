package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

// @ID						CreateActionConfig
// @Summary				create action config
// @Description.markdown	create_action_workflow_config.md
// @Param					app_id		path	string	true	"app ID"
// @Param					action_id	path	string	true	"action ID"
// @Tags					actions
// @Accept					json
// @Param					req	body	CreateActionWorkflowConfigRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ActionWorkflowConfig
// @Router					/v1/apps/{app_id}/actions/{action_id}/configs [post]
func (s *service) CreateAppActionConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	awID := ctx.Param("action_id")
	aw, err := s.findActionWorkflow(ctx, org.ID, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow %s: %w", awID, err))
		return
	}

	parentApp, err := s.findApp(ctx, org.ID, aw.AppID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", parentApp.ID, err))
		return
	}

	var req CreateActionWorkflowConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	awc, err := s.createActionWorkflowConfig(ctx, parentApp, org.ID, awID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	s.evClient.Send(ctx, awID, &signals.Signal{
		Type: signals.OperationConfigCreated,

		ConfigID: awc.ID,
	})

	ctx.JSON(http.StatusCreated, awc)
}

type CreateActionWorkflowConfigRequest struct {
	AppConfigID string                                     `json:"app_config_id" validate:"required"`
	Triggers    []CreateActionWorkflowConfigTriggerRequest `json:"triggers" validate:"required,dive"`
	Steps       []CreateActionWorkflowConfigStepRequest    `json:"steps" validate:"required,dive"`
	Timeout     time.Duration                              `json:"timeout" swaggertype:"primitive,integer"`

	Dependencies []string `json:"dependencies"`
	References   []string `json:"references"`

	BreakGlassRoleARN string `json:"break_glass_role_arn"`
}

type CreateActionWorkflowConfigTriggerRequest struct {
	Index         int                           `json:"index,omitempty" swaggertype:"primitive,integer"`
	Type          app.ActionWorkflowTriggerType `json:"type" validate:"required"`
	CronSchedule  string                        `json:"cron_schedule,omitempty" validate:"cron_schedule"`
	ComponentName string                        `json:"component_name"`
}

type CreateActionWorkflowConfigStepRequest struct {
	basicVCSConfigRequest
	Name    string             `json:"name" validate:"required"`
	EnvVars map[string]*string `json:"env_vars"`

	Command        string `json:"command"`
	InlineContents string `json:"inline_contents"`

	References []string `json:"references"`
}

const (
	maxTimeout     = time.Hour
	defaultTimeout = 5 * time.Minute
)

func (c *CreateActionWorkflowConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if c.Timeout > maxTimeout {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid timeout"),
			Description: "timeout cannot exceed " + maxTimeout.String(),
		}
	}

	// verify crons
	cronCount := 0
	for _, trigger := range c.Triggers {
		if trigger.Type == app.ActionWorkflowTriggerTypeCron {
			cronCount += 1
		}
	}
	if cronCount > 1 {
		return stderr.ErrUser{
			Err:         errors.New("more than one cron trigger defined"),
			Description: "only one cron trigger can be defined at a time",
		}
	}

	// verify component is set for component dependency
	for _, trigger := range c.Triggers {
		if !generics.SliceContains(app.ActionWorkflowTriggerType(trigger.Type), app.AllActionWorkflowTriggerTypes) {
			return stderr.ErrUser{
				Err: fmt.Errorf("invalid trigger type %s", trigger.Type),
				Description: fmt.Sprintf("trigger type must be one of (%s)",
					strings.Join(generics.ToStringSlice(app.AllActionWorkflowTriggerTypes), ", ")),
			}
		}

		if generics.SliceContains(app.ActionWorkflowTriggerType(trigger.Type), app.AllActionWorkflowComponentTriggerTypes) {
			if trigger.ComponentName == "" {
				return stderr.ErrUser{
					Err:         errors.New(fmt.Sprintf("component_name must be set on %s trigger", trigger.Type)),
					Description: fmt.Sprintf("component_name must be set on %s trigger", trigger.Type),
				}
			}
		}

		if trigger.ComponentName != "" && !generics.SliceContains(app.ActionWorkflowTriggerType(trigger.Type), app.AllActionWorkflowComponentTriggerTypes) {
			return stderr.ErrUser{
				Err: errors.New(fmt.Sprintf("component_name not supported for %s trigger", trigger.Type)),
				Description: fmt.Sprintf("component_name only available for (%s) triggers",
					strings.Join(generics.ToStringSlice(app.AllActionWorkflowComponentTriggerTypes), ", ")),
			}
		}
	}

	return nil
}

// @ID						CreateActionWorkflowConfig
// @Summary				create action workflow config
// @Description.markdown	create_action_workflow_config.md
// @Param					action_workflow_id	path	string	true	"action workflow ID"
// @Tags					actions
// @Accept					json
// @Param					req	body	CreateActionWorkflowConfigRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated  			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ActionWorkflowConfig
// @Router					/v1/action-workflows/{action_workflow_id}/configs [post]
func (s *service) CreateActionWorkflowConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	awID := ctx.Param("action_workflow_id")
	aw, err := s.findActionWorkflow(ctx, org.ID, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow %s: %w", awID, err))
		return
	}

	parentApp, err := s.findApp(ctx, org.ID, aw.AppID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", parentApp.ID, err))
		return
	}

	var req CreateActionWorkflowConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	awc, err := s.createActionWorkflowConfig(ctx, parentApp, org.ID, awID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	s.evClient.Send(ctx, awID, &signals.Signal{
		Type: signals.OperationConfigCreated,

		ConfigID: awc.ID,
	})

	ctx.JSON(http.StatusCreated, awc)
}

func (s *service) createActionWorkflowConfig(ctx context.Context, parentApp *app.App, orgID string, awID string, req *CreateActionWorkflowConfigRequest) (*app.ActionWorkflowConfig, error) {
	timeout := req.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	depIDs, err := s.compHelpers.GetComponentIDs(ctx, parentApp.ID, req.Dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component ids")
	}

	awc := app.ActionWorkflowConfig{
		AppID:                  parentApp.ID,
		AppConfigID:            req.AppConfigID,
		OrgID:                  orgID,
		ActionWorkflowID:       awID,
		Timeout:                timeout,
		ComponentDependencyIDs: pq.StringArray(depIDs),
		References:             pq.StringArray(req.References),
		BreakGlassRoleARN:      generics.NewNullString(req.BreakGlassRoleARN),
	}

	res := s.db.WithContext(ctx).
		Create(&awc)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create action workflow config: %w", res.Error)
	}

	if err := s.createActionWorkflowTriggers(ctx, orgID, parentApp.ID, req.AppConfigID, awc.ID, req.Triggers); err != nil {
		return nil, fmt.Errorf("unable to create action workflow triggers: %w", err)
	}

	if err := s.createActionWorkflowSteps(ctx, parentApp, orgID, req.AppConfigID, awc.ID, req.Steps); err != nil {
		return nil, fmt.Errorf("unable to create action workflow steps: %w", err)
	}

	return &awc, nil
}

func (s *service) createActionWorkflowTriggers(ctx context.Context, orgId, appID, appConfigID, awcID string, triggers []CreateActionWorkflowConfigTriggerRequest) error {
	for _, trigger := range triggers {
		var componentID string
		if trigger.ComponentName != "" {
			comp, err := s.compHelpers.GetComponentByName(ctx, appID, trigger.ComponentName)
			if err != nil {
				return stderr.ErrUser{
					Err:         err,
					Description: "unable to find component " + trigger.ComponentName,
				}
			}

			componentID = comp.ID
		}

		newTrigger := app.ActionWorkflowTriggerConfig{
			OrgID:                  orgId,
			AppID:                  appID,
			AppConfigID:            appConfigID,
			ActionWorkflowConfigID: awcID,
			Type:                   trigger.Type,
			CronSchedule:           trigger.CronSchedule,
			Index:                  trigger.Index,
			ComponentID:            generics.NewNullString(componentID),
		}

		res := s.db.WithContext(ctx).
			Create(&newTrigger)
		if res.Error != nil {
			return fmt.Errorf("unable to create action workflow trigger: %w", res.Error)
		}
	}

	return nil
}

func (s *service) createActionWorkflowSteps(ctx context.Context, parentApp *app.App, orgId, appConfigID, awcID string, steps []CreateActionWorkflowConfigStepRequest) error {
	stepIdx := 0
	prevStepId := ""

	for _, step := range steps {
		githubVCSConfig, err := step.connectedGithubVCSConfig(ctx, parentApp, s.vcsHelpers)
		if err != nil {
			return fmt.Errorf("unable to create connected github vcs config: %w", err)
		}

		publicGitConfig, err := step.publicGitVCSConfig()
		if err != nil {
			return fmt.Errorf("unable to get public git config: %w", err)
		}

		newStep := app.ActionWorkflowStepConfig{
			OrgID:                    orgId,
			AppID:                    parentApp.ID,
			AppConfigID:              appConfigID,
			ActionWorkflowConfigID:   awcID,
			Name:                     step.Name,
			EnvVars:                  step.EnvVars,
			Command:                  step.Command,
			InlineContents:           step.InlineContents,
			Idx:                      stepIdx,
			PreviousStepID:           prevStepId,
			PublicGitVCSConfig:       publicGitConfig,
			ConnectedGithubVCSConfig: githubVCSConfig,
			References:               pq.StringArray(step.References),
		}

		res := s.db.WithContext(ctx).
			Create(&newStep)
		if res.Error != nil {
			return fmt.Errorf("unable to create action workflow step: %w", res.Error)
		}

		prevStepId = newStep.ID
		stepIdx++
	}

	return nil
}
