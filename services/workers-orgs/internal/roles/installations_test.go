package roles

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInstallationsIAMPolicy(t *testing.T) {
	bucketName := "nuon-org-installations-test"
	orgID := uuid.NewString()
	sandboxARN := "arn:aws:s3:::nuon-sandboxes"
	keyARN := "arn:aws:kms:us-west-2:431927561584:key/4d47db50-340c-4204-a02b-7e10a37308a2"

	doc, err := InstallationsIAMPolicy(bucketName, orgID, sandboxARN, keyARN)
	assert.NoError(t, err)

	t.Log(string(doc))

	var policy iamRolePolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)
	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.Equal(t, "s3:*", policy.Statement[0].Action[0])
	assert.Contains(t, policy.Statement[0].Resource, "orgID="+orgID)
	assert.Contains(t, policy.Statement[0].Resource, bucketName)
	assert.Equal(t, policy.Statement[1].Resource, sandboxARN)
	assert.Contains(t, policy.Statement[2].Resource, sandboxARN)
	assert.Equal(t, policy.Statement[3].Resource, keyARN)
}

func TestInstallationsIAMName(t *testing.T) {
	orgID := uuid.NewString()
	iamName := InstallationsIAMName(orgID)

	assert.Contains(t, iamName, orgID)
	assert.Contains(t, iamName, "org-")
	assert.Contains(t, iamName, "installations-")
}

func TestInstallationsIAMTrustPolicy(t *testing.T) {
	doc, err := InstallationsIAMTrustPolicy(
		"arn:aws:iam::676549690856:role/eks/eks-workers-*",
		"arn:aws:iam::766121324316:role/nuon-internal-support-stage",
		"arn:aws:iam::766121324316:role/odr",
	)
	assert.NoError(t, err)

	var policy iamRoleTrustPolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	// TODO(jm): add better tests of the actual trust policy once we know what we have works
}
