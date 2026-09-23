package terraform

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/terraform/workspace"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const oldProviderState = `{"version":4,"resources":[
  {"mode":"managed","type":"aws_eks_addon","name":"ebs_csi","instances":[{"schema_version":0,"attributes":{"resolve_conflicts":null}}]},
  {"module":"module.eks.module.nodes[\"acme\"]","mode":"managed","type":"aws_launch_template","name":"this","instances":[{"index_key":0,"schema_version":0,"attributes":{}}]}
]}`

const legacyPolicyState = `{"version":4,"resources":[
  {"mode":"managed","type":"kubectl_manifest","name":"vendor_policies","instances":[
    {"index_key":"0.yaml","attributes":{"yaml_body":"kind: ClusterRole\nmetadata:\n  name: acme:manage\n"}},
    {"index_key":"1.yaml","attributes":{"yaml_body":"kind: ClusterPolicy\nmetadata:\n  name: acme-authz\n"}}
  ]}
]}`

func TestFindLegacyPolicyKeyMigrations(t *testing.T) {
	for name, state := range map[string]string{
		"new workspace":        "",
		"empty workspace":      `{"version":4,"resources":[]}`,
		"old provider schemas": oldProviderState,
		"already migrated":     `{"version":4,"resources":[{"mode":"managed","type":"kubectl_manifest","name":"vendor_policies","instances":[{"index_key":"clusterrole-acme.yaml","attributes":{}}]}]}`,
		"unrelated resources": `{"version":4,"resources":[
          {"mode":"data","type":"kubectl_manifest","name":"vendor_policies","instances":[{"index_key":"0.yaml"}]},
          {"mode":"managed","type":"kubectl_manifest","name":"other_policies","instances":[{"index_key":"0.yaml"}]},
          {"mode":"managed","type":"other_resource","name":"vendor_policies","instances":[{"index_key":"0.yaml"}]},
          {"mode":"managed","type":"kubectl_manifest","name":"vendor_policies","instances":[{"index_key":0},{"index_key":"0.yaml","deposed":"deadbeef"},{}]}
        ]}`,
	} {
		t.Run(name, func(t *testing.T) {
			moves, err := findLegacyPolicyKeyMigrations(state)
			require.NoError(t, err)
			assert.Empty(t, moves)
		})
	}

	t.Run("legacy keys", func(t *testing.T) {
		moves, err := findLegacyPolicyKeyMigrations(legacyPolicyState)
		require.NoError(t, err)
		assert.Equal(t, []policyKeyMigration{
			{sourceAddress: `kubectl_manifest.vendor_policies["0.yaml"]`, destinationAddress: `kubectl_manifest.vendor_policies["clusterrole-acme-manage.yaml"]`, oldKey: "0.yaml", newKey: "clusterrole-acme-manage.yaml", kind: "ClusterRole", name: "acme:manage"},
			{sourceAddress: `kubectl_manifest.vendor_policies["1.yaml"]`, destinationAddress: `kubectl_manifest.vendor_policies["clusterpolicy-acme-authz.yaml"]`, oldKey: "1.yaml", newKey: "clusterpolicy-acme-authz.yaml", kind: "ClusterPolicy", name: "acme-authz"},
		}, moves)
	})

	t.Run("indexed nested module with mixed keys and old provider state", func(t *testing.T) {
		state := strings.Replace(oldProviderState, `"resources":[`, `"resources":[
          {"module":"module.sandbox[\"acme\"].module.policies[1]","mode":"managed","type":"kubectl_manifest","name":"vendor_policies","instances":[
            {"index_key":"2.yaml","attributes":{"yaml_body":"kind: ClusterRole\nmetadata:\n  name: nested\n"}},
            {"index_key":"clusterpolicy-existing.yaml","attributes":{}}
          ]},`, 1)
		moves, err := findLegacyPolicyKeyMigrations(state)
		require.NoError(t, err)
		require.Len(t, moves, 1)
		assert.Equal(t, `module.sandbox["acme"].module.policies[1].kubectl_manifest.vendor_policies["2.yaml"]`, moves[0].sourceAddress)
		assert.Equal(t, `module.sandbox["acme"].module.policies[1].kubectl_manifest.vendor_policies["clusterrole-nested.yaml"]`, moves[0].destinationAddress)
	})

	for name, tc := range map[string]struct{ state, errorText string }{
		"invalid JSON":          {`{"version":`, "unable to decode state"},
		"unsupported version":   {`{"version":5}`, "unsupported state version 5"},
		"not native state":      {`{"format_version":"1.0","values":{}}`, "unsupported state version 0"},
		"missing manifest":      {`{"version":4,"resources":[{"mode":"managed","type":"kubectl_manifest","name":"vendor_policies","instances":[{"index_key":"0.yaml"}]}]}`, "no yaml_body attribute"},
		"invalid manifest type": {`{"version":4,"resources":[{"mode":"managed","type":"kubectl_manifest","name":"vendor_policies","instances":[{"index_key":"0.yaml","attributes":{"yaml_body":123}}]}]}`, "unable to decode policy attributes"},
		"invalid YAML":          {`{"version":4,"resources":[{"mode":"managed","type":"kubectl_manifest","name":"vendor_policies","instances":[{"index_key":"0.yaml","attributes":{"yaml_body":"kind: ["}}]}]}`, "unable to derive new key"},
	} {
		t.Run(name, func(t *testing.T) {
			moves, err := findLegacyPolicyKeyMigrations(tc.state)
			require.ErrorContains(t, err, tc.errorText)
			assert.Empty(t, moves)
		})
	}
}

type policyMigrationWorkspace struct {
	workspace.Workspace
	state   string
	pullErr error
	moveErr error
	pulls   int
	moves   [][2]string
}

func (w *policyMigrationWorkspace) StatePull(context.Context, hclog.Logger) (string, error) {
	w.pulls++
	return w.state, w.pullErr
}

func (w *policyMigrationWorkspace) StateMv(_ context.Context, _ hclog.Logger, source, destination string) error {
	w.moves = append(w.moves, [2]string{source, destination})
	return w.moveErr
}

func TestMigrateLegacyPolicyKeys(t *testing.T) {
	ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
	for name, tc := range map[string]struct {
		state            string
		pullErr, moveErr error
		wantMoves        int
		errorText        string
	}{
		"new workspace":                            {},
		"old provider state bypasses show":         {state: oldProviderState},
		"migrates legacy policies":                 {state: legacyPolicyState, wantMoves: 2},
		"pull failure":                             {pullErr: errors.New("backend unavailable"), errorText: "unable to read state for policy key migration: backend unavailable"},
		"invalid state":                            {state: "{", errorText: "unable to decode state"},
		"invalid second policy prevents all moves": {state: strings.Replace(legacyPolicyState, `kind: ClusterPolicy\nmetadata:\n  name: acme-authz\n`, "", 1), errorText: "no yaml_body attribute"},
		"move failure stops migration":             {state: legacyPolicyState, moveErr: errors.New("locked"), wantMoves: 1, errorText: "unable to migrate policy state key"},
	} {
		t.Run(name, func(t *testing.T) {
			ws := &policyMigrationWorkspace{state: tc.state, pullErr: tc.pullErr, moveErr: tc.moveErr}
			err := (&handler{}).migrateLegacyPolicyKeys(ctx, hclog.NewNullLogger(), ws)
			if tc.errorText != "" {
				require.ErrorContains(t, err, tc.errorText)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, 1, ws.pulls)
			require.Len(t, ws.moves, tc.wantMoves)
			if tc.wantMoves > 0 {
				assert.Equal(t, [2]string{`kubectl_manifest.vendor_policies["0.yaml"]`, `kubectl_manifest.vendor_policies["clusterrole-acme-manage.yaml"]`}, ws.moves[0])
			}
			if tc.wantMoves > 1 {
				assert.Equal(t, [2]string{`kubectl_manifest.vendor_policies["1.yaml"]`, `kubectl_manifest.vendor_policies["clusterpolicy-acme-authz.yaml"]`}, ws.moves[1])
			}
		})
	}
}
