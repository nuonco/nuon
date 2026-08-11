package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/syncinstalls"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type RespondInstallCreationApprovalRequest struct {
	ResponseType string `json:"response_type" binding:"required,oneof=approve deny"`
}

// @ID						RespondInstallCreationApproval
// @Summary				respond to an install creation approval
// @Description			Approves or denies an install creation approval. On approve, creates the missing installs and re-triggers the sync. On deny, marks the approval as denied.
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					app_id			path	string										true	"app ID"
// @Param					sync_id			path	string										true	"sync ID"
// @Param					approval_id		path	string										true	"approval ID"
// @Param					req				body	RespondInstallCreationApprovalRequest		true	"Input"
// @Success				202	{object}	map[string]string
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/install-syncs/{sync_id}/approvals/{approval_id}/response [post]
func (s *service) RespondInstallCreationApproval(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	var req RespondInstallCreationApprovalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	appID := ctx.Param("app_id")
	approvalID := ctx.Param("approval_id")

	var approval app.InstallCreationApproval
	if err := s.db.WithContext(ctx).
		Where(app.InstallCreationApproval{ID: approvalID, AppID: appID, OrgID: org.ID}).
		First(&approval).Error; err != nil {
		ctx.Error(fmt.Errorf("approval not found: %w", err))
		return
	}

	if approval.Status != app.InstallCreationApprovalStatusPending {
		ctx.Error(fmt.Errorf("approval is not pending (current: %s)", approval.Status))
		return
	}

	switch req.ResponseType {
	case "approve":
		s.approveInstallCreation(ctx, &approval, appID, account.ID)
	case "deny":
		s.denyInstallCreation(ctx, &approval, account.ID)
	}
}

func (s *service) approveInstallCreation(ctx *gin.Context, approval *app.InstallCreationApproval, appID, accountID string) {
	for _, proposed := range approval.ProposedInstalls {
		_, err := s.installsHelpers.CreateInstall(ctx, appID, &installshelpers.CreateInstallParams{
			Name: proposed.Name,
		})
		if err != nil {
			ctx.Error(fmt.Errorf("unable to create install %s: %w", proposed.Name, err))
			return
		}
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(approval).
		Select("status", "approved_at", "approved_by_id").
		Updates(app.InstallCreationApproval{
			Status:       app.InstallCreationApprovalStatusApproved,
			ApprovedAt:   &now,
			ApprovedByID: &accountID,
		}).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update approval status: %w", err))
		return
	}

	queue, err := s.queueClient.GetQueueByOwnerAndName(ctx, appID, "apps", appshelpers.AppInstallSyncsQueueName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find queue for app: %w", err))
		return
	}

	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   appID,
		OwnerType: "apps",
		Signal: &syncinstalls.Signal{
			AppID:       appID,
			TriggeredBy: "approval",
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to re-trigger sync: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, map[string]string{
		"status": "approved",
	})
}

func (s *service) denyInstallCreation(ctx *gin.Context, approval *app.InstallCreationApproval, accountID string) {
	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(approval).
		Select("status", "approved_at", "approved_by_id").
		Updates(app.InstallCreationApproval{
			Status:       app.InstallCreationApprovalStatusDenied,
			ApprovedAt:   &now,
			ApprovedByID: &accountID,
		}).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update approval status: %w", err))
		return
	}

	syncID := approval.AppInstallConfigSyncID
	if syncID != "" {
		s.db.WithContext(ctx).
			Model(&app.AppInstallConfigSync{}).
			Where(app.AppInstallConfigSync{ID: syncID}).
			Updates(map[string]any{
				"status": app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "approval denied",
				},
			})
	}

	ctx.JSON(http.StatusAccepted, map[string]string{
		"status": "denied",
	})
}
