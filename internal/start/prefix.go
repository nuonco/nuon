package start

import (
	"fmt"

	"github.com/powertoolsdev/go-common/shortid"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
)

// getS3Prefix returns the prefix to be used for the plan and it's encompassed files
func getS3Prefix(orgID, appID, componentName, deploymentID string) string {
	return fmt.Sprintf("org=%s/app=%s/component=%s/deployment=%s",
		orgID,
		appID,
		componentName,
		deploymentID)
}

func getS3PrefixFromRequest(req *deploymentsv1.StartRequest) (string, error) {
	shortIDs, err := shortid.ParseStrings(req.OrgId, req.AppId, req.DeploymentId)
	if err != nil {
		return "", fmt.Errorf("unable to parse short IDs: %w", err)
	}

	orgID, appID, deploymentID := shortIDs[0], shortIDs[1], shortIDs[2]
	return getS3Prefix(orgID, appID, req.Component.Name, deploymentID), nil
}
