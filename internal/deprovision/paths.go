package deprovision

import "fmt"

func getSandboxBucketKey(name, version string) string {
	return fmt.Sprintf("sandboxes/%s_%s.tar.gz", name, version)
}

func getStateBucketKey(orgID, appID, installID string) string {
	return fmt.Sprintf("org=%s/app=%s/install=%s/%s", orgID, appID, installID, defaultStateFilename)
}

func getS3Prefix(bucket, orgID, appID, installID string) string {
	return fmt.Sprintf("s3://%s/org=%s/app=%s/install=%s", bucket, orgID, appID, installID)
}
