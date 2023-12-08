package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StopCanaryCronRequest struct {
	SandboxMode bool `json:"sandbox_mode"`
}

//	@BasePath	/v1/general
//
// Stop a canary
//
//	@Summary	stop canary cron
//	@Schemes
//	@Description	stop canary cron
//	@Param			req	body	StopCanaryCronRequest	true	"Input"
//	@Tags			general/admin
//	@Accept			json
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/general/stop-canary-cron [post]
func (c *service) StopCanaryCron(ctx *gin.Context) {
	var req StopCanaryCronRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	if err := c.stopCanaryCron(ctx, realCanaryCronID); err != nil {
		ctx.Error(fmt.Errorf("unable to stop sandbox cron: %w", err))
		return
	}

	if err := c.stopCanaryCron(ctx, sandboxCanaryCronID); err != nil {
		ctx.Error(fmt.Errorf("unable to stop sandbox cron: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (c *service) stopCanaryCron(ctx context.Context, id string) error {
	if err := c.temporalClient.CancelWorkflowInNamespace(ctx, "canary", id, ""); err != nil {
		return fmt.Errorf("unable to stop canary: %w", err)
	}

	return nil
}
