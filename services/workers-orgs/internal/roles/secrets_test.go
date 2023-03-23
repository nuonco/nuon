package roles

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestKeyValuesKMSKeyPolicy(t *testing.T) {
	arn := uuid.NewString()
	currentServiceRoleARN := uuid.NewString()
	rootAccountRoleARN := uuid.NewString()

	doc, err := SecretsKMSKeyPolicy(arn, currentServiceRoleARN, rootAccountRoleARN)
	assert.NoError(t, err)

	var policy iamRoleTrustPolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	// assert kms admin permissions
	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)
	assert.Equal(t, "Allow", policy.Statement[0].Effect)
	assert.NotEmpty(t, policy.Statement[0].Action)
	// assert principal condition
	assert.Equal(t, currentServiceRoleARN, policy.Statement[0].Principal.AWS)

	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)
	assert.Equal(t, "Allow", policy.Statement[1].Effect)
	assert.NotEmpty(t, policy.Statement[1].Action)
	// assert principal condition
	assert.Equal(t, rootAccountRoleARN, policy.Statement[1].Principal.AWS)

	// assert kms permissions
	assert.Equal(t, defaultIAMPolicyVersion, policy.Version)
	assert.Equal(t, "Allow", policy.Statement[2].Effect)
	assert.Equal(t, "kms:*", policy.Statement[2].Action[0])
	// assert principal condition
	assert.Equal(t, arn, policy.Statement[2].Condition.StringLike["aws:PrincipalArn"])
}

func TestKeyValuesIAMPolicy(t *testing.T) {
	bucketName := uuid.NewString()
	orgID := uuid.NewString()

	byts, err := SecretsIAMPolicy(bucketName, orgID)
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
	iamName := SecretsIAMName(orgID)

	assert.Contains(t, iamName, orgID)
	assert.Contains(t, iamName, "org-")
	assert.Contains(t, iamName, "secrets-")
}
