package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/forgotten"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminForgetAccountInstallsRequest struct {
	AccountID string `json:"account_id" validate:"required"`
}

func (c *AdminForgetAccountInstallsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AdminForgetAccountInstalls
// @Summary				forget all installs for an account
// @Description.markdown	forget_account_installs.md
// @Param					req	body	AdminForgetAccountInstallsRequest	true	"Input"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/installs/admin-forget-account-installs [POST]
func (s *service) ForgetAccountInstalls(ctx *gin.Context) {
	var req AdminForgetAccountInstallsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
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

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}

	for _, install := range installs {
		if err := s.addForgottenByLabels(ctx, install.ID); err != nil {
			s.l.Warn("unable to add forgotten-by labels", zap.Error(err))
		}

		err = s.forgetInstall(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		if useQueues {
			queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
			if err != nil {
				ctx.Error(err)
				return
			}
			if err := s.enqueueInstallSignal(ctx, queueID, &forgotten.Signal{
				InstallID: install.ID,
			}, "", ""); err != nil {
				ctx.Error(fmt.Errorf("enqueue signal: %w", err))
				return
			}
		} else {
			s.evClient.Send(ctx, install.ID, &signals.Signal{
				Type: signals.OperationForget,
			})
		}
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getAccountInstalls(ctx context.Context, accountID string) ([]app.Install, error) {
	var installs []app.Install
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("GCPAccount").
		Preload("App").
		Preload("App.Org").
		Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	// NOTE(jm): unfortunately, it's non trivial to use LIKE %foo% queries in gorm, so we just filter locally.
	var accountInstalls []app.Install
	for _, install := range installs {
		if install.App.Org.SandboxMode {
			continue
		}
		if install.AWSAccount == nil {
			continue
		}

		if !strings.Contains(install.AWSAccount.IAMRoleARN, accountID) {
			continue
		}

		accountInstalls = append(accountInstalls, install)
	}

	return accountInstalls, nil
}
