package cloudformation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func phoneHomeTestInput(installID string) *stacks.TemplateInput {
	return &stacks.TemplateInput{
		Install: &app.Install{ID: installID},
		AppCfg:  &app.AppConfig{},
		CloudFormationStackVersion: &app.InstallStackVersion{
			PhoneHomeURL: "https://example.com/phone-home",
		},
	}
}

// The role name is load-bearing: a cross-account grant names this principal, and IAM
// happily accepts a policy referencing a role that does not exist, so a mismatch only
// shows up as an AccessDeniedException at phone-home time. Assert the literal.
func TestGetRunnerPhoneHomeLambdaRole_DeterministicName(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	require.NotNil(t, role.RoleName)
	assert.Equal(t, "instabcdefghijklmnopqrstuv-phone-home", *role.RoleName)
	assert.Equal(t, stacks.PhoneHomeRoleName(inp.Install.ID), *role.RoleName)
}

// IAM caps role names at 64 characters. Install IDs are 26, so the suffix leaves
// plenty of room — this guards against the suffix growing past the limit.
func TestPhoneHomeRoleName_WithinIAMLimit(t *testing.T) {
	name := stacks.PhoneHomeRoleName(strings.Repeat("i", 26))

	assert.LessOrEqual(t, len(name), 64,
		"IAM role names cannot exceed 64 characters")
	assert.Equal(t, strings.Repeat("i", 26)+"-phone-home", name)
}

// The role name must not leak into the phone-home payload. phonehome.py POSTs every
// ResourceProperty it receives, and anything echoed there lands in the install's
// stack outputs.
func TestGetRunnerPhoneHomeProps_DoesNotEchoRoleName(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	props := tpl.getRunnerPhoneHomeProps(inp, nil)

	require.NotNil(t, props)
	for key, value := range props.Properties {
		if str, ok := value.(string); ok {
			assert.NotContains(t, str, "-phone-home",
				"property %q echoes the phone-home role name", key)
		}
	}
}
