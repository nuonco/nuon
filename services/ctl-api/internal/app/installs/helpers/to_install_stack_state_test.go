package helpers

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
