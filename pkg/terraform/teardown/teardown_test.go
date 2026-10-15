package teardown

import (
	"encoding/json"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testStateJSON = `{
  "format_version": "1.0",
  "terraform_version": "1.11.3",
  "values": {
    "root_module": {
      "resources": [
        {
          "address": "aws_rds_cluster.main",
          "mode": "managed",
          "type": "aws_rds_cluster",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "values": {"cluster_identifier": "acme-db", "master_password": "hunter2"},
          "sensitive_values": {"master_password": true}
        },
        {
          "address": "aws_db_subnet_group.main",
          "mode": "managed",
          "type": "aws_db_subnet_group",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "values": {"name": "acme-db"}
        },
        {
          "address": "data.aws_caller_identity.current",
          "mode": "data",
          "type": "aws_caller_identity",
          "name": "current",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "values": {}
        }
      ],
      "child_modules": [
        {
          "address": "module.buckets",
          "resources": [
            {
              "address": "module.buckets.aws_s3_bucket.data[\"logs\"]",
              "mode": "managed",
              "type": "aws_s3_bucket",
              "name": "data",
              "index": "logs",
              "provider_name": "registry.terraform.io/hashicorp/aws",
              "values": {"bucket": "acme-logs"}
            }
          ]
        }
      ]
    }
  }
}`

func testState(t *testing.T) *tfjson.State {
	t.Helper()

	var state tfjson.State
	require.NoError(t, json.Unmarshal([]byte(testStateJSON), &state))
	return &state
}

func TestManagedAddresses(t *testing.T) {
	tests := map[string]struct {
		state *tfjson.State
		want  []string
	}{
		"nil state": {
			state: nil,
			want:  []string{},
		},
		"empty state": {
			state: &tfjson.State{},
			want:  []string{},
		},
		"skips data sources and walks child modules": {
			state: testState(t),
			want: []string{
				"aws_db_subnet_group.main",
				"aws_rds_cluster.main",
				`module.buckets.aws_s3_bucket.data["logs"]`,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, ManagedAddresses(tc.state))
		})
	}
}

func TestPartition(t *testing.T) {
	stateAddrs := []string{
		"aws_db_subnet_group.main",
		"aws_rds_cluster.main",
		`module.buckets.aws_s3_bucket.data["logs"]`,
		`module.buckets.aws_s3_bucket.data["assets"]`,
	}

	tests := map[string]struct {
		retain        []string
		wantRetained  []string
		wantDestroy   []string
		wantUnmatched []string
		wantErr       string
	}{
		"no retain entries destroys everything": {
			wantDestroy: stateAddrs,
		},
		"single resource": {
			retain:       []string{"aws_rds_cluster.main"},
			wantRetained: []string{"aws_rds_cluster.main"},
			wantDestroy: []string{
				"aws_db_subnet_group.main",
				`module.buckets.aws_s3_bucket.data["logs"]`,
				`module.buckets.aws_s3_bucket.data["assets"]`,
			},
		},
		"module retains everything inside it": {
			retain: []string{"module.buckets"},
			wantRetained: []string{
				`module.buckets.aws_s3_bucket.data["logs"]`,
				`module.buckets.aws_s3_bucket.data["assets"]`,
			},
			wantDestroy: []string{"aws_db_subnet_group.main", "aws_rds_cluster.main"},
		},
		"indexed entry retains one instance": {
			retain:       []string{`module.buckets.aws_s3_bucket.data["logs"]`},
			wantRetained: []string{`module.buckets.aws_s3_bucket.data["logs"]`},
			wantDestroy: []string{
				"aws_db_subnet_group.main",
				"aws_rds_cluster.main",
				`module.buckets.aws_s3_bucket.data["assets"]`,
			},
		},
		"overlapping entries retain an address once": {
			retain: []string{"module.buckets", "module.buckets.aws_s3_bucket.data"},
			wantRetained: []string{
				`module.buckets.aws_s3_bucket.data["logs"]`,
				`module.buckets.aws_s3_bucket.data["assets"]`,
			},
			wantDestroy: []string{"aws_db_subnet_group.main", "aws_rds_cluster.main"},
		},
		"everything retained": {
			retain:       []string{"aws_db_subnet_group.main", "aws_rds_cluster.main", "module.buckets"},
			wantRetained: stateAddrs,
		},
		"unmatched entries are reported": {
			retain:        []string{"aws_rds_cluster.main", "aws_elasticache_cluster.cache"},
			wantRetained:  []string{"aws_rds_cluster.main"},
			wantUnmatched: []string{"aws_elasticache_cluster.cache"},
			wantDestroy: []string{
				"aws_db_subnet_group.main",
				`module.buckets.aws_s3_bucket.data["logs"]`,
				`module.buckets.aws_s3_bucket.data["assets"]`,
			},
		},
		"invalid entry": {
			retain:  []string{"aws_rds_cluster"},
			wantErr: "expected a resource address",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			retained, destroy, unmatched, err := Partition(stateAddrs, tc.retain)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantRetained, retained)
			assert.Equal(t, tc.wantDestroy, destroy)
			assert.Equal(t, tc.wantUnmatched, unmatched)
		})
	}
}

func resourceChange(addr, mode string, actions ...tfjson.Action) *tfjson.ResourceChange {
	return &tfjson.ResourceChange{
		Address: addr,
		Mode:    tfjson.ResourceMode(mode),
		Change:  &tfjson.Change{Actions: actions},
	}
}

func TestCheckPlan(t *testing.T) {
	tests := map[string]struct {
		plan    *tfjson.Plan
		retain  []string
		wantErr string
	}{
		"nil plan": {
			retain: []string{"aws_rds_cluster.main"},
		},
		"no retain entries": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				resourceChange("aws_rds_cluster.main", "managed", tfjson.ActionDelete),
			}},
		},
		"only non-retained resources deleted": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				resourceChange("aws_db_subnet_group.main", "managed", tfjson.ActionDelete),
			}},
			retain: []string{"aws_rds_cluster.main"},
		},
		"retained resource with a non-destructive action": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				resourceChange("aws_rds_cluster.main", "managed", tfjson.ActionNoop),
				resourceChange("aws_rds_cluster.main", "managed", tfjson.ActionForget),
			}},
			retain: []string{"aws_rds_cluster.main"},
		},
		"data source matching an entry is ignored": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				resourceChange("module.buckets.data.aws_iam_policy_document.read", "data", tfjson.ActionDelete),
			}},
			retain: []string{"module.buckets"},
		},
		"retained resource deleted": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				resourceChange("aws_db_subnet_group.main", "managed", tfjson.ActionDelete),
				resourceChange("aws_rds_cluster.main", "managed", tfjson.ActionDelete),
			}},
			retain:  []string{"aws_rds_cluster.main"},
			wantErr: "would destroy retained resources: aws_rds_cluster.main",
		},
		"retained resource replaced": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				resourceChange(`module.buckets.aws_s3_bucket.data["logs"]`, "managed", tfjson.ActionDelete, tfjson.ActionCreate),
			}},
			retain:  []string{"module.buckets"},
			wantErr: `module.buckets.aws_s3_bucket.data["logs"]`,
		},
		"retained resource replaced create before destroy": {
			plan: &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{
				resourceChange("aws_rds_cluster.main", "managed", tfjson.ActionCreate, tfjson.ActionDelete),
			}},
			retain:  []string{"aws_rds_cluster.main"},
			wantErr: "aws_rds_cluster.main",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := CheckPlan(tc.plan, tc.retain)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAddForgetChanges(t *testing.T) {
	state := testState(t)

	tests := map[string]struct {
		retained []string
		want     []*tfjson.ResourceChange
	}{
		"nothing retained": {},
		"root resource carries attributes from state": {
			retained: []string{"aws_rds_cluster.main"},
			want: []*tfjson.ResourceChange{{
				Address:      "aws_rds_cluster.main",
				Mode:         tfjson.ManagedResourceMode,
				Type:         "aws_rds_cluster",
				Name:         "main",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Change: &tfjson.Change{
					Actions:         tfjson.Actions{tfjson.ActionForget},
					Before:          map[string]any{"cluster_identifier": "acme-db", "master_password": "hunter2"},
					BeforeSensitive: map[string]any{"master_password": true},
				},
			}},
		},
		"module resource carries module address and index": {
			retained: []string{`module.buckets.aws_s3_bucket.data["logs"]`},
			want: []*tfjson.ResourceChange{{
				Address:       `module.buckets.aws_s3_bucket.data["logs"]`,
				ModuleAddress: "module.buckets",
				Mode:          tfjson.ManagedResourceMode,
				Type:          "aws_s3_bucket",
				Name:          "data",
				Index:         "logs",
				ProviderName:  "registry.terraform.io/hashicorp/aws",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionForget},
					Before:  map[string]any{"bucket": "acme-logs"},
				},
			}},
		},
		"address missing from state still gets a change": {
			retained: []string{"aws_s3_bucket.gone"},
			want: []*tfjson.ResourceChange{{
				Address: "aws_s3_bucket.gone",
				Mode:    tfjson.ManagedResourceMode,
				Change:  &tfjson.Change{Actions: tfjson.Actions{tfjson.ActionForget}},
			}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			existing := resourceChange("aws_db_subnet_group.main", "managed", tfjson.ActionDelete)
			plan := &tfjson.Plan{ResourceChanges: []*tfjson.ResourceChange{existing}}

			AddForgetChanges(plan, state, tc.retained)

			require.Equal(t, existing, plan.ResourceChanges[0])
			require.Equal(t, tc.want, nilIfEmpty(plan.ResourceChanges[1:]))
		})
	}
}

func nilIfEmpty(changes []*tfjson.ResourceChange) []*tfjson.ResourceChange {
	if len(changes) == 0 {
		return nil
	}
	return changes
}
