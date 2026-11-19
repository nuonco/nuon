package workspace

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTofu struct {
	Tofu

	applyOpt []ApplyOption
}

func (f *fakeTofu) ApplyJSON(_ context.Context, _ io.Writer, opts ...ApplyOption) error {
	f.applyOpt = opts
	return nil
}

func TestApplyPlanIncludesVarFiles(t *testing.T) {
	// Var files are unconditional with a saved plan: the minimum supported
	// tofu version is 1.13.0, which accepts -var-file with a plan file
	// argument (the plan file itself is passed last).
	const baseOpts = 2 // -refresh plus the saved plan itself

	client := &fakeTofu{}
	w := &workspace{
		v:         validator.New(),
		root:      t.TempDir(),
		varsPaths: []string{"vars-0.json", "vars-1.tfvars"},
	}

	_, err := w.applyPlan(context.Background(), client, hclog.NewNullLogger())
	require.NoError(t, err)

	assert.Len(t, client.applyOpt, baseOpts+len(w.varsPaths))

	var planFile string
	var varFileCount int
	for _, opt := range client.applyOpt {
		switch v := opt.(type) {
		case optDirOrPlan:
			planFile = string(v)
		case optVarFile:
			varFileCount++
		}
	}
	assert.Equal(t, filepath.Join(w.root, "tfplan"), planFile)
	assert.Equal(t, len(w.varsPaths), varFileCount)
}
