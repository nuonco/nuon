package vars

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewVarConfigurator(t *testing.T) {
	t.Parallel()
	v := validator.New()
	tests := map[string]struct {
		m           map[string]interface{}
		v           *validator.Validate
		errExpected error
	}{
		"valid": {
			m: map[string]interface{}{},
			v: v,
		},
		"missing vars": {
			m:           nil,
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'M' failed on the 'required' tag"),
		},
		"missing validator": {
			m:           map[string]interface{}{},
			v:           nil,
			errExpected: fmt.Errorf("validator is nil"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			vc, err := New(test.v, WithVars(test.m))
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, vc)
		})
	}
}

type mockWriter struct{ mock.Mock }

func (m *mockWriter) Write(bs []byte) (int, error) {
	args := m.Called(bs)
	return args.Int(0), args.Error(1)
}

func TestVarConfigurator_JSON(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		m           map[string]interface{}
		w           func(*testing.T) io.Writer
		errExpected error
	}{
		"happy path": {
			m: map[string]interface{}{"something": "cool"},
			w: func(t *testing.T) io.Writer {
				var b bytes.Buffer
				return &b
			},
		},
		"error on write": {
			m: map[string]interface{}{"something": "cool"},
			w: func(t *testing.T) io.Writer {
				m := &mockWriter{}
				m.On("Write", []byte(`{"something":"cool"}`)).Return(0, fmt.Errorf("error on write"))
				return m
			},
			errExpected: fmt.Errorf("error on write"),
		},
		"error marshalling": {
			m: map[string]interface{}{"something": func() {}},
			w: func(t *testing.T) io.Writer {
				var b bytes.Buffer
				return &b
			},
			errExpected: fmt.Errorf("json: unsupported type: func()"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := &varConfigurator{M: test.m}
			assert.NotNil(t, s)

			w := test.w(t)
			err := s.JSON(w)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			if w, ok := w.(*bytes.Buffer); ok {
				assert.Equal(t, `{"something":"cool"}`, w.String())
			}
		})
	}
}
