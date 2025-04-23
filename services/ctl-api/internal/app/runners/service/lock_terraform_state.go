package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (s *service) LockTerraformState(ctx *gin.Context) {
	var lock app.TerraformLock
	if err := ctx.BindJSON(&lock); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	state, err := s.helpers.GetTerraformState(ctx, workspaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(err)
		return
	}

	if state == nil {
		state = &app.TerraformState{
			TerraformWorkspaceID: workspaceID,
			Data:                 nil,
		}
	}
	state.Lock = &lock

	_, err = s.helpers.InsertTerraformState(ctx, workspaceID, state.Lock, state.Data)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update terraform state: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "")
}
