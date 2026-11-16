package command

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		cmd     string
		args    []string
		env     map[string]string
		wantErr bool
	}{
		{
			name: "plain build",
			in:   "go build -o ~/bin/nuon-dev",
			cmd:  "go",
			args: []string{"build", "-o", "~/bin/nuon-dev"},
		},
		{
			name: "quoted ldflags stay one argument",
			in:   `go build -ldflags="-s -w" -o ~/bin/nuon-dev`,
			cmd:  "go",
			args: []string{"build", "-ldflags=-s -w", "-o", "~/bin/nuon-dev"},
		},
		{
			name: "single quotes",
			in:   `go build -ldflags='-s -w' -o out`,
			cmd:  "go",
			args: []string{"build", "-ldflags=-s -w", "-o", "out"},
		},
		{
			name: "leading env assignments",
			in:   "SERVICE_TYPE=api SERVICE_DEPLOYMENT=local /tmp/svc-ctl-api api",
			cmd:  "/tmp/svc-ctl-api",
			args: []string{"api"},
			env:  map[string]string{"SERVICE_TYPE": "api", "SERVICE_DEPLOYMENT": "local"},
		},
		{
			name:    "unterminated quote",
			in:      `go build -ldflags="-s -w`,
			wantErr: true,
		},
		{
			name:    "empty",
			in:      "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, env, err := ParseCommand(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.cmd, cmd)
			require.Equal(t, tt.args, args)
			if tt.env == nil {
				require.Empty(t, env)
			} else {
				require.Equal(t, tt.env, env)
			}
		})
	}
}
