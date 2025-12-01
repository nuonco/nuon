package approvalplan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesApprovalPlan_IsNoop(t *testing.T) { //nolint:funlen
	tests := []struct {
		name     string
		planJSON string
		want     bool
		wantErr  bool
	}{
		{
			name:     "empty plan json",
			planJSON: `{}`,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "plan with empty content_diff",
			planJSON: `{"k8s_content_diff": []}`,
			want:     true,
			wantErr:  false,
		},
		{
			name: "plan with unchanged entries (type 0)",
			planJSON: `{
				"k8s_content_diff": [
					{
						"type": 0,
						"name": "deployment",
						"kind": "Deployment",
						"entries": []
					},
					{
						"type": 0,
						"name": "service",
						"kind": "Service",
						"entries": []
					}
				]
			}`,
			want:    true,
			wantErr: false,
		},
		{
			name: "plan with mixed entries (unchanged and changed)",
			planJSON: `{
				"k8s_content_diff": [
					{
						"type": 0,
						"name": "deployment",
						"kind": "Deployment"
					},
					{
						"type": 3,
						"name": "service",
						"kind": "Service"
					}
				]
			}`,
			want:    false,
			wantErr: false,
		},
		{
			name: "plan with added entries (type 2)",
			planJSON: `{
				"k8s_content_diff": [
					{
						"type": 2,
						"name": "deployment",
						"kind": "Deployment"
					}
				]
			}`,
			want:    false,
			wantErr: false,
		},
		{
			name: "plan with removed entries (type 1)",
			planJSON: `{
				"k8s_content_diff": [
					{
						"type": 1,
						"name": "deployment",
						"kind": "Deployment"
					}
				]
			}`,
			want:    false,
			wantErr: false,
		},
		{
			name:     "invalid json",
			planJSON: `{"k8s_content_diff": [ invalid json`,
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKubernetesApprovalPlan([]byte(tt.planJSON))
			got, err := k.IsNoop()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
