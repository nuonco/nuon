package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type WaitlistRequest struct {
	OrgName string `json:"org_name" validate:"required"`
}

func (c *WaitlistRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

//	@ID						CreateWaitlist
//	@Summary				Allow user to be added to an org waitlist.
//	@Description.markdown	create_waitlist.md
//	@Param					req	body	WaitlistRequest	true	"Input"
//	@Tags					general
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Success				200	{object}	app.Waitlist
//	@Router					/v1/general/waitlist [POST]
func (s *service) CreateWaitlist(ctx *gin.Context) {
	var req WaitlistRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if acct == nil {
		ctx.Error(fmt.Errorf("unable to find account in context"))
		return
	}

	waitlist, err := s.createWaitlist(ctx, req, *acct)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to join waitlist: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, waitlist)
}

func (s *service) createWaitlist(ctx context.Context, req WaitlistRequest, acct app.Account) (*app.Waitlist, error) {
	waitlist := app.Waitlist{
		CreatedByID: acct.ID,
		OrgName:     req.OrgName,
	}

	res := s.db.WithContext(ctx).
		Create(&waitlist)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create waitlist record: %w", res.Error)
	}

	return &waitlist, nil
}
