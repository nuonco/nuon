package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

	orgIDs, err := normalizeReconcileOrgIDs(req.OrgIDs)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.l.Info("reconciling org roles",
		zap.Bool("all_orgs", len(orgIDs) == 0),
		zap.Int("requested", len(orgIDs)),
	)

	reconciled, err := s.reconcileOrgRoles(ctx, orgIDs)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.l.Info("reconciled org roles", zap.Int("reconciled", reconciled))
	ctx.JSON(http.StatusOK, AdminReconcileOrgRolesResponse{Reconciled: reconciled})
}

func normalizeReconcileOrgIDs(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			return nil, stderr.NewInvalidRequest(fmt.Errorf("org_ids contains an empty id"))
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func (s *service) reconcileOrgRoles(ctx context.Context, orgIDs []string) (int, error) {
	if len(orgIDs) > 0 {
		orgs, err := s.loadOrgsByID(ctx, orgIDs)
		if err != nil {
			return 0, err
		}
		return reconcileOrgs(ctx, s.db, orgs)
	}

	const batchSize = 50
	var reconciled int
	var lastID string
	var lastCreatedAt time.Time
	for {
		var orgs []app.Org
		q := s.db.WithContext(ctx).
			Select("id", "created_by_id", "created_at").
			Order("created_at ASC").
			Order("id ASC").
			Limit(batchSize)
		if lastID != "" {
			q = q.Where("(created_at > ?) OR (created_at = ? AND id > ?)", lastCreatedAt, lastCreatedAt, lastID)
		}
		if err := q.Find(&orgs).Error; err != nil {
			return reconciled, fmt.Errorf("unable to load orgs: %w", err)
		}
		if len(orgs) == 0 {
			return reconciled, nil
		}

		n, err := reconcileOrgs(ctx, s.db, orgs)
		reconciled += n
		if err != nil {
			return reconciled, err
		}
		last := orgs[len(orgs)-1]
		lastID = last.ID
		lastCreatedAt = last.CreatedAt
	}
}

func (s *service) loadOrgsByID(ctx context.Context, orgIDs []string) ([]app.Org, error) {
	var orgs []app.Org
	if err := s.db.WithContext(ctx).
		Select("id", "created_by_id").
		Where("id IN ?", orgIDs).
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("unable to load orgs: %w", err)
	}

	found := make(map[string]struct{}, len(orgs))
	for _, org := range orgs {
		found[org.ID] = struct{}{}
	}
	var missing []string
	for _, id := range orgIDs {
		if _, ok := found[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, stderr.ErrNotFound{
			Err:         fmt.Errorf("orgs not found: %s", strings.Join(missing, ", ")),
			Description: "one or more org ids were not found",
		}
	}
	return orgs, nil
}

func reconcileOrgs(ctx context.Context, db *gorm.DB, orgs []app.Org) (int, error) {
	for i := range orgs {
		if err := authz.ReconcileOrgRoles(ctx, db, orgs[i]); err != nil {
			return i, fmt.Errorf("unable to reconcile roles for org %s: %w", orgs[i].ID, err)
		}
	}
	return len(orgs), nil
}
