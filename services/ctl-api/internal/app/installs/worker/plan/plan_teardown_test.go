package plan

import (
	"testing"

	"github.com/stretchr/testify/require"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
)

func TestTeardownRetain(t *testing.T) {
	ccc := func(componentID string, retain ...string) app.ComponentConfigConnection {
		c := app.ComponentConfigConnection{ComponentID: componentID}
		if len(retain) > 0 {
			c.Teardown = &app.ComponentTeardown{Retain: retain}
		}
		return c
	}
	build := &app.ComponentBuild{ComponentConfigConnection: ccc("cmp-storage", "aws_rds_cluster.main")}

	tests := map[string]struct {
		appCfg *app.AppConfig
		want   []string
	}{
		"current app config wins over the deployed build": {
			appCfg: &app.AppConfig{ComponentConfigConnections: []app.ComponentConfigConnection{
				ccc("cmp-other", "aws_s3_bucket.other"),
				ccc("cmp-storage", "aws_rds_cluster.main", "module.buckets"),
			}},
			want: []string{"aws_rds_cluster.main", "module.buckets"},
		},
		"retain removed in current app config": {
			appCfg: &app.AppConfig{ComponentConfigConnections: []app.ComponentConfigConnection{
				ccc("cmp-storage"),
			}},
			want: nil,
		},
		"component removed from app config falls back to the deployed build": {
			appCfg: &app.AppConfig{ComponentConfigConnections: []app.ComponentConfigConnection{
				ccc("cmp-other", "aws_s3_bucket.other"),
			}},
			want: []string{"aws_rds_cluster.main"},
		},
		"no app config falls back to the deployed build": {
			want: []string{"aws_rds_cluster.main"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, teardownRetain(tc.appCfg, build))
		})
	}
}

func TestRenderTeardownPlan(t *testing.T) {
	stateData := map[string]any{
		"install": map[string]any{"id": "inst-abc123"},
		"inputs": map[string]any{
			"inputs": map[string]any{
				"retain_data":  "true",
				"discard_data": "false",
				"bucket_key":   "logs",
			},
		},
	}

	tests := map[string]struct {
		retain  []string
		want    *plantypes.TeardownPlan
		wantErr string
	}{
		"static entries": {
			retain: []string{"aws_rds_cluster.main", "module.buckets"},
			want:   &plantypes.TeardownPlan{Retain: []string{"aws_rds_cluster.main", "module.buckets"}},
		},
		"templated index": {
			retain: []string{`aws_s3_bucket.data["{{ .nuon.inputs.inputs.bucket_key }}"]`},
			want:   &plantypes.TeardownPlan{Retain: []string{`aws_s3_bucket.data["logs"]`}},
		},
		"conditional entry kept": {
			retain: []string{`{{ if eq .nuon.inputs.inputs.retain_data "true" }}aws_rds_cluster.main{{ end }}`},
			want:   &plantypes.TeardownPlan{Retain: []string{"aws_rds_cluster.main"}},
		},
		"conditional entry dropped": {
			retain: []string{
				"module.buckets",
				`{{ if eq .nuon.inputs.inputs.discard_data "true" }}aws_rds_cluster.main{{ end }}`,
			},
			want: &plantypes.TeardownPlan{Retain: []string{"module.buckets"}},
		},
		"every entry dropped": {
			retain: []string{`{{ if eq .nuon.inputs.inputs.discard_data "true" }}aws_rds_cluster.main{{ end }}`},
			want:   nil,
		},
		"surrounding whitespace is trimmed": {
			retain: []string{"{{ if eq .nuon.inputs.inputs.retain_data \"true\" }}\n  aws_rds_cluster.main\n{{ end }}"},
			want:   &plantypes.TeardownPlan{Retain: []string{"aws_rds_cluster.main"}},
		},
		"entries that render to the same address are deduplicated": {
			retain: []string{
				"aws_rds_cluster.main",
				`{{ if eq .nuon.inputs.inputs.retain_data "true" }}aws_rds_cluster.main{{ end }}`,
			},
			want: &plantypes.TeardownPlan{Retain: []string{"aws_rds_cluster.main"}},
		},
		"rendered address is invalid": {
			retain:  []string{"{{ .nuon.inputs.inputs.bucket_key }}"},
			wantErr: "invalid teardown.retain[0]",
		},
		"template fails to execute": {
			retain:  []string{"{{ .nuon.inputs.inputs.retain_data.nope }}"},
			wantErr: "unable to render teardown.retain[0]",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := RenderTeardownPlan(tc.retain, stateData)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.True(t, deployerrors.IsDeployPlanRenderFailed(err))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
