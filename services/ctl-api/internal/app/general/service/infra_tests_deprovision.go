package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

// consider making this a shared type
type InfraTestsDeprovisionRequest struct {
	SandboxName      string                 `json:"sandbox_name"`
	SandboxRef       string                 `json:"sandbox_ref"`
	TerraformVersion string                 `json:"terraform_version"`
	Region           string                 `json:"region"`
	SandboxVars      map[string]interface{} `json:"sandbox_vars"`
	OrgID            string                 `json:"org_id"`
	CanaryID         string                 `json:"canary_id"`
	Directory        string                 `json:"directory"`
	ClusterName      string                 `json:"cluster_name"`
	Account          map[string]interface{} `json:"account"`
	Profile          string                 `json:"profile"`
}

// @ID						InfraTestsDeprovision
// @Summary					deprovision an infra test
// @Description.markdown	infra_tests_deprovision.md
// @Param					req	body	InfraTestsDeprovisionRequest	true	"Input"
// @Tags					general/admin
// @Accept					json
// @Produce					json
// @Success					201	{string}	ok
// @Router					/v1/general/infra-tests/deprovision [post]
func (c *service) InfraTestsDeprovision(ctx *gin.Context) {
	var req InfraTestsDeprovisionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-deprovision", req.CanaryID),
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"sandbox-name": req.SandboxName,
			"started-by":   "ctl-api",
		},
	}

	_, err := c.temporalClient.ExecuteWorkflowInNamespace(ctx, "infra-tests", opts, "Deprovision", req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to provision infra-tests: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}
