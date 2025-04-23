package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (s *service) UnlockTerraformState(ctx *gin.Context) {
	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	var lock app.TerraformLock
	if err := ctx.BindJSON(&lock); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	state, err := s.helpers.GetTerraformState(ctx, workspaceID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get terraform state"))
		return
	}

	if _, err := s.helpers.InsertTerraformState(ctx, workspaceID, nil, state.Data); err != nil {
		ctx.Error(errors.Wrap(err, "unable to insert state"))
		return
	}

	ctx.JSON(http.StatusOK, "")
}
