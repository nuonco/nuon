package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTerraformClient struct {
	mock.Mock
}

var (
	_ initer    = (*mockTerraformClient)(nil)
	_ planner   = (*mockTerraformClient)(nil)
	_ applier   = (*mockTerraformClient)(nil)
	_ destroyer = (*mockTerraformClient)(nil)
	_ outputter = (*mockTerraformClient)(nil)
)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		workingDir  string
		execPath    string
		backend     string
		vars        string
		create      func(*testing.T, string)
		errExpected error
	}{
		"valid": {
			v:          v,
			workingDir: "wd",
			execPath:   "tf",
			backend:    "backend",
			vars:       "vars",
			create: func(t *testing.T, tmpdir string) {
				err := os.MkdirAll(filepath.Join(tmpdir, "wd"), 0o777)
				assert.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpdir, "tf"), []byte(t.Name()), 0o600)
				assert.NoError(t, err)
			},
		},
		"missing validator": {
			v:          nil,
			workingDir: "wd",
			execPath:   "tf",
			backend:    "backend",
			vars:       "vars",
			create: func(t *testing.T, tmpdir string) {
				err := os.MkdirAll(filepath.Join(tmpdir, "wd"), 0o777)
				assert.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpdir, "tf"), []byte(t.Name()), 0o600)
				assert.NoError(t, err)
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing working dir": {
			v:          v,
			workingDir: "",
			execPath:   "tf",
			backend:    "backend",
			vars:       "vars",
			create: func(t *testing.T, tmpdir string) {
				err := os.WriteFile(filepath.Join(tmpdir, "tf"), []byte(t.Name()), 0o600)
				assert.NoError(t, err)
			},
			errExpected: fmt.Errorf("Field validation for 'WorkingDir' failed on the 'required' tag"),
		},
		"invalid working dir": {
			v:          v,
			workingDir: "wd",
			execPath:   "tf",
			backend:    "backend",
			vars:       "vars",
			create: func(t *testing.T, tmpdir string) {
				err := os.WriteFile(filepath.Join(tmpdir, "tf"), []byte(t.Name()), 0o600)
				assert.NoError(t, err)
			},
			errExpected: fmt.Errorf("Field validation for 'WorkingDir' failed on the 'dir' tag"),
		},
		"missing exec path": {
			v:          v,
			workingDir: "wd",
			execPath:   "",
			backend:    "backend",
			vars:       "vars",
			create: func(t *testing.T, tmpdir string) {
				err := os.MkdirAll(filepath.Join(tmpdir, "wd"), 0o777)
				assert.NoError(t, err)
			},
			errExpected: fmt.Errorf("Field validation for 'ExecPath' failed on the 'required' tag"),
		},
		"invalid exec path": {
			v:          v,
			workingDir: "wd",
			execPath:   "tf",
			backend:    "backend",
			vars:       "vars",
			create: func(t *testing.T, tmpdir string) {
				err := os.MkdirAll(filepath.Join(tmpdir, "wd"), 0o777)
				assert.NoError(t, err)
			},
			errExpected: fmt.Errorf("Field validation for 'ExecPath' failed on the 'file' tag"),
		},
		"missing backend config": {
			v:          v,
			workingDir: "wd",
			execPath:   "tf",
			backend:    "",
			vars:       "vars",
			create: func(t *testing.T, tmpdir string) {
				err := os.MkdirAll(filepath.Join(tmpdir, "wd"), 0o777)
				assert.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpdir, "tf"), []byte(t.Name()), 0o600)
				assert.NoError(t, err)
			},
			errExpected: fmt.Errorf("Field validation for 'BackendConfigFile' failed on the 'required' tag"),
		},
		"missing vars file": {
			v:          v,
			workingDir: "wd",
			execPath:   "tf",
			backend:    "backend",
			vars:       "",
			create: func(t *testing.T, tmpdir string) {
				err := os.MkdirAll(filepath.Join(tmpdir, "wd"), 0o777)
				assert.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpdir, "tf"), []byte(t.Name()), 0o600)
				assert.NoError(t, err)
			},
			errExpected: fmt.Errorf("Field validation for 'VarFile' failed on the 'required' tag"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpdir := t.TempDir()
			test.create(t, tmpdir)

			wd := test.workingDir
			if wd != "" {
				wd = filepath.Join(tmpdir, test.workingDir)
			}
			ep := test.execPath
			if ep != "" {
				ep = filepath.Join(tmpdir, test.execPath)
			}

			w, err := New(
				test.v,
				WithWorkingDir(wd),
				WithTerraformExecPath(ep),
				WithBackendConfigFile(test.backend),
				WithVarFile(test.vars),
			)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)
			assert.Equal(t, wd, w.WorkingDir)
			assert.Equal(t, ep, w.ExecPath)
			assert.Equal(t, test.backend, w.BackendConfigFile)
			assert.Equal(t, test.vars, w.VarFile)
		})
	}
}
