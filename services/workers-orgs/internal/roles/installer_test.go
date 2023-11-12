package roles

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInstallerIAMPolicy(t *testing.T) {
	bucketName := "nuon-org-installations-test"
	orgID := uuid.NewString()

	doc, err := InstallerIAMPolicy(bucketName, orgID)
	assert.NoError(t, err)

	var policy iamRolePolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)

	// assert that installations bucket perms exist
	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.Equal(t, "s3:*", policy.Statement[0].Action[0])
	assert.Contains(t, policy.Statement[0].Resource, "orgID="+orgID)
	assert.Contains(t, policy.Statement[0].Resource, bucketName)

	// assert that assume role exists
	assert.Equal(t, "Allow", policy.Statement[1].Effect)
	assert.Equal(t, "sts:AssumeRole", policy.Statement[1].Action[0])
	assert.Equal(t, "*", policy.Statement[1].Resource)
}

func TestInstallerIAMName(t *testing.T) {
	orgID := uuid.NewString()
	iamName := InstallerIAMName(orgID)

	assert.Contains(t, iamName, orgID)
	assert.Contains(t, iamName, "org-")
	assert.Contains(t, iamName, "installer-")
}

func TestInstallerIAMTrustPolicy(t *testing.T) {
	doc, err := InstallerIAMTrustPolicy(
		"arn:aws:iam::676549690856:role/eks/eks-workers-*",
		"arn:aws:iam::766121324316:role/nuon-internal-support-stage",
		"arn:aws:iam::766121324316:role/runner",
	)
	assert.NoError(t, err)

	var policy iamRoleTrustPolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	// TODO(jm): add better tests of the actual trust policy once we know what we have works
}
