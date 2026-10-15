package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponentTeardownScanValue(t *testing.T) {
	tests := map[string]struct {
		in   *ComponentTeardown
		want *ComponentTeardown
	}{
		"nil": {},
		"retain list": {
			in:   &ComponentTeardown{Retain: []string{"aws_rds_cluster.main", "module.buckets"}},
			want: &ComponentTeardown{Retain: []string{"aws_rds_cluster.main", "module.buckets"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := tc.in.Value()
			require.NoError(t, err)

			var out ComponentTeardown
			if val == nil {
				require.NoError(t, out.Scan(nil))
				require.Nil(t, tc.want)
				require.Empty(t, out.Retain)
				return
			}
			require.NoError(t, out.Scan(val))
			require.Equal(t, tc.want, &out)
		})
	}
}

func TestComponentConfigConnectionGetTeardownRetain(t *testing.T) {
	tests := map[string]struct {
		ccc  ComponentConfigConnection
		want []string
	}{
		"no teardown": {},
		"retain list": {
			ccc:  ComponentConfigConnection{Teardown: &ComponentTeardown{Retain: []string{"module.buckets"}}},
			want: []string{"module.buckets"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, tc.ccc.GetTeardownRetain())
		})
	}
}
