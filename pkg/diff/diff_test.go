package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectChanges(t *testing.T) {
	tests := []struct {
		name         string
		original     map[string]interface{}
		modified     map[string]interface{}
		ignoreFields []string
		wantChanges  bool
		wantEntries  int
		wantTypes    []DiffEntryType
	}{
		{
			name:        "no changes",
			original:    map[string]interface{}{"key": "value"},
			modified:    map[string]interface{}{"key": "value"},
			wantChanges: false,
			wantEntries: 0,
		},
		{
			name:        "added value",
			original:    map[string]interface{}{},
			modified:    map[string]interface{}{"key": "value"},
			wantChanges: true,
			wantEntries: 1,
			wantTypes:   []DiffEntryType{EntryAdded},
		},
		{
			name:        "removed value",
			original:    map[string]interface{}{"key": "value"},
			modified:    map[string]interface{}{},
			wantChanges: true,
			wantEntries: 1,
			wantTypes:   []DiffEntryType{EntryRemoved},
		},
		{
			name:        "modified value",
			original:    map[string]interface{}{"key": "value1"},
			modified:    map[string]interface{}{"key": "value2"},
			wantChanges: true,
			wantEntries: 1,
			wantTypes:   []DiffEntryType{EntryModified},
		},
		{
			name:        "nested changes",
			original:    map[string]interface{}{"parent": map[string]interface{}{"child": "old"}},
			modified:    map[string]interface{}{"parent": map[string]interface{}{"child": "new"}},
			wantChanges: true,
			wantEntries: 1,
			wantTypes:   []DiffEntryType{EntryModified},
		},
		{
			name: "multiple changes",
			original: map[string]interface{}{
				"unchanged": "same",
				"modified":  "old",
				"removed":   "will be gone",
				"nested":    map[string]interface{}{"a": "1", "b": "2"},
			},
			modified: map[string]interface{}{
				"unchanged": "same",
				"modified":  "new",
				"added":     "new field",
				"nested":    map[string]interface{}{"a": "1", "b": "changed", "c": "added"},
			},
			wantChanges: true,
			wantEntries: 5, // modified, removed, added, 2 nested changes
		},
		{
			name: "with ignored fields",
			original: map[string]interface{}{
				"important": "value1",
				"status":    "ignore me",
				"metadata": map[string]interface{}{
					"name":              "test",
					"creationTimestamp": "2023-01-01T00:00:00Z",
				},
			},
			modified: map[string]interface{}{
				"important": "value2",
				"status":    "new status",
				"metadata": map[string]interface{}{
					"name":              "test",
					"creationTimestamp": "2023-01-02T00:00:00Z",
				},
			},
			ignoreFields: []string{"status", "metadata.creationTimestamp"},
			wantChanges:  true,
			wantEntries:  1, // Only important field should be detected as changed
		},
		{
			name: "kubernetes configmap example",
			original: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "demo",
					"namespace": "default",
				},
				"data": map[string]interface{}{
					"sample_data": "3",
				},
			},
			modified: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "demo",
					"namespace": "default",
				},
				"data": map[string]interface{}{
					"sample_data": "4",
				},
			},
			ignoreFields: []string{
				"metadata.creationTimestamp",
				"metadata.resourceVersion",
				"metadata.uid",
				"metadata.managedFields",
				"status",
			},
			wantChanges: true,
			wantEntries: 1,
		},
		{
			name: "array changes",
			original: map[string]interface{}{
				"items": []interface{}{
					"item1",
					"item2",
				},
			},
			modified: map[string]interface{}{
				"items": []interface{}{
					"item1",
					"item3", // changed
					"item4", // added
				},
			},
			wantChanges: true,
			wantEntries: 2, // Array changes should be detected
		},
		{
			name: "nested array object changes",
			original: map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "container1",
							"image": "image:v1",
						},
					},
				},
			},
			modified: map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "container1",
							"image": "image:v2", // changed
						},
					},
				},
			},
			wantChanges: true,
			wantEntries: 1,
		},
		{
			name: "complex kubernetes example",
			original: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginx",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
				},
				"spec": map[string]interface{}{
					"replicas": float64(3),
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "nginx",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "nginx",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.14.2",
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": float64(80),
										},
									},
								},
							},
						},
					},
				},
			},
			modified: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginx",
					"labels": map[string]interface{}{
						"app":         "nginx",
						"environment": "production", // added
					},
				},
				"spec": map[string]interface{}{
					"replicas": float64(5), // changed
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "nginx",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app":         "nginx",
								"environment": "production", // added
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.15.0", // changed
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": float64(80),
										},
									},
								},
							},
						},
					},
				},
			},
			ignoreFields: []string{
				"metadata.creationTimestamp",
				"metadata.resourceVersion",
				"metadata.uid",
				"metadata.managedFields",
				"status",
			},
			wantChanges: true,
			wantEntries: 4, // Complex changes should be detected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, hasChanges := DetectChanges(tt.original, tt.modified, tt.ignoreFields)

			assert.Equal(t, tt.wantChanges, hasChanges, "hasChanges mismatch")

			if !hasChanges {
				assert.Empty(t, entries, "entries should be empty when no changes")
				return
			}

			// Check if expected number of entries is correct
			if tt.wantEntries > 0 {
				assert.Len(t, entries, tt.wantEntries, "incorrect number of entries")
			}

			// Check entry types if specified
			if len(tt.wantTypes) > 0 {
				for i, wantType := range tt.wantTypes {
					if i < len(entries) {
						assert.Equal(t, wantType, entries[i].Type, "entry type mismatch at index %d", i)
					}
				}
			}
		})
	}
}

func TestParseRawResourceName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantNamespace string
		wantName      string
		wantKind      string
		wantApiPath   string
	}{
		{
			name:          "valid format",
			input:         "default, nginx-deployment, Deployment (apps/v1)",
			wantNamespace: "default",
			wantName:      "nginx-deployment",
			wantKind:      "Deployment",
			wantApiPath:   "apps/v1",
		},
		{
			name:          "with spaces",
			input:         "  kube-system,  coredns,  Deployment  (  apps/v1  )",
			wantNamespace: "kube-system",
			wantName:      "coredns",
			wantKind:      "Deployment",
			wantApiPath:   "apps/v1",
		},
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "invalid format",
			input: "this is not a valid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespace, name, kind, apiPath := ParseRawResourceName(tt.input)
			assert.Equal(t, tt.wantNamespace, namespace)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantKind, kind)
			assert.Equal(t, tt.wantApiPath, apiPath)
		})
	}
}

func TestCompressAndEncodeDecode(t *testing.T) {
	// Test data
	type TestPlan struct {
		Name   string                 `json:"name"`
		Action string                 `json:"action"`
		Data   map[string]interface{} `json:"data"`
	}

	originalPlan := TestPlan{
		Name:   "test-plan",
		Action: "apply",
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 42.0,
			"nested": map[string]interface{}{
				"innerKey": "innerValue",
			},
		},
	}

	// Compress and encode
	encoded, err := CompressAndEncodeObject(originalPlan)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode and decompress
	var decodedPlan TestPlan
	err = DecompressAndDecodePlan(encoded, &decodedPlan)
	assert.NoError(t, err)

	// Verify the decoded plan matches the original
	assert.Equal(t, originalPlan.Name, decodedPlan.Name)
	assert.Equal(t, originalPlan.Action, decodedPlan.Action)
	assert.Equal(t, originalPlan.Data["key1"], decodedPlan.Data["key1"])
	assert.Equal(t, originalPlan.Data["key2"], decodedPlan.Data["key2"])

	// Check nested data
	nestedOriginal := originalPlan.Data["nested"].(map[string]interface{})
	nestedDecoded := decodedPlan.Data["nested"].(map[string]interface{})
	assert.Equal(t, nestedOriginal["innerKey"], nestedDecoded["innerKey"])
}
