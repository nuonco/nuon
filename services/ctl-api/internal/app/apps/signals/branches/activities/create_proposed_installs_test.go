package activities

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestCreateProposedInstallsInput(t *testing.T) {
	for name, tt := range map[string]struct {
		payload string
		wantErr bool
	}{
		"already scheduled names-only payload": {
			payload: `{"app_id":"app-acme","names":["acme"]}`,
		},
		"full install configs": {
			payload: `{"app_id":"app-acme","app_install_config_sync_id":"sync-acme","install_config_sync_id":"install-sync-acme","installs":[{"name":"acme","config":{"name":"acme"}}]}`,
		},
		"new payload requires sync ids": {
			payload: `{"app_id":"app-acme","installs":[{"name":"acme","config":{"name":"acme"}}]}`,
			wantErr: true,
		},
		"app is required for legacy payloads": {
			payload: `{"names":["acme"]}`,
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var input CreateProposedInstallsInput
			require.NoError(t, json.Unmarshal([]byte(tt.payload), &input))
			err := validator.New().Struct(input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
