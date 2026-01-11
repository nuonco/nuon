package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
	validatoradapter "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AppInputRequest struct {
	DisplayName string `json:"display_name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
	Sensitive   bool   `json:"sensitive"`
	Group       string `json:"group" validate:"required"`
	Index       int    `json:"index" validate:"required"`

	// New, optional fields
	Internal bool               `json:"internal"`
	Type     string             `json:"type"`
	Source   app.AppInputSource `json:"source"`
}

type AppGroupRequest struct {
	DisplayName string `json:"display_name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Index       int    `json:"index" validate:"required"`
}

type CreateAppInputConfigRequest struct {
	AppConfigID string                     `json:"app_config_id"`
	Inputs      map[string]AppInputRequest `json:"inputs" validate:"required"`
	Groups      map[string]AppGroupRequest `json:"groups" validate:"required"`
}

func (c *CreateAppInputConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	for k, input := range c.Inputs {
		if err := validatoradapter.InterpolatedName(v, k); err != nil {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid input %s - %w", k, err),
				Description: fmt.Sprintf("Please use a valid input name: %s", k),
			}
		}

		if _, ok := c.Groups[input.Group]; !ok {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid group %s", input.Group),
				Description: fmt.Sprintf("Please use a valid group, or add %s as a group", input.Group),
			}
		}

		if !generics.SliceContains(app.AppInputType(input.Type), []app.AppInputType{
			app.AppInputTypeBool,
			app.AppInputTypeJSON,
			app.AppInputTypeList,
			app.AppInputTypeNumber,
			app.AppInputTypeString,
		}) {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid input type %s", input.Type),
				Description: "Please use a valid input type",
			}
		}

		if input.Type == "json" {
			if input.Default != "" && !json.Valid([]byte(input.Default)) {
				return stderr.ErrUser{
					Description: fmt.Sprintf("input %s has an invalid JSON string", input.DisplayName),
					Err:         fmt.Errorf("input %s default value is not valid JSON string", input.DisplayName),
				}
			}
		}

	}

	return nil
}

// @ID						CreateAppInputConfig
// @Description.markdown	create_app_input_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppInputConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppInputConfig
// @Router					/v1/apps/{app_id}/input-config [post]
func (s *service) CreateAppInputsConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	var req CreateAppInputConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// id a required input is added to app config, and there exists installs associated to this app, we allow addition
	// of new inputs only if default value is provided. Else there is no way to satisfy config contract since there's
	// value to newly added input
	if err := s.validateRequiredInputsWithInstalls(ctx, appID, &req); err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.createAppInputGroups(ctx, org.ID, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app input groups and config: %w", err))
		return
	}

	inputs, err := s.createAppInputs(ctx, cfg, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app inputs config: %w", err))
		return
	}
	cfg.AppInputs = inputs

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createAppInputGroups(ctx context.Context, orgID, appID string, req *CreateAppInputConfigRequest) (*app.AppInputConfig, error) {
	groups := make([]app.AppInputGroup, 0, len(req.Groups))
	for name, grp := range req.Groups {
		groups = append(groups, app.AppInputGroup{
			Name:        name,
			Description: grp.Description,
			DisplayName: grp.DisplayName,
			Index:       grp.Index,
		})
	}

	cfg := app.AppInputConfig{
		AppConfigID:    req.AppConfigID,
		OrgID:          orgID,
		AppID:          appID,
		AppInputGroups: groups,
	}

	res := s.db.WithContext(ctx).Create(&cfg)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app groups: %w", res.Error)
	}

	return &cfg, nil
}

func (s *service) createAppInputs(ctx context.Context, cfg *app.AppInputConfig, req *CreateAppInputConfigRequest) ([]app.AppInput, error) {
	if len(req.Inputs) == 0 {
		return []app.AppInput{}, nil
	}

	inputs := make([]app.AppInput, 0, len(req.Inputs))

	for name, input := range req.Inputs {
		var groupID string
		for _, group := range cfg.AppInputGroups {
			if group.Name == input.Group {
				groupID = group.ID
				break
			}
		}

		source := input.Source
		if source == "" {
			source = app.AppInputSourceVendor
		}

		inputs = append(inputs, app.AppInput{
			OrgID:            cfg.OrgID,
			AppInputConfigID: cfg.ID,
			AppInputGroupID:  groupID,
			Name:             name,
			Description:      input.Description,
			DisplayName:      input.DisplayName,
			Required:         input.Required,
			Default:          input.Default,
			Sensitive:        input.Sensitive,
			Type:             app.AppInputType(input.Type),
			Internal:         input.Internal,
			Index:            input.Index,
			Source:           source,
		})
	}

	res := s.db.WithContext(ctx).Create(&inputs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app inputs: %w", res.Error)
	}

	return inputs, nil
}

func (s *service) validateRequiredInputsWithInstalls(ctx context.Context, appID string, req *CreateAppInputConfigRequest) error {
	var installCount int64
	res := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("app_id = ?", appID).
		Count(&installCount)

	if res.Error != nil {
		return fmt.Errorf("unable to check install count: %w", res.Error)
	}

	if installCount == 0 {
		return nil
	}

	var existingInputConfig app.AppInputConfig
	res = s.db.WithContext(ctx).
		Where("app_id = ?", appID).
		Preload("AppInputs").
		Order("created_at DESC").
		First(&existingInputConfig)

	existingInputNames := make(map[string]bool)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("unable to check existing input config: %w", res.Error)
	}

	// empty map will be formed if no records are found
	for _, inp := range existingInputConfig.AppInputs {
		existingInputNames[inp.Name] = true
	}

	// Installs exist - validate NEW required inputs have defaults
	for name, inputReq := range req.Inputs {
		// no need to check for old input values
		if existingInputNames[name] {
			continue
		}

		if !inputReq.Required {
			continue
		}

		if inputReq.Default == "" {
			return stderr.ErrUser{
				Err: fmt.Errorf("required input '%s' is missing a default value", name),
				Description: fmt.Sprintf(
					"Cannot add new required input '%s' without a default value because %d install(s) exist for this app. "+
						"When existing installs are present, all new required inputs must have default values to ensure "+
						"data integrity during migration. Existing inputs from previous syncs are not affected. "+
						"Please add a default value to this input in your app config.",
					name, installCount,
				),
			}
		}
	}

	return nil
}
