package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustArgs materializes an args builder for assertions.
func mustArgs(t *testing.T, fn func() ([]string, error)) []string {
	t.Helper()
	args, err := fn()
	require.NoError(t, err)
	return args
}

func TestApplyArgs(t *testing.T) {
	tests := map[string]struct {
		opts    []ApplyOption
		want    []string
		wantErr string
	}{
		"refresh plus var files": {
			opts: []ApplyOption{Refresh(true), VarFile("a.tfvars"), VarFile("b.tfvars")},
			want: []string{"apply", "-auto-approve", "-refresh=true", "-var-file=a.tfvars", "-var-file=b.tfvars", "-json"},
		},
		"saved plan": {
			opts: []ApplyOption{Refresh(true), DirOrPlan("/tmp/ws/tfplan")},
			want: []string{"apply", "-auto-approve", "-refresh=true", "-json", "/tmp/ws/tfplan"},
		},
		"saved plan plus var files": {
			opts: []ApplyOption{Refresh(true), DirOrPlan("tfplan"), VarFile("vars-0.json")},
			want: []string{"apply", "-auto-approve", "-refresh=true", "-var-file=vars-0.json", "-json", "tfplan"},
		},
		"destroy via apply": {
			opts: []ApplyOption{Refresh(true), Destroy(true), VarFile("vars-0.json")},
			want: []string{"apply", "-auto-approve", "-refresh=true", "-destroy", "-var-file=vars-0.json", "-json"},
		},
		"destroy with a saved plan is rejected": {
			opts:    []ApplyOption{DirOrPlan("tfplan"), Destroy(true)},
			wantErr: "cannot apply a saved plan file with -destroy",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			o := &applyOpts{}
			require.NoError(t, o.add(tc.opts...))

			args, err := o.args()
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, args)
		})
	}
}

func TestPlanArgs(t *testing.T) {
	tests := map[string]struct {
		opts []PlanOption
		want []string
	}{
		"refresh, out and var files": {
			opts: []PlanOption{Refresh(true), Out("tfplan"), VarFile("vars-0.json")},
			want: []string{"plan", "-refresh=true", "-out=tfplan", "-var-file=vars-0.json", "-json"},
		},
		"destroy": {
			opts: []PlanOption{Refresh(true), Destroy(true)},
			want: []string{"plan", "-refresh=true", "-destroy", "-json"},
		},
		"refresh only": {
			opts: []PlanOption{RefreshOnly(true)},
			want: []string{"plan", "-refresh=true", "-refresh-only", "-json"},
		},
		"targeted destroy": {
			opts: []PlanOption{Destroy(true), Target("aws_rds_cluster.main"), Target(`module.buckets.aws_s3_bucket.data["logs"]`)},
			want: []string{
				"plan",
				"-refresh=true",
				"-destroy",
				"-target=aws_rds_cluster.main",
				`-target=module.buckets.aws_s3_bucket.data["logs"]`,
				"-json",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			o := &planOpts{}
			require.NoError(t, o.add(tc.opts...))

			args, err := o.args()
			require.NoError(t, err)
			assert.Equal(t, tc.want, args)
		})
	}
}

func TestDestroyAndRefreshArgs(t *testing.T) {
	d := &destroyOpts{}
	require.NoError(t, d.add(Refresh(true), VarFile("vars-0.json")))
	assert.Equal(t, []string{"destroy", "-auto-approve", "-refresh=true", "-var-file=vars-0.json", "-json"}, mustArgs(t, d.args))

	r := &refreshOpts{}
	require.NoError(t, r.add(VarFile("vars-0.json")))
	assert.Equal(t, []string{"refresh", "-var-file=vars-0.json", "-json"}, mustArgs(t, r.args))
}

func TestBoolArg(t *testing.T) {
	tr := true
	fa := false
	assert.Equal(t, "true", boolArg(&tr))
	assert.Equal(t, "false", boolArg(&fa))
	// unset falls back to terraform's default of refreshing
	assert.Equal(t, "true", boolArg(nil))
}

func TestEnvSlice(t *testing.T) {
	client := newTofuClient("", "", map[string]string{"FOO": "bar"}, nil)
	assert.Contains(t, client.envSlice(), "FOO=bar")

	// nil env mirrors tfexec's explicit-nil SetEnv: copy os.Environ.
	assert.NotEmpty(t, newTofuClient("", "", nil, nil).envSlice())
}
