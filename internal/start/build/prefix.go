package build

import "fmt"

func getS3Prefix(req BuildRequest) string {
	return fmt.Sprintf(
		"deployments/org=%s/app=%s/deployment=%s/component=%s",
		req.OrgID,
		req.AppID,
		req.DeploymentID,
		req.Component.Name,
	)
}
