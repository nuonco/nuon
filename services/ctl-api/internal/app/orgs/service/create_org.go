package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	orgcreated "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/created"
	orgprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/provision"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateOrgRequest struct {
	Name           string   `json:"name" validate:"required"`
	UseSandboxMode bool     `json:"use_sandbox_mode"`
	Tags           []string `json:"tags" swaggertype:"array,string"`
}

func (c *CreateOrgRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateOrg
// @Summary				create a new org
// @Description.markdown	create_org.md
// @Security				APIKey
// @Param					req	body	CreateOrgRequest	true	"Input"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Org
// @Router					/v1/orgs [POST]
func (s *service) CreateOrg(ctx *gin.Context) {
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if acct.AccountType == app.AccountTypeService {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("This email is not allowed to create new orgs."),
			Description: "Please reach out to team@nuon.co for access.",
		})
		return
	}

	req := CreateOrgRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	newOrg, err := s.helpers.CreateOrg(ctx, acct, &orgshelpers.CreateOrgParams{
		Name:           req.Name,
		UseSandboxMode: req.UseSandboxMode,
		Tags:           req.Tags,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}
	cctx.SetOrgGinContext(ctx, newOrg)

	// Always use v2 queue signals for org creation — the org was just created
	// so feature flags won't be set yet.
	signalsQueueID, err := s.getOrgSignalsQueueID(ctx, newOrg.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org signals queue: %w", err))
		return
	}
	if err := s.enqueueOrgSignal(ctx, signalsQueueID, &orgcreated.Signal{OrgID: newOrg.ID}, newOrg.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue org created signal: %w", err))
		return
	}
	if err := s.enqueueOrgSignal(ctx, signalsQueueID, &orgprovision.Signal{OrgID: newOrg.ID}, newOrg.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue org provision signal: %w", err))
		return
	}

	// Update user journey for first org creation
	if err := s.accountsHelpers.UpdateUserJourneyStepForFirstOrg(ctx, acct.ID, newOrg.ID); err != nil {
		// Log error but don't fail org creation
		s.l.Warn("failed to update user journey for first org",
			zap.String("account_id", acct.ID),
			zap.String("org_id", newOrg.ID),
			zap.Error(err))
	}

	ctx.JSON(http.StatusCreated, newOrg)
}
