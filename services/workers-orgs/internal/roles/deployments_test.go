package roles

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentsIAMPolicy(t *testing.T) {
	bucketName := "nuon-org-deployments-test"
	orgID := uuid.NewString()

	doc, err := DeploymentsIAMPolicy(bucketName, orgID)
	assert.NoError(t, err)

	var policy iamRolePolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)
	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.Equal(t, "s3:*", policy.Statement[0].Action[0])
	assert.Contains(t, policy.Statement[0].Resource, "orgID="+orgID)
	assert.Contains(t, policy.Statement[0].Resource, bucketName)
}

func TestDeploymentsIAMName(t *testing.T) {
	orgID := uuid.NewString()
	iamName := DeploymentsIAMName(orgID)

	assert.Contains(t, iamName, orgID)
	assert.Contains(t, iamName, "org-")
	assert.Contains(t, iamName, "deployments-")
}

func TestDeploymentsIAMTrustPolicy(t *testing.T) {
	doc, err := DeploymentsIAMTrustPolicy("arn:aws:iam::676549690856:role/eks/eks-workers-*",
		"arn:aws:iam::766121324316:role/nuon-internal-support-stage",
		"arn:aws:iam::766121324316:role/org-odr-runner",
	)
	assert.NoError(t, err)

	var policy iamRoleTrustPolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	// TODO(jm): add better tests of the actual trust policy once we know what we have works
}
