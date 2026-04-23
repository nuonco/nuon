package terraform

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindLegacyPolicyKeyMigrations(t *testing.T) {
	t.Run("nil_state", func(t *testing.T) {
		mvs, err := findLegacyPolicyKeyMigrations(nil)
		require.NoError(t, err)
		assert.Empty(t, mvs)
	})

	t.Run("no_vendor_policies", func(t *testing.T) {
		state := &tfjson.State{
			Values: &tfjson.StateValues{
				RootModule: &tfjson.StateModule{
					Resources: []*tfjson.StateResource{
						{
							Address: "aws_s3_bucket.foo",
							Type:    "aws_s3_bucket",
							Name:    "foo",
						},
					},
				},
			},
		}
		mvs, err := findLegacyPolicyKeyMigrations(state)
		require.NoError(t, err)
		assert.Empty(t, mvs)
	})

	t.Run("all_legacy_keys_migrated", func(t *testing.T) {
		state := &tfjson.State{
			Values: &tfjson.StateValues{
				RootModule: &tfjson.StateModule{
					Resources: []*tfjson.StateResource{
						{
							Address: `kubectl_manifest.vendor_policies["0.yaml"]`,
							Type:    "kubectl_manifest",
							Name:    "vendor_policies",
							Index:   "0.yaml",
							AttributeValues: map[string]any{
								"yaml_body": "kind: ClusterRole\nmetadata:\n  name: kyverno:linkerd:manage\n",
							},
						},
						{
							Address: `kubectl_manifest.vendor_policies["1.yaml"]`,
							Type:    "kubectl_manifest",
							Name:    "vendor_policies",
							Index:   "1.yaml",
							AttributeValues: map[string]any{
								"yaml_body": "kind: ClusterPolicy\nmetadata:\n  name: linkerd-authz\n",
							},
						},
					},
				},
			},
		}

		mvs, err := findLegacyPolicyKeyMigrations(state)
		require.NoError(t, err)
		require.Len(t, mvs, 2)

		assert.Equal(t, "0.yaml", mvs[0].oldKey)
		assert.Equal(t, "clusterrole-kyverno-linkerd-manage.yaml", mvs[0].newKey)
		assert.Equal(t, `kubectl_manifest.vendor_policies["0.yaml"]`, mvs[0].sourceAddress)
		assert.Equal(t, `kubectl_manifest.vendor_policies["clusterrole-kyverno-linkerd-manage.yaml"]`, mvs[0].destinationAddress)
		assert.Equal(t, "ClusterRole", mvs[0].kind)
		assert.Equal(t, "kyverno:linkerd:manage", mvs[0].name)

		assert.Equal(t, "1.yaml", mvs[1].oldKey)
		assert.Equal(t, "clusterpolicy-linkerd-authz.yaml", mvs[1].newKey)
	})

	t.Run("already_migrated_keys_skipped", func(t *testing.T) {
		state := &tfjson.State{
			Values: &tfjson.StateValues{
				RootModule: &tfjson.StateModule{
					Resources: []*tfjson.StateResource{
						{
							Address: `kubectl_manifest.vendor_policies["clusterrole-foo.yaml"]`,
							Type:    "kubectl_manifest",
							Name:    "vendor_policies",
							Index:   "clusterrole-foo.yaml",
							AttributeValues: map[string]any{
								"yaml_body": "kind: ClusterRole\nmetadata:\n  name: foo\n",
							},
						},
					},
				},
			},
		}
		mvs, err := findLegacyPolicyKeyMigrations(state)
		require.NoError(t, err)
		assert.Empty(t, mvs)
	})

	t.Run("mixed_legacy_and_migrated", func(t *testing.T) {
		state := &tfjson.State{
			Values: &tfjson.StateValues{
				RootModule: &tfjson.StateModule{
					Resources: []*tfjson.StateResource{
						{
							Address: `kubectl_manifest.vendor_policies["0.yaml"]`,
							Type:    "kubectl_manifest",
							Name:    "vendor_policies",
							Index:   "0.yaml",
							AttributeValues: map[string]any{
								"yaml_body": "kind: ClusterRole\nmetadata:\n  name: a\n",
							},
						},
						{
							Address: `kubectl_manifest.vendor_policies["clusterpolicy-b.yaml"]`,
							Type:    "kubectl_manifest",
							Name:    "vendor_policies",
							Index:   "clusterpolicy-b.yaml",
							AttributeValues: map[string]any{
								"yaml_body": "kind: ClusterPolicy\nmetadata:\n  name: b\n",
							},
						},
					},
				},
			},
		}
		mvs, err := findLegacyPolicyKeyMigrations(state)
		require.NoError(t, err)
		require.Len(t, mvs, 1)
		assert.Equal(t, "0.yaml", mvs[0].oldKey)
	})

	t.Run("walks_child_modules", func(t *testing.T) {
		state := &tfjson.State{
			Values: &tfjson.StateValues{
				RootModule: &tfjson.StateModule{
					ChildModules: []*tfjson.StateModule{
						{
							Address: "module.sandbox",
							Resources: []*tfjson.StateResource{
								{
									Address: `module.sandbox.kubectl_manifest.vendor_policies["2.yaml"]`,
									Type:    "kubectl_manifest",
									Name:    "vendor_policies",
									Index:   "2.yaml",
									AttributeValues: map[string]any{
										"yaml_body": "kind: ClusterRole\nmetadata:\n  name: nested\n",
									},
								},
							},
						},
					},
				},
			},
		}
		mvs, err := findLegacyPolicyKeyMigrations(state)
		require.NoError(t, err)
		require.Len(t, mvs, 1)
		assert.Equal(t, `module.sandbox.kubectl_manifest.vendor_policies["2.yaml"]`, mvs[0].sourceAddress)
		assert.Equal(t, `module.sandbox.kubectl_manifest.vendor_policies["clusterrole-nested.yaml"]`, mvs[0].destinationAddress)
	})

	t.Run("legacy_key_missing_yaml_body_errors", func(t *testing.T) {
		state := &tfjson.State{
			Values: &tfjson.StateValues{
				RootModule: &tfjson.StateModule{
					Resources: []*tfjson.StateResource{
						{
							Address:         `kubectl_manifest.vendor_policies["0.yaml"]`,
							Type:            "kubectl_manifest",
							Name:            "vendor_policies",
							Index:           "0.yaml",
							AttributeValues: map[string]any{},
						},
					},
				},
			},
		}
		_, err := findLegacyPolicyKeyMigrations(state)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no yaml_body attribute")
	})
}

func TestReplaceIndexInAddress(t *testing.T) {
	tests := []struct {
		address string
		oldKey  string
		newKey  string
		want    string
	}{
		{
			address: `kubectl_manifest.vendor_policies["0.yaml"]`,
			oldKey:  "0.yaml",
			newKey:  "clusterrole-x.yaml",
			want:    `kubectl_manifest.vendor_policies["clusterrole-x.yaml"]`,
		},
		{
			address: `module.sandbox.kubectl_manifest.vendor_policies["12.yaml"]`,
			oldKey:  "12.yaml",
			newKey:  "clusterpolicy-y.yaml",
			want:    `module.sandbox.kubectl_manifest.vendor_policies["clusterpolicy-y.yaml"]`,
		},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, replaceIndexInAddress(tc.address, tc.oldKey, tc.newKey))
	}
}
