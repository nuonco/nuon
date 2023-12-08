package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

type DeprovisionCanaryRequest struct {
	CanaryID string `json:"canary_id"`
}

//	@BasePath	/v1/general
//
// Deprovision a canary
//
//	@Summary	provision a canary
//	@Schemes
//	@Description	provision a canary
//	@Param			req	body	DeprovisionCanaryRequest	true	"Input"
//	@Tags			general/admin
//	@Accept			json
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/general/deprovision-canary [post]
func (c *service) DeprovisionCanary(ctx *gin.Context) {
	var req DeprovisionCanaryRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	wkfowReq := &canaryv1.DeprovisionRequest{
		CanaryId: req.CanaryID,
	}

	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-deprovision", req.CanaryID),
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"canary-id":  wkfowReq.CanaryId,
			"started-by": "ctl-api",
		},
	}

	_, err := c.temporalClient.ExecuteWorkflowInNamespace(ctx, "canary", opts, "Deprovision", wkfowReq)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to deprovision canary: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
