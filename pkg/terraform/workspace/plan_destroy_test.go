package workspace

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePlanTerraform struct {
	Terraform

	planOpts []tfexec.PlanOption
	stateRm  []string
	rmErr    error
}

func (f *fakePlanTerraform) PlanJSON(_ context.Context, _ io.Writer, opts ...tfexec.PlanOption) (bool, error) {
	f.planOpts = opts
	return true, nil
}

func (f *fakePlanTerraform) StateRm(_ context.Context, address string, _ ...tfexec.StateRmCmdOption) error {
	if f.rmErr != nil {
		return f.rmErr
	}
	f.stateRm = append(f.stateRm, address)
	return nil
}

func TestPlanDestroyModes(t *testing.T) {
	tests := map[string]struct {
		modeOpts []tfexec.PlanOption
		want     []tfexec.PlanOption
		notWant  []tfexec.PlanOption
	}{
		"destroy": {
			modeOpts: []tfexec.PlanOption{tfexec.Destroy(true)},
			want:     []tfexec.PlanOption{tfexec.Destroy(true)},
			notWant:  []tfexec.PlanOption{tfexec.RefreshOnly(true)},
		},
		"targeted destroy": {
			modeOpts: []tfexec.PlanOption{
				tfexec.Destroy(true),
				tfexec.Target("aws_db_subnet_group.main"),
				tfexec.Target(`module.buckets.aws_s3_bucket.data["logs"]`),
			},
			want: []tfexec.PlanOption{
				tfexec.Destroy(true),
				tfexec.Target("aws_db_subnet_group.main"),
				tfexec.Target(`module.buckets.aws_s3_bucket.data["logs"]`),
			},
		},
		"refresh only": {
			modeOpts: []tfexec.PlanOption{tfexec.RefreshOnly(true)},
			want:     []tfexec.PlanOption{tfexec.RefreshOnly(true)},
			notWant:  []tfexec.PlanOption{tfexec.Destroy(true)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &fakePlanTerraform{}
			w := &workspace{
				v:         validator.New(),
				root:      t.TempDir(),
				varsPaths: []string{"vars-0.json"},
			}

			_, err := w.planDestroy(context.Background(), client, hclog.NewNullLogger(), tc.modeOpts...)
			require.NoError(t, err)

			assert.Contains(t, client.planOpts, tfexec.Refresh(true))
			assert.Contains(t, client.planOpts, tfexec.Out(filepath.Join(w.root, "tfplan")))
			assert.Contains(t, client.planOpts, tfexec.VarFile("vars-0.json"))
			for _, opt := range tc.want {
				assert.Contains(t, client.planOpts, opt)
			}
			for _, opt := range tc.notWant {
				assert.NotContains(t, client.planOpts, opt)
			}
		})
	}
}

func TestPlanDestroyTargetsRequiresTargets(t *testing.T) {
	w := &workspace{v: validator.New(), root: t.TempDir()}

	_, err := w.PlanDestroyTargets(context.Background(), hclog.NewNullLogger(), nil)
	require.ErrorContains(t, err, "untargeted destroy plan destroys everything")
}

func TestStateRm(t *testing.T) {
	tests := map[string]struct {
		addrs   []string
		rmErr   error
		want    []string
		wantErr string
	}{
		"no addresses": {},
		"removes each address": {
			addrs: []string{"aws_rds_cluster.main", `module.buckets.aws_s3_bucket.data["logs"]`},
			want:  []string{"aws_rds_cluster.main", `module.buckets.aws_s3_bucket.data["logs"]`},
		},
		"error names the address": {
			addrs:   []string{"aws_rds_cluster.main"},
			rmErr:   errors.New("state locked"),
			wantErr: `unable to run state rm "aws_rds_cluster.main": state locked`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &fakePlanTerraform{rmErr: tc.rmErr}
			w := &workspace{v: validator.New(), root: t.TempDir()}

			err := w.stateRm(context.Background(), client, tc.addrs)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, client.stateRm)
		})
	}
}
