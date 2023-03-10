package backend

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewS3Configurator(t *testing.T) {
	t.Parallel()
	v := validator.New()
	tests := map[string]struct {
		c           *S3Config
		v           *validator.Validate
		errExpected error
	}{
		"valid": {
			c: &S3Config{
				BucketName:   "valid",
				BucketKey:    "valid",
				BucketRegion: "us-east-2",
			},
			v: v,
		},
		"missing config": {
			c:           nil,
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'S3BackendConfig' failed on the 'required' tag"),
		},
		"missing keys": {
			c: &S3Config{
				BucketName:   "",
				BucketKey:    "",
				BucketRegion: "",
			},
			v:           v,
			errExpected: fmt.Errorf("Error:Field validation"),
		},
		"missing validator": {
			c: &S3Config{
				BucketName:   "valid",
				BucketKey:    "valid",
				BucketRegion: "us-east-2",
			},
			v:           nil,
			errExpected: fmt.Errorf("validator is nil"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s, err := NewS3Configurator(test.v, WithBackendConfig(test.c))
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, s)
		})
	}
}

type mockWriter struct{ mock.Mock }

func (m *mockWriter) Write(bs []byte) (int, error) {
	args := m.Called(bs)
	return args.Int(0), args.Error(1)
}

func TestS3Configurator_JSON(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		c           *S3Config
		w           func(*testing.T) io.Writer
		errExpected error
	}{
		"happy path": {
			c: &S3Config{
				BucketName:   "valid",
				BucketKey:    "valid",
				BucketRegion: "us-east-2",
			},
			w: func(t *testing.T) io.Writer {
				var b bytes.Buffer
				return &b
			},
		},
		"error on write": {
			c: &S3Config{
				BucketName:   "error",
				BucketKey:    "error",
				BucketRegion: "us-east-2",
			},
			w: func(t *testing.T) io.Writer {
				m := &mockWriter{}
				m.On("Write", []byte(`{"bucket":"error","key":"error","region":"us-east-2"}`)).Return(0, fmt.Errorf("error on write"))
				return m
			},
			errExpected: fmt.Errorf("error on write"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := &s3Configurator{S3BackendConfig: test.c}
			assert.NotNil(t, s)

			w := test.w(t)
			err := s.JSON(w)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			if w, ok := w.(*bytes.Buffer); ok {
				assert.Equal(t, `{"bucket":"valid","key":"valid","region":"us-east-2"}`, w.String())
			}
		})
	}
}
