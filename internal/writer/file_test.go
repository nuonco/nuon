package writer

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNewFile(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		opts        func(*testing.T) []fileEventWriterOption
		errExpected error
	}{
		"happy path": {
			v: v,
			opts: func(t *testing.T) []fileEventWriterOption {
				tmpDir := t.TempDir()
				f, err := os.Create(path.Join(tmpDir, "happypath"))
				if err != nil {
					t.FailNow()
				}

				return []fileEventWriterOption{
					WithFile(f),
				}
			},
		},
		"missing validator": {
			v: nil,
			opts: func(t *testing.T) []fileEventWriterOption {
				tmpDir := t.TempDir()
				f, err := os.Create(path.Join(tmpDir, "missingvalidator"))
				if err != nil {
					t.FailNow()
				}

				return []fileEventWriterOption{
					WithFile(f),
				}
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing file": {
			v:           v,
			opts:        func(t *testing.T) []fileEventWriterOption { return []fileEventWriterOption{} },
			errExpected: fmt.Errorf("Field validation for 'File' failed on the 'required' tag"),
		},
		"error on conifg": {
			v: v,
			opts: func(t *testing.T) []fileEventWriterOption {
				return []fileEventWriterOption{func(*fileEventWriter) error { return fmt.Errorf("error on config") }}
			},
			errExpected: fmt.Errorf("error on config"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q, err := NewFile(test.v, test.opts(t)...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, q)
		})
	}
}
