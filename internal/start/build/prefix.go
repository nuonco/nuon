package build

import (
	"fmt"

	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
)

// TODO(jm): fix this once we have a better way of outputting the prefix
func getS3Prefix(req *buildv1.BuildRequest) string {
	return fmt.Sprintf(
		"deployments/org=%s/app=%s/deployment=%s/component=%s",
		req.OrgId,
		req.AppId,
		req.DeploymentId,
		"httpbin",
	)
}
