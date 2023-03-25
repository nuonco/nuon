package module

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	errfetch "github.com/powertoolsdev/mono/pkg/fetch/error"
	filefetch "github.com/powertoolsdev/mono/pkg/fetch/file"
	stringfetch "github.com/powertoolsdev/mono/pkg/fetch/string"
	"github.com/stretchr/testify/assert"
)

const (
	emptyTarGz = "testdata/empty_0.8.33.tar.gz"
	emptyTar   = "testdata/empty_0.8.33.tar"
)

type testWriteFactory struct {
	d string
}

func (f *testWriteFactory) GetWriter(path string) (io.WriteCloser, error) {
	fp, err := filepath.Abs(filepath.Join(f.d, path))
	if err != nil {
		return nil, err
	}
	if fp == f.d {
		return nil, nil
	}
	return os.OpenFile(fp, os.O_CREATE|os.O_RDWR, 0o777)
}

var _ writeFactory = (*testWriteFactory)(nil)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		fetcher     func(*testing.T) fetcher
		wf          writeFactory
		v           *validator.Validate
		errExpected error
	}{
		"valid": {
			fetcher: func(t *testing.T) fetcher {
				return stringfetch.New(t.Name())
			},
			wf: &testWriteFactory{"valid"},
			v:  v,
		},
		"missing write factory": {
			fetcher: func(t *testing.T) fetcher {
				return stringfetch.New(t.Name())
			},
			v:           v,
			errExpected: fmt.Errorf("Error:Field validation for 'WriteFactory' failed on the 'required' tag"),
		},
		"missing validator": {
			fetcher: func(t *testing.T) fetcher {
				return stringfetch.New(t.Name())
			},
			wf:          &testWriteFactory{"valid"},
			v:           nil,
			errExpected: fmt.Errorf("validator is nil"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f, err := New(
				test.v,
				WithFetcher(test.fetcher(t)),
				WithWriteFactory(test.wf),
			)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, f)
		})
	}
}

func TestModule_Install(t *testing.T) {
	// TODO(jm): fix these once they work correctly
	return
	t.Parallel()
	tests := map[string]struct {
		fetcher     func(*testing.T) fetcher
		assertions  func(*testing.T, string)
		errExpected error
	}{
		"happy path": {
			fetcher: func(t *testing.T) fetcher {
				f, err := filefetch.New(validator.New(), filefetch.WithFile(emptyTarGz))
				assert.NoError(t, err)
				return f
			},
			assertions: func(t *testing.T, td string) {
				// NOTE(jdt): just diff the output directories
				err := exec.Command("diff", "-r", td, "testdata/outdir").Run()
				assert.NoError(t, err)
			},
		},
		"fetch failure": {
			fetcher: func(t *testing.T) fetcher {
				return errfetch.New(fmt.Errorf("fetch failure"))
			},
			errExpected: fmt.Errorf("fetch failure"),
		},
		"extract failure": {
			fetcher: func(t *testing.T) fetcher {
				f, err := filefetch.New(validator.New(), filefetch.WithFile(emptyTar))
				assert.NoError(t, err)
				return f
			},
			errExpected: fmt.Errorf("gzip: invalid header"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpdir := t.TempDir()
			fetcher := test.fetcher(t)

			m := &module{Fetcher: fetcher, WriteFactory: &testWriteFactory{d: tmpdir}}
			err := m.Install(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertions(t, tmpdir)
		})
	}
}

func TestModule_extractModule(t *testing.T) {
	return
	t.Parallel()
	ctx := context.Background()

	tests := map[string]struct {
		archive     func(*testing.T) io.Reader
		assertions  func(*testing.T, string)
		errExpected error
	}{
		"happy path": {
			archive: func(t *testing.T) io.Reader {
				f, err := os.Open(emptyTarGz)
				assert.NoError(t, err)

				t.Cleanup(func() { assert.NoError(t, f.Close()) })

				return f
			},
			assertions: func(t *testing.T, tmpDir string) {
				// NOTE(jdt): just diff the output directories
				err := exec.Command("diff", "-r", tmpDir, "testdata/outdir").Run()
				assert.NoError(t, err)
			},
		},
		"not gzipped errors": {
			archive: func(t *testing.T) io.Reader {
				f, err := os.Open(emptyTar)
				assert.NoError(t, err)

				t.Cleanup(func() { assert.NoError(t, f.Close()) })

				return f
			},
			errExpected: fmt.Errorf("gzip: invalid header"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			s := &module{WriteFactory: &testWriteFactory{d: tmpDir}}

			ior := test.archive(t)
			err := s.extractModule(ctx, ior)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertions(t, tmpDir)
		})
	}
}
