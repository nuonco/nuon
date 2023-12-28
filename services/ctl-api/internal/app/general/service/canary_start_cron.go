package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

const (
	sandboxCanaryCron   = "0 */6 * * *"
	sandboxCanaryCronID = "canary-cron-sandbox"

	realCanaryCron   = "0 16 * * *"
	realCanaryCronID = "canary-cron"
)

type StartCanaryCronRequest struct {
	SandboxMode bool `json:"sandbox_mode"`
}

// @ID StartCanaryCron
// @Summary	start canary cron
// @Description.markdown	start_canary_cron.md
// @Param			req	body	StartCanaryCronRequest	true	"Input"
// @Tags			general/admin
// @Accept			json
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/general/start-canary-cron [post]
func (c *service) StartCanaryCron(ctx *gin.Context) {
	var req StartCanaryCronRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	if err := c.startCanaryCron(ctx, sandboxCanaryCronID, true, sandboxCanaryCron); err != nil {
		ctx.Error(fmt.Errorf("unable to create sandbox cron: %w", err))
		return
	}
	if err := c.startCanaryCron(ctx, realCanaryCronID, true, realCanaryCron); err != nil {
		ctx.Error(fmt.Errorf("unable to create sandbox cron: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (c *service) startCanaryCron(ctx context.Context, id string, sandboxMode bool, schedule string) error {
	opts := tclient.StartWorkflowOptions{
		ID:           id,
		CronSchedule: schedule,
		TaskQueue:    workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"started-by": "ctl-api",
		},
	}
	wkflowReq := &canaryv1.ProvisionRequest{
		SandboxMode: sandboxMode,
	}

	_, err := c.temporalClient.ExecuteWorkflowInNamespace(ctx, "canary", opts, "Provision", wkflowReq)
	if err != nil {
		return fmt.Errorf("unable to provision canary: %w", err)
	}

	return nil
}
