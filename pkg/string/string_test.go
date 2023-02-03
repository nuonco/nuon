package string

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringFetcher_Fetch(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		s string
	}{
		"empty string": {s: ""},
		"single char":  {s: "a"},
		"multiline": {s: `

        this
                is
            multi
line
        `},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := New(test.s)
			iorc, err := f.Fetch(context.Background())
			assert.NoError(t, err)

			bs, err := io.ReadAll(iorc)
			assert.NoError(t, err)
			assert.Equal(t, test.s, string(bs))
		})
	}
}

func TestStringFetcher_Close(t *testing.T) {
	t.Parallel()
	f := New(t.Name())
	iorc, err := f.Fetch(context.Background())
	assert.NoError(t, err)

	bs, err := io.ReadAll(iorc)
	assert.NoError(t, err)
	assert.Equal(t, t.Name(), string(bs))

	assert.NoError(t, f.Close())
}
