package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminForgetAccountInstallsRequest struct {
	AccountID string
}

func (c *AdminForgetAccountInstallsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID AdminForgetAccountInstalls
// @Summary	forget all installs for an org
// @Description.markdown forget_account_installs.md
// @Param			req		body	AdminForgetAccountInstallsRequest	true	"Input"
// @Tags			installs/admin
// @Accept			json
// @Produce		json
// @Failure		400	{object}	stderr.ErrResponse
// @Failure		404	{object}	stderr.ErrResponse
// @Failure		500	{object}	stderr.ErrResponse
// @Success		200	{boolean}	true
// @Router			/v1/installs/admin-forget-account-installs [POST]
func (s *service) ForgetAccountInstalls(ctx *gin.Context) {
	var req AdminForgetAccountInstallsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installs, err := s.getAccountInstalls(ctx, req.AccountID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get account installs: %w", err))
		return
	}

	for _, install := range installs {
		err := s.forgetInstall(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		s.hooks.Forgotten(ctx, install.ID)
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getAccountInstalls(ctx context.Context, accountID string) ([]app.Install, error) {
	var installs []app.Install
	res := s.db.WithContext(ctx).
		Joins("JOIN aws_accounts on aws_accounts.install_id=installs.id").
		Where("aws_accounts.iam_role_arn LIKE ?", "%"+accountID+"%").
		Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, nil
}
