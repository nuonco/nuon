package roles

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInstancesIAMPolicy(t *testing.T) {
	orgID := uuid.NewString()

	doc, err := InstancesIAMPolicy(orgID)
	assert.NoError(t, err)

	var policy iamRolePolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)
	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.Equal(t, "ecr:*", policy.Statement[0].Action[0])
	assert.Contains(t, policy.Statement[0].Resource, "*")
}

func TestInstancesIAMName(t *testing.T) {
	orgID := uuid.NewString()
	iamName := InstancesIAMName(orgID)

	assert.Contains(t, iamName, orgID)
	assert.Contains(t, iamName, "org-")
	assert.Contains(t, iamName, "instances-")
}

func TestInstancesIAMTrustPolicy(t *testing.T) {
	doc, err := InstancesIAMTrustPolicy("arn:aws:iam::676549690856:role/eks/eks-workers-*", "arn:aws:iam::766121324316:role/nuon-internal-support-stage", "runner-role")
	assert.NoError(t, err)

	var policy iamRoleTrustPolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	// TODO(jm): add better tests of the actual trust policy once we know what we have works
}
