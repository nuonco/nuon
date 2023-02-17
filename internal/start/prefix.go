package start

import (
	"github.com/powertoolsdev/go-workflows-meta/prefix"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
)

// getS3Prefix returns the prefix to be used for the plan and it's encompassed files
func getS3Prefix(orgID, appID, componentID, deploymentID string) string {
	return prefix.DeploymentPath(orgID, appID, componentID, deploymentID)
}

func getS3PrefixFromRequest(req *deploymentsv1.StartRequest) string {
	return getS3Prefix(req.OrgId, req.AppId, req.Component.Id, req.DeploymentId)
}
