package manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		id          string
		v           *validator.Validate
		expected    *manager
		errExpected error
	}{
		"valid": {
			id:       "valid",
			v:        v,
			expected: &manager{ID: "valid", pattern: "nuon-module-valid", validator: v},
		},
		"missing id": {
			id:          "",
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'ID' failed on the 'required' tag"),
		},
		"missing validator": {
			id:          "valid",
			v:           nil,
			errExpected: fmt.Errorf("validator is nil"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w, err := New(test.v, WithID(test.id))
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)
			assert.Equal(t, test.expected, w)
		})
	}
}

func TestDirwriter_Init(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		w           *manager
		errExpected error
	}{
		"valid": {
			w: &manager{ID: "valid", pattern: "valid"},
		},
		"pattern can't be a directory": {
			w:           &manager{ID: "valid", pattern: "this/is/not/valid"},
			errExpected: fmt.Errorf("pattern contains path separator"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cleanup, err := test.w.Init(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Contains(t, test.w.tmpDir, test.w.pattern)
			_ = cleanup()
		})
	}
}

func TestDirwriter_GetWorkingDir(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		wd          string
		expected    string
		errExpected error
	}{
		"valid":               {wd: "set", expected: "set"},
		"not yet initialized": {wd: "", errExpected: fmt.Errorf("working directory unset")},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := &manager{tmpDir: test.wd}
			got, err := m.GetWorkingDir()
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestDirwriter_GetWriter(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		filename    string
		w           *manager
		assertions  func(*testing.T, io.WriteCloser)
		errExpected error
	}{
		"valid": {
			filename: "valid.txt",
			w:        &manager{ID: "valid", pattern: "valid"},
			assertions: func(t *testing.T, w io.WriteCloser) {
				assert.NotNil(t, w)
				defer w.Close()

				fname := w.(*workspaceWriteCloser).f.Name()
				_, err := w.Write([]byte(t.Name()))
				assert.NoError(t, err)
				assert.FileExists(t, fname)

				bs, err := os.ReadFile(fname)
				assert.NoError(t, err)
				assert.Equal(t, t.Name(), string(bs))
			},
		},
		"directory root returns nil": {
			filename: "",
			w:        &manager{ID: "valid", pattern: "valid"},
			assertions: func(t *testing.T, w io.WriteCloser) {
				assert.Nil(t, w)
			},
		},
		// NOTE(jdt): testing os.Abs errors is pretty hard across environments
		// NOTE(jdt): testing os.OpenFile errors is pretty hard across environments
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			td := t.TempDir()
			test.w.tmpDir = td

			iowc, err := test.w.GetWriter(test.filename)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			test.assertions(t, iowc)
		})
	}
}
func TestDirwriter_cleanup(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		w          func(string) *manager
		assertions func(*testing.T, string)

		errExpected error
	}{
		"valid": {
			w: func(s string) *manager {
				return &manager{tmpDir: s}
			},
			assertions: func(t *testing.T, s string) {
				assert.NoDirExists(t, s)
			},
		},
		"temp dir does not exist": {
			w: func(s string) *manager {
				return &manager{tmpDir: "/home/user/doesnotexist"}
			},
			assertions: func(t *testing.T, s string) {
				// NOTE(jdt): removeall doesn't error on directory not existing
				// just make sure it didn't get created?
				assert.NoDirExists(t, "/home/user/doesnotexist")
			},
		},
		"temp dir invalid": {
			w: func(s string) *manager {
				return &manager{tmpDir: "."}
			},
			errExpected: fmt.Errorf("invalid argument"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := t.TempDir()
			w := test.w(d)

			err := w.cleanup()
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertions(t, d)
		})
	}
}
