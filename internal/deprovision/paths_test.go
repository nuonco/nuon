package deprovision

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getSandboxBucketKey(t *testing.T) {
	sandboxName := "aws-eks"
	sandboxVersion := "v0.0.1"

	expected := fmt.Sprintf("sandboxes/%s_%s.tar.gz", sandboxName, sandboxVersion)
	assert.Equal(t, expected, getSandboxBucketKey(sandboxName, sandboxVersion))
}

func Test_getStateBucketKey(t *testing.T) {
	orgID := "org123"
	appID := "app123"
	installID := "install123"

	expected := fmt.Sprintf("org=%s/app=%s/install=%s/%s", orgID, appID, installID, defaultStateFilename)
	assert.Equal(t, expected, getStateBucketKey(orgID, appID, installID))
}

func Test_getS3Path(t *testing.T) {
	bucket := "nuon-installations-stage"
	orgID := "org123"
	appID := "app123"
	installID := "install123"

	expected := fmt.Sprintf("s3://%s/org=%s/app=%s/install=%s", bucket, orgID, appID, installID)
	assert.Equal(t, expected, getS3Prefix(bucket, orgID, appID, installID))
}
