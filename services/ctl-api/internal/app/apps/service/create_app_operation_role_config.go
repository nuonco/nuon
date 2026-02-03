package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type OperationRoleRuleRequest struct {
	Principal string            `json:"principal" validate:"required"`
	Operation app.OperationType `json:"operation" validate:"required"`
	Role      string            `json:"role" validate:"required"`
}

type CreateAppOperationRoleConfigRequest struct {
	AppConfigID string                     `json:"app_config_id" validate:"required"`
	Rules       []OperationRoleRuleRequest `json:"rules" validate:"required,dive"`
}

func (c *CreateAppOperationRoleConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateAppOperationRoleConfig
// @Summary				create operation role config
// @Description			Create operation role rules for an app config
// @Tags					apps
// @Accept					json
// @Param					app_id	path	string										true	"app ID"
// @Param					req		body	CreateAppOperationRoleConfigRequest			true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppOperationRoleConfig
// @Router					/v1/apps/{app_id}/operation-role-configs [post]
func (s *service) CreateAppOperationRoleConfig(ctx *gin.Context) {
	var req CreateAppOperationRoleConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Create rule config
	cfg, err := s.createAppOperationRoleConfigRecord(ctx, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create operation role config: %w", err))
		return
	}

	// Create rules
	rules, err := s.createOperationRoleRules(ctx, cfg, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create operation role rules: %w", err))
		return
	}
	cfg.Rules = rules

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createAppOperationRoleConfigRecord(ctx context.Context, req *CreateAppOperationRoleConfigRequest) (*app.AppOperationRoleConfig, error) {
	cfg := app.AppOperationRoleConfig{
		AppConfigID: req.AppConfigID,
	}

	res := s.db.WithContext(ctx).Create(&cfg)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create operation role config: %w", res.Error)
	}

	return &cfg, nil
}

func (s *service) createOperationRoleRules(ctx context.Context, cfg *app.AppOperationRoleConfig, req *CreateAppOperationRoleConfigRequest) ([]*app.OperationRoleRule, error) {
	if len(req.Rules) == 0 {
		return []*app.OperationRoleRule{}, nil
	}

	rules := make([]*app.OperationRoleRule, 0, len(req.Rules))

	for _, ruleReq := range req.Rules {
		// Validate operation type
		operationType := app.OperationType(ruleReq.Operation)

		if !slices.Contains(app.ValidOperations, operationType) {
			return nil, fmt.Errorf("invalid operation type: %s", ruleReq.Operation)
		}

		// Parse principal to extract type and name
		p, err := principal.ParsePrincipal(ruleReq.Principal)
		if err != nil {
			return nil, fmt.Errorf("invalid principal %q: %w", ruleReq.Principal, err)
		}

		rules = append(rules, &app.OperationRoleRule{
			AppOperationRoleConfigID: cfg.ID,
			PrincipalType:            p.Type,
			PrincipalName:            p.Name,
			Operation:                operationType,
			Role:                     ruleReq.Role,
		})
	}

	res := s.db.WithContext(ctx).Create(&rules)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create operation role rules: %w", res.Error)
	}

	return rules, nil
}
