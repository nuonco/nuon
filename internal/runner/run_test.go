package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validJSON   = "testdata/valid_request.json"
	invalidJSON = "testdata/invalid.json"
)

func TestRunner_parseRequest(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		ior         func(t *testing.T) io.Reader
		expected    *Request
		errExpected error
	}{
		"happy path": {
			ior: func(t *testing.T) io.Reader {
				r, err := os.Open(validJSON)
				assert.NoError(t, err)
				t.Cleanup(func() { _ = r.Close() })
				return r
			},
			expected: &Request{
				ID:      "testid",
				RunType: RunTypeApply,
				Sandbox: Object{BucketName: "sandboxtest", Key: "prefix/key.tar.gz", Region: "us-east-2"},
				Backend: Object{BucketName: "backendtest", Key: "prefix/state.json", Region: "us-east-2"},
				Vars:    map[string]interface{}{},
			},
		},
		"invalid json": {
			ior: func(t *testing.T) io.Reader {
				r, err := os.Open(invalidJSON)
				assert.NoError(t, err)
				t.Cleanup(func() { _ = r.Close() })
				return r
			},
			errExpected: fmt.Errorf("invalid character ','"),
		},

		"empty json": {
			ior: func(t *testing.T) io.Reader {
				return strings.NewReader("")
			},
			errExpected: fmt.Errorf("unexpected end of JSON input"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := &runner{}
			req, err := r.parseRequest(test.ior(t))
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, req)
		})
	}
}

type mockModuleFetcher struct{ mock.Mock }

var _ moduleFetcher = (*mockModuleFetcher)(nil)

func (m *mockModuleFetcher) fetchModule(ctx context.Context) (io.ReadCloser, error) {
	args := m.Called(ctx)
	err := args.Error(1)
	if args.Get(0) != nil {
		return args.Get(0).(io.ReadCloser), err
	}
	return nil, err
}

type mockRequestParser struct{ mock.Mock }

var _ requestParser = (*mockRequestParser)(nil)

func (m *mockRequestParser) parseRequest(ior io.Reader) (*Request, error) {
	args := m.Called(ior)
	err := args.Error(1)
	if args.Get(0) != nil {
		return args.Get(0).(*Request), err
	}
	return nil, err
}

type mockWorkspaceSetuper struct{ mock.Mock }

var _ workspaceSetuper = (*mockWorkspaceSetuper)(nil)

func (m *mockWorkspaceSetuper) setupWorkspace(ctx context.Context, req *Request) (executor, error) {
	args := m.Called(ctx, req)
	err := args.Error(1)
	if args.Get(0) != nil {
		return args.Get(0).(executor), err
	}
	return nil, err
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := map[string]struct {
		mf          func(*testing.T) *mockModuleFetcher
		rp          func(*testing.T) *mockRequestParser
		ws          func(*testing.T) *mockWorkspaceSetuper
		expected    map[string]interface{}
		errExpected error
	}{
		"happiest path": {
			mf: func(t *testing.T) *mockModuleFetcher {
				// this is just a random io.ReadCloser
				iorc, err := os.Open(validJSON)
				assert.NoError(t, err)
				m := &mockModuleFetcher{}
				m.On("fetchModule", ctx).Return(iorc, nil)
				return m
			},
			rp: func(t *testing.T) *mockRequestParser {
				m := &mockRequestParser{}
				m.On("parseRequest", mock.AnythingOfType("*os.File")).Return(&Request{RunType: RunTypeApply}, nil)
				return m
			},
			ws: func(t *testing.T) *mockWorkspaceSetuper {
				me := &mockExecutor{}
				me.On("Init", ctx).Return(nil)
				me.On("Apply", ctx).Return(nil)
				me.On("Output", ctx).Return(map[string]interface{}{"got": "outputs"}, nil)
				m := &mockWorkspaceSetuper{}
				m.On("setupWorkspace", ctx, &Request{RunType: RunTypeApply}).Return(me, nil)
				return m
			},
			expected: map[string]interface{}{"got": "outputs"},
		},
		"failed fetching module": {
			mf: func(t *testing.T) *mockModuleFetcher {
				m := &mockModuleFetcher{}
				m.On("fetchModule", ctx).Return(nil, fmt.Errorf("failed fetching module"))
				return m
			},
			rp:          func(t *testing.T) *mockRequestParser { return &mockRequestParser{} },
			ws:          func(t *testing.T) *mockWorkspaceSetuper { return &mockWorkspaceSetuper{} },
			errExpected: fmt.Errorf("failed fetching module"),
		},
		"failed parsing request": {
			mf: func(t *testing.T) *mockModuleFetcher {
				// this is just a random io.ReadCloser
				iorc, err := os.Open(validJSON)
				assert.NoError(t, err)
				m := &mockModuleFetcher{}
				m.On("fetchModule", ctx).Return(iorc, nil)
				return m
			},
			rp: func(t *testing.T) *mockRequestParser {
				m := &mockRequestParser{}
				m.On("parseRequest", mock.AnythingOfType("*os.File")).Return(nil, fmt.Errorf("failed parsing request"))
				return m
			},
			ws:          func(t *testing.T) *mockWorkspaceSetuper { return &mockWorkspaceSetuper{} },
			errExpected: fmt.Errorf("failed parsing request"),
		},
		"failed setting up workspace": {
			mf: func(t *testing.T) *mockModuleFetcher {
				// this is just a random io.ReadCloser
				iorc, err := os.Open(validJSON)
				assert.NoError(t, err)
				m := &mockModuleFetcher{}
				m.On("fetchModule", ctx).Return(iorc, nil)
				return m
			},
			rp: func(t *testing.T) *mockRequestParser {
				m := &mockRequestParser{}
				m.On("parseRequest", mock.AnythingOfType("*os.File")).Return(&Request{RunType: RunTypeApply}, nil)
				return m
			},
			ws: func(t *testing.T) *mockWorkspaceSetuper {
				m := &mockWorkspaceSetuper{}
				m.On("setupWorkspace", ctx, &Request{RunType: RunTypeApply}).Return(nil, fmt.Errorf("failed setting up workspace"))
				return m
			},
			errExpected: fmt.Errorf("failed setting up workspace"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mf := test.mf(t)
			rp := test.rp(t)
			ws := test.ws(t)
			r := &runner{
				moduleFetcher:    mf,
				requestParser:    rp,
				workspaceSetuper: ws,
			}

			got, err := r.Run(ctx)
			mf.AssertExpectations(t)
			rp.AssertExpectations(t)
			ws.AssertExpectations(t)

			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

type mockExecutor struct{ mock.Mock }

func (m *mockExecutor) Init(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Plan(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Apply(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Destroy(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Output(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	err := args.Error(1)
	if m, ok := args.Get(0).(map[string]interface{}); ok {
		return m, err
	}
	return nil, err
}

func Test_run(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		executor    func(*testing.T) *mockExecutor
		runType     RunType
		expected    map[string]interface{}
		errExpected error
	}{
		"happy path - plan": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Plan", mock.Anything).Return(nil)
				return m
			},
			runType: RunTypePlan,
		},
		"happy path - apply": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(nil)
				m.On("Output", mock.Anything).Return(map[string]interface{}{}, nil)
				return m
			},
			runType:  RunTypeApply,
			expected: map[string]interface{}{},
		},
		"happy path - destroy": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Destroy", mock.Anything).Return(nil)
				return m
			},
			runType: RunTypeDestroy,
		},
		"returns outputs": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(nil)
				m.On("Output", mock.Anything).Return(map[string]interface{}{
					"a":    "bunch",
					"of":   "really",
					"cool": "outputs",
				}, nil)
				return m
			},
			runType:  RunTypeApply,
			expected: map[string]interface{}{"a": "bunch", "of": "really", "cool": "outputs"},
		},
		"invalid runtype": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				return m
			},
			runType:     "something made up",
			errExpected: fmt.Errorf("invalid run type"),
		},
		"init error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(fmt.Errorf("init error"))
				return m
			},
			runType:     RunTypeDestroy,
			errExpected: fmt.Errorf("init error"),
		},
		"plan error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Plan", mock.Anything).Return(fmt.Errorf("plan error"))
				return m
			},
			runType:     RunTypePlan,
			errExpected: fmt.Errorf("plan error"),
		},
		"destroy error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Destroy", mock.Anything).Return(fmt.Errorf("destroy error"))
				return m
			},
			runType:     RunTypeDestroy,
			errExpected: fmt.Errorf("destroy error"),
		},
		"apply error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(fmt.Errorf("apply error"))
				return m
			},
			runType:     RunTypeApply,
			errExpected: fmt.Errorf("apply error"),
		},
		"output error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(nil)
				m.On("Output", mock.Anything).Return(nil, fmt.Errorf("output error"))
				return m
			},
			runType:     RunTypeApply,
			errExpected: fmt.Errorf("output error"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			e := test.executor(t)

			got, err := run(context.Background(), e, test.runType)
			assert.Equal(t, test.expected, got)

			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			e.AssertExpectations(t)
		})
	}
}
