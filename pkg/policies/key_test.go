package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestKeyFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
		wantErr  string
	}{
		{
			name: "cluster_role",
			contents: `kind: ClusterRole
metadata:
  name: kyverno:linkerd:manage
`,
			want: "clusterrole-kyverno-linkerd-manage.yaml",
		},
		{
			name: "cluster_policy",
			contents: `kind: ClusterPolicy
metadata:
  name: linkerd-authz-for-restate
`,
			want: "clusterpolicy-linkerd-authz-for-restate.yaml",
		},
		{
			name: "namespaced_includes_namespace",
			contents: `kind: Role
metadata:
  name: view
  namespace: restate
`,
			want: "role-restate-view.yaml",
		},
		{
			name: "trims_leading_trailing_dashes",
			contents: `kind: ClusterRole
metadata:
  name: ":weird:"
`,
			want: "clusterrole-weird.yaml",
		},
		{
			name: "missing_kind",
			contents: `metadata:
  name: x
`,
			wantErr: "missing 'kind'",
		},
		{
			name: "missing_metadata",
			contents: `kind: ClusterRole
`,
			wantErr: "missing 'metadata'",
		},
		{
			name: "missing_name",
			contents: `kind: ClusterRole
metadata:
  labels:
    foo: bar
`,
			wantErr: "missing 'metadata.name'",
		},
		{
			name:     "invalid_yaml",
			contents: "kind: ClusterRole\n  bad-indent: oops\n broken",
			wantErr:  "parse manifest yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ManifestKeyFromYAML(tc.contents)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsLegacyKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"0.yaml", true},
		{"1.yaml", true},
		{"12.yaml", true},
		{"clusterrole-x.yaml", false},
		{"x.yaml", false},
		{"0.yml", false},
		{"0", false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, IsLegacyKey(tc.key), "key=%q", tc.key)
	}
}
