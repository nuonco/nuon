package app

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
)

func TestInstallStackOutputsRunnerDisabled(t *testing.T) {
	for name, tc := range map[string]struct {
		data pgtype.Hstore
		want bool
	}{
		"runner_enabled false": {pgtype.Hstore{"runner_enabled": generics.ToPtr("false")}, true},
		"runner_enabled true":  {pgtype.Hstore{"runner_enabled": generics.ToPtr("true")}, false},
		"key absent":           {pgtype.Hstore{"account_id": generics.ToPtr("123")}, false},
		"key present but null": {pgtype.Hstore{"runner_enabled": nil}, false},
	} {
		t.Run(name, func(t *testing.T) {
			for _, platform := range []string{"aws", "gcp"} {
				data := pgtype.Hstore{}
				for k, v := range tc.data {
					data[k] = v
				}
				if platform == "gcp" {
					data["runner_service_account_email"] = generics.ToPtr("runner@example.com")
				}

				outputs := InstallStackOutputs{Data: data}
				require.NoError(t, outputs.AfterQuery(nil))
				require.Equal(t, tc.want, outputs.RunnerDisabled(), platform)
			}
		})
	}
}
