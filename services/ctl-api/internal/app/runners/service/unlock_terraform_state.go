package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *service) UnlockTerraformState(ctx *gin.Context) {
	sid, err := s.GetStateID(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get state ID: %w", err))
		return
	}
	var lock app.TerraformLock
	if err := json.NewDecoder(ctx.Request.Body).Decode(&lock); err != nil {
		ctx.Error(fmt.Errorf("unable to decode lock: %w", err))
		return
	}

	currentState, err := s.validateTerraformStateLock(ctx, sid, lock.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if currentState == nil || string(currentState.Lock) == "" {
		ctx.JSON(http.StatusOK, "")
		return
	}

	currentState.Lock = nil
	err = s.helpers.InsertTerraformState(ctx, sid, currentState)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update terraform state: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, "")
}
