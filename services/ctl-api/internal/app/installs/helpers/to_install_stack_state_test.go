package helpers

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestToInstallStackStateTelemetryEndpoint(t *testing.T) {
	for _, endpoint := range []string{"http://internal-acme-telemetry.elb.us-east-1.amazonaws.com:4318", ""} {
		t.Run(endpoint, func(t *testing.T) {
			outputs := app.InstallStackOutputs{
				Data: pgtype.Hstore{"telemetry_endpoint": &endpoint},
			}
			require.NoError(t, outputs.AfterQuery(nil))
			stack := &app.InstallStack{
				InstallStackVersions: []app.InstallStackVersion{{}},
				InstallStackOutputs:  outputs,
			}
			installState := state.State{InstallStack: ToInstallStackState(stack)}
			data, err := installState.AsMap()
			require.NoError(t, err)
			got, err := render.RenderTextV2("{{ .nuon.install_stack.outputs.telemetry_endpoint }}", data)
			require.NoError(t, err)
			assert.Equal(t, endpoint, got)
		})
	}
}

func TestToInstallStackStateNamedPolicyARN(t *testing.T) {
	const policyARN = "arn:aws:iam::123456789012:policy/install-grafana-lgtm-cloudwatch"
	outputs := app.InstallStackOutputs{
		Data: pgtype.Hstore{
			"named_policy_arns": generics.ToPtr(`{"grafana-lgtm-cloudwatch":"` + policyARN + `"}`),
		},
	}
	require.NoError(t, outputs.AfterQuery(nil))
	stack := &app.InstallStack{
		InstallStackVersions: []app.InstallStackVersion{{}},
		InstallStackOutputs:  outputs,
	}
	installState := state.State{InstallStack: ToInstallStackState(stack)}
	data, err := installState.AsMap()
	require.NoError(t, err)

	got, err := render.RenderTextV2(
		`{{ index .nuon.install_stack.outputs.named_policy_arns "grafana-lgtm-cloudwatch" }}`,
		data,
	)

	require.NoError(t, err)
	assert.Equal(t, policyARN, got)
}
