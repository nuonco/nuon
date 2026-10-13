package attribution

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

type fakeClient struct {
	clientVersion string
	agent         string
	userAgent     string
	command       string
}

func (f *fakeClient) SetClientVersion(v string) { f.clientVersion = v }
func (f *fakeClient) SetAgentClient(n string)   { f.agent = n }
func (f *fakeClient) SetUserAgent(u string)     { f.userAgent = u }
func (f *fakeClient) SetCommand(c string)       { f.command = c }

func TestApply(t *testing.T) {
	tests := map[string]struct {
		cfg  *config.Config
		want fakeClient
	}{
		"agent with command": {
			cfg: &config.Config{Agent: "cursor", AgentCommand: "nuon auth login"},
			want: fakeClient{
				clientVersion: version.Version,
				agent:         "cursor",
				userAgent:     "nuon-cli/" + version.Version + " (cursor)",
				command:       "nuon auth login",
			},
		},
		"agent without command": {
			cfg: &config.Config{Agent: "amp"},
			want: fakeClient{
				clientVersion: version.Version,
				agent:         "amp",
				userAgent:     "nuon-cli/" + version.Version + " (amp)",
			},
		},
		"no agent": {
			cfg:  &config.Config{AgentCommand: "nuon auth login"},
			want: fakeClient{clientVersion: version.Version},
		},
		"nil config": {
			want: fakeClient{clientVersion: version.Version},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := &fakeClient{}
			Apply(got, tc.cfg)
			require.Equal(t, tc.want, *got)
		})
	}
}
