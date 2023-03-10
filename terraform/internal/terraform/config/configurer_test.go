package config

import (
	"fmt"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConfigurator struct {
	mock.Mock
}

func (m *mockConfigurator) JSON(w io.Writer) error {
	args := m.Called(w)
	return args.Error(0)
}

var _ configurator = (*mockConfigurator)(nil)

type mockWriteFactory struct {
	mock.Mock
}

func (m *mockWriteFactory) GetWriter(s string) (io.WriteCloser, error) {
	args := m.Called(s)
	err := args.Error(1)
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(io.WriteCloser), err
}

var _ writeFactory = (*mockWriteFactory)(nil)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		configs      []configurator
		writeFactory writeFactory
		v            *validator.Validate
		errExpected  error
	}{
		"valid": {
			configs:      []configurator{&mockConfigurator{}},
			writeFactory: &mockWriteFactory{},
			v:            v,
		},
		"multiple configs": {
			configs:      []configurator{&mockConfigurator{}, &mockConfigurator{}},
			writeFactory: &mockWriteFactory{},
			v:            v,
		},
		"missing config": {
			configs:      []configurator{},
			writeFactory: &mockWriteFactory{},
			v:            v,
			errExpected:  fmt.Errorf("failed on the 'gt' tag"),
		},
		"invalid config": {
			configs:      []configurator{nil},
			writeFactory: &mockWriteFactory{},
			v:            v,
			errExpected:  fmt.Errorf("failed on the 'required' tag"),
		},
		"missing writeFactory": {
			configs:      []configurator{&mockConfigurator{}},
			writeFactory: nil,
			v:            v,
			errExpected:  fmt.Errorf("Field validation for 'WriteFactory' failed on the 'required' tag"),
		},
		"missing validatory": {
			configs:      []configurator{&mockConfigurator{}},
			writeFactory: &mockWriteFactory{},
			v:            nil,
			errExpected:  fmt.Errorf("validator is nil"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var cfgs []configurerOptions
			for i, c := range test.configs {
				cfgs = append(cfgs, WithConfigurator(getName(t, i), c))
			}
			cfgs = append(cfgs, WithWriteFactory(test.writeFactory))

			c, err := New(test.v, cfgs...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, c)
		})
	}
}

type nopCloser struct{ io.Writer }

func (n *nopCloser) Close() error { return nil }

func TestConfigurer_Configure(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		configs      func(*testing.T) []*mockConfigurator
		writeFactory func(*testing.T) *mockWriteFactory
		errExpected  error
	}{
		"happy path": {
			configs: func(t *testing.T) []*mockConfigurator {
				m := &mockConfigurator{}
				m.
					On("JSON", mock.MatchedBy(func(io.Writer) bool { return true })).
					Return(nil)
				return []*mockConfigurator{m}
			},
			writeFactory: func(t *testing.T) *mockWriteFactory {
				m := &mockWriteFactory{}
				m.
					On("GetWriter", getName(t, 0)).
					Return(&nopCloser{io.Discard}, nil)
				return m
			},
		},
		"multiple configurators": {
			configs: func(t *testing.T) []*mockConfigurator {
				m := &mockConfigurator{}
				m.
					On("JSON", mock.MatchedBy(func(io.Writer) bool { return true })).
					Return(nil).
					Times(3)
				return []*mockConfigurator{m, m, m}
			},
			writeFactory: func(t *testing.T) *mockWriteFactory {
				m := &mockWriteFactory{}
				for i := 0; i < 3; i++ {
					m.
						On("GetWriter", getName(t, i)).
						Return(&nopCloser{io.Discard}, nil)
				}
				return m
			},
		},

		"nil writer": {
			configs: func(t *testing.T) []*mockConfigurator {
				m := &mockConfigurator{}
				return []*mockConfigurator{m}
			},
			writeFactory: func(t *testing.T) *mockWriteFactory {
				m := &mockWriteFactory{}
				m.
					On("GetWriter", getName(t, 0)).
					Return(nil, nil)
				return m
			},
			errExpected: fmt.Errorf("nil writer"),
		},

		"error getting writer": {
			configs: func(t *testing.T) []*mockConfigurator {
				m := &mockConfigurator{}
				return []*mockConfigurator{m}
			},
			writeFactory: func(t *testing.T) *mockWriteFactory {
				m := &mockWriteFactory{}
				m.
					On("GetWriter", getName(t, 0)).
					Return(nil, fmt.Errorf("write factory error"))
				return m
			},
			errExpected: fmt.Errorf("write factory error"),
		},
		"error writing config": {
			configs: func(t *testing.T) []*mockConfigurator {
				m := &mockConfigurator{}
				m.
					On("JSON", mock.MatchedBy(func(io.Writer) bool { return true })).
					Return(fmt.Errorf("error writing config"))
				return []*mockConfigurator{m}
			},
			writeFactory: func(t *testing.T) *mockWriteFactory {
				m := &mockWriteFactory{}
				m.
					On("GetWriter", getName(t, 0)).
					Return(&nopCloser{io.Discard}, nil)
				return m
			},
			errExpected: fmt.Errorf("error writing config"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrs := test.configs(t)
			wf := test.writeFactory(t)

			var cfgs []configurerOptions
			for i, c := range ctrs {
				cfgs = append(cfgs, WithConfigurator(fmt.Sprintf("%s-%d", t.Name(), i), c))
			}
			cfgs = append(cfgs, WithWriteFactory(wf))

			c, err := New(v, cfgs...)
			assert.NoError(t, err)

			err = c.Configure()
			for _, ctr := range ctrs {
				ctr.AssertExpectations(t)
			}
			wf.AssertExpectations(t)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, c)
		})
	}
}

func getName(t *testing.T, i int) string {
	return fmt.Sprintf("%s-%d", t.Name(), i)
}
