package error

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	e := fmt.Errorf("an error")
	ef := New(e)
	assert.Equal(t, e, ef.error)
}

func TestErrorFetcher_Fetch(t *testing.T) {
	e := fmt.Errorf("an error")
	ef := New(e)
	assert.Equal(t, e, ef.error)

	iorc, err := ef.Fetch(context.Background())
	assert.ErrorContains(t, err, e.Error())
	assert.Nil(t, iorc)
}

func TestErrorFetcher_Close(t *testing.T) {
	e := fmt.Errorf("an error")
	ef := New(e)
	assert.Equal(t, e, ef.error)

	assert.ErrorContains(t, ef.Close(), e.Error())
}
