package start

import (
	"fmt"
)

// getS3Prefix returns the prefix to be used for the plan and it's encompassed files
func getS3Prefix(orgID, appID, componentName, deploymentID string) string {
	return fmt.Sprintf("org=%s/app=%s/component=%s/deployment=%s",
		orgID,
		appID,
		componentName,
		deploymentID)
}
