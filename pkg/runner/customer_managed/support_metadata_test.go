package customermanaged

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCaptureInputsRedactsSecretsAndClassifiesValues(t *testing.T) {
	now := time.Now().UTC()
	captured := CaptureInputs([]InputSpec{
		{Name: "region", Type: "string", Default: "us-west-2", Bindable: true},
		{Name: "replicas", Type: "number", Required: true, Bindable: true},
		{Name: "token", Type: "string", Default: "vendor-secret", Secret: true},
		{Name: "fixed", Type: "string", Default: "published", Bindable: false},
	}, map[string]string{"replicas": "3", "token": "customer-secret"}, now)

	require.Equal(t, now, captured.ObservedAt)
	require.Equal(t, "default", captured.Inputs[0].ValueStatus)
	require.Equal(t, "us-west-2", *captured.Inputs[0].Value)
	require.Equal(t, "provided", captured.Inputs[1].ValueStatus)
	require.Equal(t, "3", *captured.Inputs[1].Value)
	require.Equal(t, "redacted", captured.Inputs[2].ValueStatus)
	require.Nil(t, captured.Inputs[2].Value)
	require.Nil(t, captured.Inputs[2].Default)
	require.Equal(t, "embedded-in-bundle", captured.Inputs[3].ValueStatus)
	require.False(t, captured.Inputs[3].ValueAvailable)

	raw, err := json.Marshal(captured)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "vendor-secret")
	require.NotContains(t, string(raw), "customer-secret")
}

func TestCaptureRolesSupportsCloudRoleOutputShapes(t *testing.T) {
	now := time.Now().UTC()
	captured := CaptureRoles(map[string]any{
		"aws": map[string]any{
			"provision_iam_role_arn": "arn:aws:iam::123:role/provision",
			"custom_role_arns": map[string]any{
				"operator": "arn:aws:iam::123:role/operator",
			},
		},
		"gcp": map[string]any{
			"break_glass_sa_emails": map[string]any{
				"support": "support@example.iam.gserviceaccount.com",
			},
		},
	}, now)

	require.Equal(t, now, captured.ObservedAt)
	require.ElementsMatch(t, []CapturedRole{
		{Name: "Provision", Type: "provision", CloudID: "arn:aws:iam::123:role/provision", Provisioned: true},
		{Name: "operator", Type: "custom", CloudID: "arn:aws:iam::123:role/operator", Provisioned: true},
		{Name: "support", Type: "break-glass", CloudID: "support@example.iam.gserviceaccount.com", Provisioned: true},
	}, captured.Roles)
}
