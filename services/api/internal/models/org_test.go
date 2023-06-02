package models

import (
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/stretchr/testify/assert"
)

func TestOrg_ToProvisionRequest(t *testing.T) {
	org := generics.GetFakeObj[*Org]()
	org.ID = "org1234567890123456798012"
	expected := orgsv1.SignupRequest{
		OrgId:  org.ID,
		Region: "us-west-2",
	}
	actual := org.ToProvisionRequest()
	assert.Equal(t, expected.OrgId, actual.OrgId)
	assert.Equal(t, expected.Region, actual.Region)
}

func TestOrg_ToDeprovisionRequest(t *testing.T) {
	org := generics.GetFakeObj[*Org]()
	org.ID = "org1234567890123456798012"
	expected := orgsv1.TeardownRequest{
		OrgId:  org.ID,
		Region: "us-west-2",
	}
	actual := org.ToDeprovisionRequest()
	assert.Equal(t, expected.OrgId, actual.OrgId)
	assert.Equal(t, expected.Region, actual.Region)
}
