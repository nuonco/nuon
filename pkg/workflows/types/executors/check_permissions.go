package executors

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	CheckPermissionsWorkflowName string = "CheckPermissions"
)

type CheckPermissionsRequest struct {
	AWSSettings   *AWSSettings   `json:"aws_settings" temporaljson:"aws_settings"`
	AzureSettings *AzureSettings `json:"azure_settings" temporaljson:"azure_settings"`

	Metadata Metadata `json:"metadata" temporaljson:"metadata"`
}

type CheckPermissionsResponse struct{}

func CheckPermissionsIDCallback(req *CheckPermissionsRequest) string {
	return fmt.Sprintf("%s-%s-%s", req.Metadata.OrgID, req.Metadata.AppID, req.Metadata.InstallID)
}

// @id-callback CheckPermissionsIDCallback
func CheckPermissions(workflow.Context, *CheckPermissionsRequest) (*CheckPermissionsResponse, error) {
	panic("this should not be executed directly, and is only used to generate an await function.")
	return nil, nil
}
