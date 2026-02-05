package plandiff

import (
	"strings"
	"testing"
)

func TestParseTerraformPlan(t *testing.T) {
	tests := []struct {
		name           string
		plan           *TerraformPlan
		wantResCreate  int
		wantResUpdate  int
		wantResDelete  int
		wantResReplace int
		wantResRead    int
		wantOutCreate  int
		wantDriftCount int
	}{
		{
			name: "empty plan",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
			},
			wantResCreate: 0,
		},
		{
			name: "single create",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_instance.example",
						Type:    "aws_instance",
						Name:    "example",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionCreate},
						},
					},
				},
			},
			wantResCreate: 1,
		},
		{
			name: "multiple actions",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_instance.one",
						Type:    "aws_instance",
						Name:    "one",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionCreate},
						},
					},
					{
						Address: "aws_instance.two",
						Type:    "aws_instance",
						Name:    "two",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionUpdate},
						},
					},
					{
						Address: "aws_instance.three",
						Type:    "aws_instance",
						Name:    "three",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionDelete},
						},
					},
				},
			},
			wantResCreate: 1,
			wantResUpdate: 1,
			wantResDelete: 1,
		},
		{
			name: "replace action counts as delete and create",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_instance.example",
						Type:    "aws_instance",
						Name:    "example",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionReplace},
						},
					},
				},
			},
			wantResReplace: 1,
			wantResCreate:  1,
			wantResDelete:  1,
		},
		{
			name: "read action",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "data.aws_ami.example",
						Type:    "aws_ami",
						Name:    "example",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionRead},
						},
					},
				},
			},
			wantResRead: 1,
		},
		{
			name: "output changes",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
				OutputChanges: map[string]TerraformOutputChangeRaw{
					"instance_ip": {
						Actions: []TerraformChangeAction{TerraformActionCreate},
						After:   "10.0.0.1",
					},
				},
			},
			wantOutCreate: 1,
		},
		{
			name: "resource drift",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
				ResourceDrift: []TerraformResourceDrift{
					{
						Address: "aws_instance.drifted",
						Type:    "aws_instance",
						Name:    "drifted",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionUpdate},
						},
					},
				},
			},
			wantDriftCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTerraformPlan(tt.plan)

			if parsed.Resources.Summary.Create != tt.wantResCreate {
				t.Errorf("Resources.Summary.Create = %d, want %d", parsed.Resources.Summary.Create, tt.wantResCreate)
			}
			if parsed.Resources.Summary.Update != tt.wantResUpdate {
				t.Errorf("Resources.Summary.Update = %d, want %d", parsed.Resources.Summary.Update, tt.wantResUpdate)
			}
			if parsed.Resources.Summary.Delete != tt.wantResDelete {
				t.Errorf("Resources.Summary.Delete = %d, want %d", parsed.Resources.Summary.Delete, tt.wantResDelete)
			}
			if parsed.Resources.Summary.Replace != tt.wantResReplace {
				t.Errorf("Resources.Summary.Replace = %d, want %d", parsed.Resources.Summary.Replace, tt.wantResReplace)
			}
			if parsed.Resources.Summary.Read != tt.wantResRead {
				t.Errorf("Resources.Summary.Read = %d, want %d", parsed.Resources.Summary.Read, tt.wantResRead)
			}
			if parsed.Outputs.Summary.Create != tt.wantOutCreate {
				t.Errorf("Outputs.Summary.Create = %d, want %d", parsed.Outputs.Summary.Create, tt.wantOutCreate)
			}
			if len(parsed.Drift.Changes) != tt.wantDriftCount {
				t.Errorf("len(Drift.Changes) = %d, want %d", len(parsed.Drift.Changes), tt.wantDriftCount)
			}
		})
	}
}

func TestMergeAfterUnknown(t *testing.T) {
	tests := []struct {
		name         string
		after        any
		afterUnknown any
		wantKey      string
		wantValue    string
	}{
		{
			name:         "nil afterUnknown",
			after:        map[string]any{"key": "value"},
			afterUnknown: nil,
			wantKey:      "key",
			wantValue:    "value",
		},
		{
			name:         "unknown field marked true",
			after:        map[string]any{"known": "value"},
			afterUnknown: map[string]any{"unknown": true},
			wantKey:      "unknown",
			wantValue:    "(known after apply)",
		},
		{
			name:         "unknown field marked false",
			after:        map[string]any{"key": "value"},
			afterUnknown: map[string]any{"key": false},
			wantKey:      "key",
			wantValue:    "value",
		},
		{
			name:         "nil after creates map",
			after:        nil,
			afterUnknown: map[string]any{"new_field": true},
			wantKey:      "new_field",
			wantValue:    "(known after apply)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeAfterUnknown(tt.after, tt.afterUnknown)

			if result == nil {
				if tt.after != nil {
					t.Error("mergeAfterUnknown returned nil for non-nil after")
				}
				return
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Errorf("mergeAfterUnknown result is not a map: %T", result)
				return
			}

			if val, exists := resultMap[tt.wantKey]; exists {
				if valStr, ok := val.(string); ok && valStr != tt.wantValue {
					t.Errorf("resultMap[%s] = %v, want %v", tt.wantKey, valStr, tt.wantValue)
				}
			}
		})
	}
}

func TestFormatTerraformPlan(t *testing.T) {
	tests := []struct {
		name     string
		plan     *TerraformPlan
		contains []string
	}{
		{
			name: "no changes",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
			},
			contains: []string{"No changes"},
		},
		{
			name: "create action",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_instance.example",
						Type:    "aws_instance",
						Name:    "example",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionCreate},
						},
					},
				},
			},
			contains: []string{"1 to create", "aws_instance", "aws_instance.example"},
		},
		{
			name: "with module address",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address:       "module.vpc.aws_vpc.main",
						ModuleAddress: strPtr("module.vpc"),
						Type:          "aws_vpc",
						Name:          "main",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionCreate},
						},
					},
				},
			},
			contains: []string{"module.vpc"},
		},
		{
			name: "drift section",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
				ResourceDrift: []TerraformResourceDrift{
					{
						Address: "aws_instance.drifted",
						Type:    "aws_instance",
						Name:    "drifted",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionUpdate},
						},
					},
				},
			},
			contains: []string{"Resource Drift", "aws_instance.drifted"},
		},
		{
			name: "update with field-level diff",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_instance.web",
						Type:    "aws_instance",
						Name:    "web",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionUpdate},
							Before: map[string]any{
								"ami":           "ami-old",
								"instance_type": "t2.micro",
								"tags":          map[string]any{"Name": "web"},
							},
							After: map[string]any{
								"ami":           "ami-new",
								"instance_type": "t2.small",
								"tags":          map[string]any{"Name": "web"},
							},
						},
					},
				},
			},
			contains: []string{
				"1 to update",
				"aws_instance.web",
				"~ ami",
				`"ami-old"`,
				`"ami-new"`,
				"~ instance_type",
				`"t2.micro"`,
				`"t2.small"`,
			},
		},
		{
			name: "create with new fields",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_s3_bucket.logs",
						Type:    "aws_s3_bucket",
						Name:    "logs",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionCreate},
							Before:  nil,
							After: map[string]any{
								"bucket": "my-logs-bucket",
								"acl":    "private",
								"id":     "(known after apply)",
							},
						},
					},
				},
			},
			contains: []string{
				"1 to create",
				"+ bucket",
				`"my-logs-bucket"`,
				"+ acl",
				`"private"`,
				"+ id",
				"(known after apply)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTerraformPlan(tt.plan)
			output := FormatTerraformPlan(parsed)

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("FormatTerraformPlan output missing %q", want)
				}
			}
		})
	}
}

func TestHasTerraformChanges(t *testing.T) {
	tests := []struct {
		name       string
		plan       *TerraformPlan
		wantChange bool
	}{
		{
			name: "no changes",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
			},
			wantChange: false,
		},
		{
			name: "only no-op changes",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_instance.example",
						Type:    "aws_instance",
						Name:    "example",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionNoOp},
						},
					},
				},
			},
			wantChange: false,
		},
		{
			name: "has create",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{
					{
						Address: "aws_instance.example",
						Type:    "aws_instance",
						Name:    "example",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionCreate},
						},
					},
				},
			},
			wantChange: true,
		},
		{
			name: "has output change",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
				OutputChanges: map[string]TerraformOutputChangeRaw{
					"output1": {Actions: []TerraformChangeAction{TerraformActionUpdate}},
				},
			},
			wantChange: true,
		},
		{
			name: "has drift",
			plan: &TerraformPlan{
				ResourceChanges: []TerraformResourceChange{},
				ResourceDrift: []TerraformResourceDrift{
					{
						Address: "aws_instance.drifted",
						Type:    "aws_instance",
						Name:    "drifted",
						Change: TerraformResourceChangeData{
							Actions: []TerraformChangeAction{TerraformActionUpdate},
						},
					},
				},
			},
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseTerraformPlan(tt.plan)
			got := HasTerraformChanges(parsed)

			if got != tt.wantChange {
				t.Errorf("HasTerraformChanges() = %v, want %v", got, tt.wantChange)
			}
		})
	}
}

func TestIncrementSummary(t *testing.T) {
	tests := []struct {
		action  TerraformChangeAction
		checkFn func(s *Summary) int
		wantVal int
	}{
		{TerraformActionCreate, func(s *Summary) int { return s.Create }, 1},
		{TerraformActionUpdate, func(s *Summary) int { return s.Update }, 1},
		{TerraformActionDelete, func(s *Summary) int { return s.Delete }, 1},
		{TerraformActionReplace, func(s *Summary) int { return s.Replace }, 1},
		{TerraformActionRead, func(s *Summary) int { return s.Read }, 1},
		{TerraformActionNoOp, func(s *Summary) int { return s.NoOp }, 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			s := &Summary{}
			incrementSummary(s, tt.action)

			if got := tt.checkFn(s); got != tt.wantVal {
				t.Errorf("incrementSummary(%s) counter = %d, want %d", tt.action, got, tt.wantVal)
			}
		})
	}
}

func TestIsSensitive(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"true bool", true, true},
		{"false bool", false, false},
		{"string", "sensitive", false},
		{"number", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSensitive(tt.value); got != tt.want {
				t.Errorf("isSensitive(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestFormatTerraformFieldDiff(t *testing.T) {
	tests := []struct {
		name     string
		before   any
		after    any
		action   TerraformChangeAction
		contains []string
		excludes []string
	}{
		{
			name:   "field added",
			before: map[string]any{},
			after:  map[string]any{"instance_type": "t2.micro"},
			action: TerraformActionCreate,
			contains: []string{
				"+ instance_type",
				`"t2.micro"`,
			},
		},
		{
			name:   "field removed",
			before: map[string]any{"old_field": "value"},
			after:  map[string]any{},
			action: TerraformActionDelete,
			contains: []string{
				"- old_field",
			},
		},
		{
			name:   "field changed",
			before: map[string]any{"ami": "ami-123"},
			after:  map[string]any{"ami": "ami-456"},
			action: TerraformActionUpdate,
			contains: []string{
				"~ ami",
				`"ami-123"`,
				`"ami-456"`,
			},
		},
		{
			name: "multiple fields with mixed changes",
			before: map[string]any{
				"name":          "old-name",
				"instance_type": "t2.micro",
				"removed_field": "gone",
			},
			after: map[string]any{
				"name":          "new-name",
				"instance_type": "t2.micro", // unchanged
				"new_field":     "added",
			},
			action: TerraformActionUpdate,
			contains: []string{
				"~ name",
				"+ new_field",
				"- removed_field",
			},
			excludes: []string{
				"instance_type", // unchanged fields should not appear
			},
		},
		{
			name:   "nested map expands fields",
			before: nil,
			after: map[string]any{
				"tags": map[string]any{"Name": "test", "Env": "dev"},
			},
			action: TerraformActionCreate,
			contains: []string{
				"+ tags",
				"Name = \"test\"",
				"Env = \"dev\"",
			},
		},
		{
			name:   "array expands items",
			before: nil,
			after: map[string]any{
				"security_groups": []any{"sg-123", "sg-456"},
			},
			action: TerraformActionCreate,
			contains: []string{
				"+ security_groups",
				"\"sg-123\"",
				"\"sg-456\"",
			},
		},
		{
			name:   "boolean values",
			before: map[string]any{"enabled": false},
			after:  map[string]any{"enabled": true},
			action: TerraformActionUpdate,
			contains: []string{
				"~ enabled",
				"false",
				"true",
			},
		},
		{
			name:   "numeric values",
			before: map[string]any{"count": float64(1)},
			after:  map[string]any{"count": float64(5)},
			action: TerraformActionUpdate,
			contains: []string{
				"~ count",
				"1",
				"5",
			},
		},
		{
			name:   "known after apply",
			before: nil,
			after:  map[string]any{"id": "(known after apply)"},
			action: TerraformActionCreate,
			contains: []string{
				"+ id",
				"(known after apply)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatTerraformFieldDiff(tt.before, tt.after, tt.action)

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("formatTerraformFieldDiff output missing %q\nOutput:\n%s", want, output)
				}
			}

			for _, notWant := range tt.excludes {
				if strings.Contains(output, notWant) {
					t.Errorf("formatTerraformFieldDiff output should not contain %q\nOutput:\n%s", notWant, output)
				}
			}
		})
	}
}
