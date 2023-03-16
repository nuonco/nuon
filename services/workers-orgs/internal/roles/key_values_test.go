package roles

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestKeyValuesKMSKeyPolicy(t *testing.T) {
	arn := uuid.NewString()

	doc, err := KeyValuesKMSKeyPolicy(arn)
	assert.NoError(t, err)

	var policy iamRoleTrustPolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	// assert permissions
	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)
	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.Equal(t, "kms:*", policy.Statement[0].Action)

	// assert principal condition
	assert.Equal(t, arn, policy.Statement[0].Condition.StringLike["aws:PrincipalArn"])
}

func TestKeyValuesIAMPolicy(t *testing.T) {
	bucketName := uuid.NewString()
	orgID := uuid.NewString()

	byts, err := KeyValuesIAMPolicy(bucketName, orgID)
	assert.NoError(t, err)

	var policy iamRolePolicy
	err = json.Unmarshal(byts, &policy)
	assert.NoError(t, err)

	// assert permissions
	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)

	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.Equal(t, "s3:*", policy.Statement[0].Action[0])
	assert.Contains(t, policy.Statement[0].Resource, bucketName)

	assert.Equal(t, "Allow", policy.Statement[1].Effect)
	assert.Equal(t, "kms:*", policy.Statement[1].Action[0])
	assert.Equal(t, "*", policy.Statement[1].Resource)
}

func TestKeyValuesIAMName(t *testing.T) {
	orgID := uuid.NewString()
	iamName := KeyValuesIAMName(orgID)

	assert.Contains(t, iamName, orgID)
	assert.Contains(t, iamName, "org-")
	assert.Contains(t, iamName, "key-values-")
}
