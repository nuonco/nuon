package iam

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-generics"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/stretchr/testify/assert"
)

func Test_installationsIAMPolicy(t *testing.T) {
	bucketName := "nuon-org-installations-test"
	orgID := uuid.NewString()

	doc, err := installationsIAMPolicy(bucketName, orgID)
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

func Test_installationsIAMName(t *testing.T) {
	orgID := uuid.NewString()
	iamName := installationsIAMName(orgID)

	assert.Contains(t, iamName, orgID)
	assert.Contains(t, iamName, "org-")
	assert.Contains(t, iamName, "installations-")
}

func Test_installationsIAMTrustPolicy(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	doc, err := installationsIAMTrustPolicy(cfg)
	assert.NoError(t, err)

	var policy iamRoleTrustPolicy
	err = json.Unmarshal(doc, &policy)
	assert.NoError(t, err)

	// TODO(jm): add better tests of the actual trust policy once we know what we have works
}
