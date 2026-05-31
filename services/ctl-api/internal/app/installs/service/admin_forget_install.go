package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/forgotten"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type AdminForgetInstallRequest struct{}

// @ID						AdminForgetInstall
// @Summary				forget an install
// @Description.markdown	forget_install.md
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	AdminForgetInstallRequest	true	"Input"
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/installs/{install_id}/admin-forget [POST]
func (s *service) AdminForgetInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.addForgottenByLabels(ctx, installID); err != nil {
		s.l.Warn("unable to add forgotten-by labels", zap.Error(err))
	}

	err = s.forgetInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
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
	ctx.JSON(http.StatusOK, true)
}

func (s *service) addForgottenByLabels(ctx context.Context, installID string) error {
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get account from context: %w", err)
	}

	var install app.Install
	if err := s.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		return fmt.Errorf("unable to get install %s: %w", installID, err)
	}

	install.Labels.Merge(labels.Labels{
		"nuon.co/forgotten-by-email":      account.Email,
		"nuon.co/forgotten-by-account-id": account.ID,
	})

	if err := s.db.WithContext(ctx).Model(&install).Select("labels").Updates(&install).Error; err != nil {
		return fmt.Errorf("unable to update install labels: %w", err)
	}

	return nil
}

func (s *service) forgetInstall(ctx context.Context, installID string) error {
	res := s.db.WithContext(ctx).Delete(&app.Install{
		ID: installID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete install: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("install not found %s %s", installID, gorm.ErrRecordNotFound)
	}
	return nil
}
