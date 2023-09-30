package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

type ProvisionCanaryRequest struct{}

//	@BasePath	/v1/general
//
// Provision a canary
//
//	@Summary	provision a canary
//	@Schemes
//	@Description	provision a canary
//	@Param			req	body	ProvisionCanaryRequest	true	"Input"
//	@Tags			general/admin
//	@Accept			json
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/general/provision-canary [post]
func (c *service) ProvisionCanary(ctx *gin.Context) {
	req := &canaryv1.ProvisionRequest{
		CanaryId:    domains.NewCanaryID(),
		Deprovision: true,
	}

	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"canary-id":  req.CanaryId,
			"started-by": "ctl-api",
		},
	}

	_, err := c.temporalClient.ExecuteWorkflowInNamespace(ctx, "canary", opts, "Provision", req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to provision canary: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
