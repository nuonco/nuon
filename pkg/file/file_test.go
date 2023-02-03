package file

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		f           string
		v           *validator.Validate
		expected    *fileFetcher
		errExpected error
	}{
		"valid": {
			f:        "testdata/empty.txt",
			v:        v,
			expected: &fileFetcher{File: "testdata/empty.txt", validator: v},
		},
		"missing validator": {
			f:           "testdata/empty.txt",
			v:           nil,
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing file": {
			f:           "",
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'File' failed on the 'required' tag"),
		},
		"invalid file": {
			f:           "testdata/doesnotexist",
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'File' failed on the 'file' tag"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f, err := New(
				test.v,
				WithFile(test.f),
			)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, f)

			assert.Equal(t, test.expected, f)
		})
	}
}

func TestFile_Fetch(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		f           string
		errExpected error
		assertions  func(*testing.T, io.ReadCloser)
	}{
		"empty file": {
			f: "testdata/empty.txt",
			assertions: func(t *testing.T, iorc io.ReadCloser) {
				bs, err := io.ReadAll(iorc)
				assert.NoError(t, err)
				assert.Equal(t, "", string(bs))
			},
		},
		"lorem": {
			f: "testdata/lorem.txt",
			assertions: func(t *testing.T, iorc io.ReadCloser) {
				bs, err := io.ReadAll(iorc)
				assert.NoError(t, err)
				assert.Contains(t, string(bs), "Lorem ipsum")
				assert.Contains(t, string(bs), "id est laborum.")
			},
		},
		"file doesn't exist": {
			f:           "testdata/doesntexist",
			errExpected: fmt.Errorf("no such file or directory"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := &fileFetcher{File: test.f}
			iorc, err := f.Fetch(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			t.Cleanup(func() { _ = iorc.Close() })
			test.assertions(t, iorc)
		})
	}
}
