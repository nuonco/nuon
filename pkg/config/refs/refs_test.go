package refs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFieldRefs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []Ref
	}{
		{
			name:  "component output",
			input: `peer = "{{ .nuon.components.database.outputs.endpoint }}"`,
			want: []Ref{{
				Type:  RefTypeComponents,
				Name:  "database",
				Value: "endpoint",
				Input: "nuon.components.database.outputs.endpoint",
			}},
		},
		{
			name:  "legacy install component path",
			input: `{{ .nuon.install.components.nlb.outputs.preview_domain }}`,
			want: []Ref{{
				Type:  RefTypeComponents,
				Name:  "nlb",
				Value: "preview_domain",
				Input: "nuon.install.components.nlb.outputs.preview_domain",
			}},
		},
		{
			name:  "legacy install sandbox path",
			input: `{{ .nuon.install.sandbox.outputs.nuon_dns.public_domain.zone_id }}`,
			want: []Ref{{
				Type:  RefTypeSandbox,
				Name:  "nuon_dns.public_domain.zone_id",
				Input: "nuon.install.sandbox.outputs.nuon_dns.public_domain.zone_id",
			}},
		},
		{
			name:  "action workflow",
			input: `{{ .nuon.actions.workflows.cluster_describe.outputs.cluster }}`,
			want: []Ref{{
				Type:  RefTypeActions,
				Name:  "cluster_describe",
				Input: "nuon.actions.workflows.cluster_describe",
			}},
		},
		{
			name:  "action output",
			input: `{{ .nuon.actions.setup.outputs.token }}`,
			want: []Ref{{
				Type:  RefTypeActions,
				Name:  "setup",
				Value: "token",
				Input: "nuon.actions.setup.outputs.token",
			}},
		},
		{
			name:  "install stack output",
			input: `{{ .nuon.install_stack.outputs.region }}`,
			want: []Ref{{
				Type:  RefTypeInstallStack,
				Name:  "region",
				Input: "nuon.install_stack.outputs.region",
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ParseFieldRefs(tc.input)
			require.Len(t, got, len(tc.want))
			for i := range tc.want {
				assert.Equal(t, tc.want[i].Type, got[i].Type)
				assert.Equal(t, tc.want[i].Name, got[i].Name)
				assert.Equal(t, tc.want[i].Value, got[i].Value)
				assert.Equal(t, tc.want[i].Input, got[i].Input)
			}
		})
	}
}
