package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/pkg/types/workflows/canary"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	tclient "go.temporal.io/sdk/client"
)

type ProvisionCanaryRequest struct {
	SandboxMode bool `json:"sandbox_mode"`
}

// @ID						ProvisionCanary
// @Summary				provision a canary
// @Description.markdown	provision_canary.md
// @Param					req	body	ProvisionCanaryRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Deprecated
// @Router					/v1/general/provision-canary [post]
func (c *service) ProvisionCanary(ctx *gin.Context) {
	var req ProvisionCanaryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	canaryID := domains.NewCanaryID()
	wkfowReq := &canary.ProvisionRequest{
		CanaryId:    canaryID,
		SandboxMode: req.SandboxMode,
	}

	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-provision", canaryID),
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"canary-id":  wkfowReq.CanaryId,
			"started-by": "ctl-api",
		},
	}

	_, err := c.temporalClient.ExecuteWorkflowInNamespace(ctx, "canary", opts, "Provision", wkfowReq)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to provision canary: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
