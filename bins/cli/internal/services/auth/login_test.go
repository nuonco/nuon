package auth

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
)

func TestSelectAPIURLAgent(t *testing.T) {
	tests := map[string]struct {
		configured string
		want       string
	}{
		"reuses configured url": {
			configured: "https://api.example.com",
			want:       "https://api.example.com",
		},
		"defaults to nuon cloud": {
			want: "https://api.nuon.co",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &config.Config{Viper: viper.New(), Agent: "cursor"}
			if tc.configured != "" {
				cfg.Set("api_url", tc.configured)
			}
			svc := &Service{cfg: cfg}

			got, err := svc.selectAPIURL()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
