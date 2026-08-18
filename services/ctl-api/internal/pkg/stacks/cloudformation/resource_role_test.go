package cloudformation

import (
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func roleConfig(stackName string, roleType app.AWSIAMRoleType) app.AppAWSIAMRoleConfig {
	return app.AppAWSIAMRoleConfig{
		Type:                         roleType,
		Name:                         stackName,
		CloudFormationStackName:      stackName,
		CloudFormationStackParamName: stackName + "Enabled",
	}
}

func rolesTestInput() *stacks.TemplateInput {
	inp := &stacks.TemplateInput{
		Install:  &app.Install{ID: "inl123"},
		AppCfg:   &app.AppConfig{},
		Runner:   &app.Runner{ID: "run123"},
		Settings: &app.RunnerGroupSettings{},
	}
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		roleConfig("ProvisionRole", app.AWSIAMRoleTypeRunnerProvision),
	}
	inp.AppCfg.BreakGlassConfig.Roles = []app.AppAWSIAMRoleConfig{
		roleConfig("BreakGlassRole", app.AWSIAMRoleTypeBreakGlass),
	}
	inp.AppCfg.PermissionsConfig.CustomRoles = []app.AppAWSIAMRoleConfig{
		roleConfig("BucketCleanupRole", app.AWSIAMRoleTypeCustom),
	}

	return inp
}

// The runner instance role's identity policy scopes sts:AssumeRole with a condition on
// this tag, so every role the runner is meant to assume has to carry it — including
// break-glass roles, which share the trust policy that names the runner principal.
func TestGetRolesResources_RunnerAssumableTag(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	assumable := tags.Tag{Key: TagKeyRunnerAssumable, Value: "true"}

	t.Run("every role family carries the tag", func(t *testing.T) {
		rsrcs := tpl.getRolesResources(rolesTestInput(), tagBuilder{installID: "inl123"})

		for _, stackName := range []string{"ProvisionRole", "BreakGlassRole", "BucketCleanupRole"} {
			role, ok := rsrcs[stackName].(*iam.Role)
			require.True(t, ok, "expected an IAM role at %s", stackName)
			assert.Contains(t, role.Tags, assumable)
		}
	})

	t.Run("roles also carry the default entity tags", func(t *testing.T) {
		tb := tagBuilder{installID: "inl123", orgID: "org123", appID: "app123"}

		rsrcs := tpl.getRolesResources(rolesTestInput(), tb)

		role, ok := rsrcs["ProvisionRole"].(*iam.Role)
		require.True(t, ok)

		applied := tagMap(role.Tags)
		assert.Equal(t, "inl123", applied["install.nuon.co/id"])
		assert.Equal(t, "org123", applied["org.nuon.co/id"])
		assert.Equal(t, "app123", applied["app.nuon.co/id"])
		assert.Equal(t, "true", applied[TagKeyRunnerAssumable])
	})

	t.Run("customer tags cannot clear the tag", func(t *testing.T) {
		tb := tagBuilder{
			installID:  "inl123",
			additional: map[string]string{TagKeyRunnerAssumable: "false"},
		}

		rsrcs := tpl.getRolesResources(rolesTestInput(), tb)

		role, ok := rsrcs["ProvisionRole"].(*iam.Role)
		require.True(t, ok)
		assert.Contains(t, role.Tags, assumable)
		assert.NotContains(t, role.Tags, tags.Tag{Key: TagKeyRunnerAssumable, Value: "false"})
	})
}

// The tag marks assume-role targets, so it must stay on the role resource rather than
// moving into tagBuilder.apply, which every tagged resource in the stack goes through.
func TestGetRunnerPhoneHomeLambdaRole_NotRunnerAssumable(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	assert.NotContains(t, role.Tags, tags.Tag{Key: TagKeyRunnerAssumable, Value: "true"})
}
