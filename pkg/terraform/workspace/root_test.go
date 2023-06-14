package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
	"github.com/stretchr/testify/assert"
)

func Test_workspace_createRoot(t *testing.T) {
	v := validator.New()

	arch := archive.NewMockArchive(nil)
	back := backend.NewMockBackend(nil)
	vars := variables.NewMockVariables(nil)
	bin := binary.NewMockBinary(nil)

	tests := map[string]struct {
		workspaceFn func(*testing.T) *workspace
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"happy path": {
			workspaceFn: func(t *testing.T) *workspace {
				w, err := New(v,
					WithBackend(back),
					WithArchive(arch),
					WithVariables(vars),
					WithBinary(bin),
				)
				assert.NoError(t, err)
				return w
			},
			assertFn: func(t *testing.T, w *workspace) {
				assert.NotEmpty(t, w.root)
				fileInfo, err := os.Stat(w.root)
				assert.NoError(t, err)
				assert.True(t, fileInfo.IsDir())
				assert.NoError(t, w.cleanup())
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			wkspace := test.workspaceFn(t)

			err := wkspace.createRoot()
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}

func Test_workspace_cleanup(t *testing.T) {
	v := validator.New()

	arch := archive.NewMockArchive(nil)
	back := backend.NewMockBackend(nil)
	vars := variables.NewMockVariables(nil)
	bin := binary.NewMockBinary(nil)

	tests := map[string]struct {
		workspaceFn func(*testing.T) *workspace
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"cleans up files and directories": {
			workspaceFn: func(t *testing.T) *workspace {
				w, err := New(v,
					WithBackend(back),
					WithArchive(arch),
					WithVariables(vars),
					WithBinary(bin),
				)
				assert.NoError(t, err)

				err = w.createRoot()
				assert.NoError(t, err)

				err = w.writeFile("test.txt", []byte("hello world"), defaultFilePermissions)
				assert.NoError(t, err)

				err = w.writeFile("test/test.txt", []byte("hello world"), defaultFilePermissions)
				assert.NoError(t, err)
				return w
			},
			assertFn: func(t *testing.T, w *workspace) {
				_, err := os.Stat(w.root)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
			errExpected: nil,
		},
		"disable cleanup does not cleanup": {
			workspaceFn: func(t *testing.T) *workspace {
				w, err := New(v,
					WithBackend(back),
					WithArchive(arch),
					WithVariables(vars),
					WithBinary(bin),
					WithDisableCleanup(true),
				)
				assert.NoError(t, err)

				err = w.createRoot()
				assert.NoError(t, err)

				err = w.writeFile("test.txt", []byte("hello world"), defaultFilePermissions)
				assert.NoError(t, err)

				err = w.writeFile("test/test.txt", []byte("hello world"), defaultFilePermissions)
				assert.NoError(t, err)
				return w
			},
			assertFn: func(t *testing.T, w *workspace) {
				_, err := os.Stat(w.root)
				assert.NoError(t, err)
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			wkspace := test.workspaceFn(t)

			err := wkspace.cleanup()
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}

func Test_workspace_writeFile(t *testing.T) {
	v := validator.New()

	arch := archive.NewMockArchive(nil)
	back := backend.NewMockBackend(nil)
	vars := variables.NewMockVariables(nil)
	bin := binary.NewMockBinary(nil)

	tests := map[string]struct {
		workspaceFn func(*testing.T) *workspace
		fileFn      func() (string, []byte)
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"writes a basic file": {
			workspaceFn: func(t *testing.T) *workspace {
				w, err := New(v,
					WithBackend(back),
					WithArchive(arch),
					WithVariables(vars),
					WithBinary(bin),
				)
				assert.NoError(t, err)
				return w
			},
			fileFn: func() (string, []byte) {
				return "test.txt", []byte("test.txt")
			},
			assertFn: func(t *testing.T, w *workspace) {
				fp := filepath.Join(w.root, "test.txt")
				info, err := os.Stat(fp)
				assert.NoError(t, err)
				assert.Equal(t, defaultFilePermissions, info.Mode().Perm())

				byts, err := os.ReadFile(fp)
				assert.NoError(t, err)
				assert.Equal(t, []byte("test.txt"), byts)
			},
			errExpected: nil,
		},
		"writes a directory": {
			workspaceFn: func(t *testing.T) *workspace {
				w, err := New(v,
					WithBackend(back),
					WithArchive(arch),
					WithVariables(vars),
					WithBinary(bin),
				)
				assert.NoError(t, err)
				return w
			},
			fileFn: func() (string, []byte) {
				return "dir/test.txt", []byte("test.txt")
			},
			assertFn: func(t *testing.T, w *workspace) {
				fp := filepath.Join(w.root, "dir/test.txt")
				info, err := os.Stat(fp)
				assert.NoError(t, err)
				assert.Equal(t, defaultFilePermissions, info.Mode().Perm())

				byts, err := os.ReadFile(fp)
				assert.NoError(t, err)
				assert.Equal(t, []byte("test.txt"), byts)
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			wkspace := test.workspaceFn(t)

			err := wkspace.createRoot()
			assert.NoError(t, err)
			file, contents := test.fileFn()

			err = wkspace.writeFile(file, contents, defaultFilePermissions)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}
