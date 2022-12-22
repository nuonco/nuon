package provision

import (
	"fmt"
)

func getInstallationPrefix(orgShortID, appShortID, installShortID string) string {
	return fmt.Sprintf("org=%s/app=%s/install=%s", orgShortID, appShortID, installShortID)
}

func getS3Prefix(bucket, orgID, appID, installID string) string {
	return fmt.Sprintf("s3://%s/org=%s/app=%s/install=%s", bucket, orgID, appID, installID)
}
