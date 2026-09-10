package workspace

import (
	"context"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTerraform struct {
	Terraform

	version  string
	applyOpt []tfexec.ApplyOption
}

func (f *fakeTerraform) Version(context.Context, bool) (*goversion.Version, map[string]*goversion.Version, error) {
	return goversion.Must(goversion.NewVersion(f.version)), nil, nil
}

func (f *fakeTerraform) ApplyJSON(_ context.Context, _ io.Writer, opts ...tfexec.ApplyOption) error {
	f.applyOpt = opts
	return nil
}

func TestApplyPlanVarFilesByTerraformVersion(t *testing.T) {
	const baseOpts = 2 // -refresh plus the saved plan itself

	for _, tc := range []struct {
		tfVersion    string
		wantVarFiles bool
	}{
		{"1.5.7", false},
		{"1.7.5", false},
		{"1.9.0", false},
		{"1.9.8", false},
		{"1.10.0", true},
		{"1.11.3", true},
		{"1.13.0", true},
		{"1.10.0-rc1", true},
	} {
		t.Run(tc.tfVersion, func(t *testing.T) {
			client := &fakeTerraform{version: tc.tfVersion}
			w := &workspace{
				v:         validator.New(),
				root:      t.TempDir(),
				varsPaths: []string{"vars-0.json", "vars-1.tfvars"},
			}

			_, err := w.applyPlan(context.Background(), client, hclog.NewNullLogger())
			require.NoError(t, err)

			want := baseOpts
			if tc.wantVarFiles {
				want += len(w.varsPaths)
			}
			assert.Len(t, client.applyOpt, want)
		})
	}
}
