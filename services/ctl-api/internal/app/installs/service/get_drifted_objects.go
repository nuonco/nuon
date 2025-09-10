package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetDriftedObjects
// @Summary				get drifted objects for an install
// @Description.markdown	get_drifted_objects.md
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.DriftedObject
// @Router					/v1/installs/{install_id}/drifted-objects [get]
func (s *service) GetDriftedObjects(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	driftedObjects, err := s.findDriftedObjects(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get drifted objects for install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, driftedObjects)
}

func (s *service) findDriftedObjects(ctx context.Context, orgID, installID string) ([]app.DriftedObject, error) {
	var driftedObjects []app.DriftedObject
	res := s.db.WithContext(ctx).
		Where("org_id = ? AND install_id = ?", orgID, installID).
		Find(&driftedObjects)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get drifted objects: %w", res.Error)
	}

	return driftedObjects, nil
}
