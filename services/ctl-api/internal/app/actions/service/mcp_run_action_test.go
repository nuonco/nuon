package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRunEnvVarKeys(t *testing.T) {
	tests := map[string]struct {
		envVars map[string]string
		wantErr string
	}{
		"nil map": {
			envVars: nil,
		},
		"valid keys": {
			envVars: map[string]string{"TEST_VAR": "value", "FOO": "bar"},
		},
		"empty key": {
			envVars: map[string]string{"": "value"},
			wantErr: "run_env_vars keys cannot be empty",
		},
		"key with space": {
			envVars: map[string]string{"MY VAR": "value"},
			wantErr: `run_env_vars key "MY VAR" cannot contain '=' or whitespace`,
		},
		"key with equals": {
			envVars: map[string]string{"FOO=BAR": "value"},
			wantErr: `run_env_vars key "FOO=BAR" cannot contain '=' or whitespace`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateRunEnvVarKeys(tc.envVars)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tc.wantErr)
		})
	}
}
