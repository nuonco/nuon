package executors

import (
	"fmt"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"go.temporal.io/sdk/workflow"
)

const (
	CheckPermissionsWorkflowName string = "CheckPermissions"
)

type CheckPermissionsResponse struct{}

func CheckPermissionsIDCallback(req *installsv1.ProvisionRequest) string {
	return fmt.Sprintf("%s-%s-%s", req.OrgId, req.AppId, req.InstallId)
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 1m
// @task-queue "executors"
// @id-callback CheckPermissionsIDCallback
func CheckPermissions(workflow.Context, *installsv1.ProvisionRequest) (*CheckPermissionsResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}
