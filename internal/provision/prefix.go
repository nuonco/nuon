package provision

import "fmt"

func getS3Prefix(req ProvisionRequest) string {
	return fmt.Sprintf(
		"instances/org=%s/app=%s/install=%s/component=%s",
		req.OrgID,
		req.AppID,
		req.InstallID,
		req.Component.Name,
	)
}
