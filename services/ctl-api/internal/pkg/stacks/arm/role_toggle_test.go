package arm

import "testing"

func TestAzureRoleEnableParameters_DefaultsFromAppliedStack(t *testing.T) {
	provision := azureOperationIdentity{kind: "provision", suffix: "provision", enabledByDefault: true}
	custom := azureOperationIdentity{kind: "custom", suffix: "custom-dns", roleName: "inl1-dns", enabledByDefault: true}
	added := azureOperationIdentity{kind: "custom", suffix: "custom-new", roleName: "inl1-new", enabledByDefault: true}
	breakGlass := azureOperationIdentity{kind: "breakglass", suffix: "bg-admin", roleName: "inl1-admin", enabledByDefault: false}
	ids := []azureOperationIdentity{provision, custom, added, breakGlass}

	cases := map[string]struct {
		outputs map[string]any
		want    map[string]bool
	}{
		"first apply uses enabled_in_stack": {
			outputs: nil,
			want:    map[string]bool{"provision": true, "custom-dns": true, "custom-new": true, "bg-admin": false},
		},
		"re-apply keeps what the customer chose": {
			outputs: map[string]any{
				"provision_identity_client_id":    "",
				"custom_identity_client_ids":      map[string]any{"inl1-dns": ""},
				"break_glass_identity_client_ids": `{"inl1-admin":"9f0c"}`,
			},
			want: map[string]bool{"provision": false, "custom-dns": false, "custom-new": true, "bg-admin": true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			params := azureRoleEnableParameters(ids, tc.outputs)
			for _, id := range ids {
				got := params[azureRoleEnableParamName(id)].DefaultValue
				if got != tc.want[id.suffix] {
					t.Errorf("%s: default %v, want %v", id.suffix, got, tc.want[id.suffix])
				}
			}
		})
	}
}
