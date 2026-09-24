package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSPolicyNameForInstall(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "inl123-alb-create", (&AppNamedIAMPolicyConfig{
		Name:       "alb-create",
		PolicyName: "alb-create",
	}).AWSPolicyNameForInstall("inl123"))

	assert.Equal(t, "inl123-alb-create", (&AppNamedIAMPolicyConfig{
		Name:       "alb-create",
		PolicyName: "inl123-alb-create",
	}).AWSPolicyNameForInstall("inl123"))

	assert.Equal(t, "alb-create", (&AppNamedIAMPolicyConfig{
		Name: "alb-create",
	}).AWSPolicyNameForInstall(""))
}

func TestNamedIAMPolicyCloudFormationStackName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "NamedPolicyAlbCreate", NamedIAMPolicyCloudFormationStackName("{{.nuon.install.id}}-alb-create"))
	assert.Equal(t, "NamedPolicyLogs", NamedIAMPolicyCloudFormationStackName("logs"))
	assert.Equal(t, "NamedPolicyPolicy", NamedIAMPolicyCloudFormationStackName("{{.nuon.install.id}}"))
}
