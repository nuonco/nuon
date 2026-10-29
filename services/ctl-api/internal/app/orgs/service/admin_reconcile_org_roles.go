package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

type AdminReconcileOrgRolesRequest struct {
	OrgIDs []string `json:"org_ids,omitempty"`
}

type AdminReconcileOrgRolesResponse struct {
	Reconciled int `json:"reconciled"`
}

// @ID			AdminReconcileOrgRoles
// @Summary		reconcile managed roles for orgs
// @Description	Creates any missing managed roles and updates role metadata (title, description, contexts, managed) to match the standard definitions. Existing role policies are left unchanged. When org_ids is omitted or empty, every org is reconciled. Idempotent.
// @Tags			orgs/admin
// @Security		AdminEmail
// @Accept			json
// @Param			req	body	AdminReconcileOrgRolesRequest	true	"Input"
// @Produce		json
// @Success		200	{object}	AdminReconcileOrgRolesResponse
// @Router			/v1/orgs/admin-reconcile-roles [POST]
func (s *service) AdminReconcileOrgRoles(ctx *gin.Context) {
	var req AdminReconcileOrgRolesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	s.l.Info("reconciling org roles",
		zap.Bool("all_orgs", len(req.OrgIDs) == 0),
		zap.Int("requested", len(req.OrgIDs)),
	)

	reconciled, err := s.reconcileOrgRoles(ctx, req.OrgIDs)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.l.Info("reconciled org roles", zap.Int("reconciled", reconciled))
	ctx.JSON(http.StatusOK, AdminReconcileOrgRolesResponse{Reconciled: reconciled})
}

func (s *service) reconcileOrgRoles(ctx context.Context, orgIDs []string) (int, error) {
	q := s.db.WithContext(ctx).
		Select("id", "created_by_id").
		Order("created_at ASC, id ASC")
	if len(orgIDs) > 0 {
		q = q.Where("id IN ?", orgIDs)
	}

	var orgs []app.Org
	if err := q.Find(&orgs).Error; err != nil {
		return 0, fmt.Errorf("unable to load orgs: %w", err)
	}
	return reconcileOrgs(ctx, s.db, orgs)
}

func reconcileOrgs(ctx context.Context, db *gorm.DB, orgs []app.Org) (int, error) {
	for i := range orgs {
		if err := authz.ReconcileOrgRoles(ctx, db, orgs[i]); err != nil {
			return i, fmt.Errorf("unable to reconcile roles for org %s: %w", orgs[i].ID, err)
		}
	}
	return len(orgs), nil
}
