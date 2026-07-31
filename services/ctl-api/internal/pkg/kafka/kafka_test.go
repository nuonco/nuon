package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestClientID(t *testing.T) {
	tests := []struct {
		name string
		cfg  internal.Config
		want string
	}{
		{
			name: "api deployment",
			cfg: internal.Config{
				ServiceName:       "ctl-api",
				ServiceType:       "api",
				ServiceDeployment: "runner",
			},
			want: "ctl-api/api-runner",
		},
		{
			name: "worker names its temporal namespace",
			cfg: internal.Config{
				ServiceName:       "ctl-api",
				ServiceType:       "worker",
				ServiceDeployment: "installs",
			},
			want: "ctl-api/worker-installs",
		},
		{
			name: "consumer names its sink, matching the group",
			cfg: internal.Config{
				ServiceName:       "ctl-api",
				ServiceType:       "consumer",
				ServiceDeployment: "clickhouse-sink",
			},
			want: "ctl-api/consumer-clickhouse-sink",
		},
		{
			name: "no deployment set",
			cfg: internal.Config{
				ServiceName: "ctl-api",
				ServiceType: "startup",
			},
			want: "ctl-api/startup",
		},
		{
			name: "explicit override wins",
			cfg: internal.Config{
				KafkaClientID:     "something-else",
				ServiceName:       "ctl-api",
				ServiceType:       "api",
				ServiceDeployment: "public",
			},
			want: "something-else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, ClientID(&tt.cfg))
		})
	}
}
