package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	infratestsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/infra_tests/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

type ProvisionInfraTestRequest struct {
	SandboxName string `json:"sandbox_name"`
}

// @ID						ProvisionInfraTest
// @Summary					provision an infra test
// @Description.markdown	provision_infra_test.md
// @Param					req	body	ProvisionInfraTestRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce					json
// @Success					201	{string}	ok
// @Router					/v1/general/provision-canary [post]
func (c *service) ProvisionInfraTest(ctx *gin.Context) {
	var req ProvisionInfraTestRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	infraTestID := domains.NewInfraTestID()
	wkflowReq := &infratestsv1.TestSandboxRequest{
		SandboxName: req.SandboxName,
	}

	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-infra-test", infraTestID),
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"sandbox-name": req.SandboxName,
			"started-by":   "ctl-api",
		},
	}

	_, err := c.temporalClient.ExecuteWorkflowInNamespace(ctx, "infra-tests", opts, "TestSandbox", wkflowReq)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to provision infra-tests: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
