package deprovision

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var errOops error = errors.New("oops")

type mockK8sNSClient func(context.Context, string, apimetav1.DeleteOptions) error

func (m mockK8sNSClient) Delete(ctx context.Context, name string, opts apimetav1.DeleteOptions) error {
	return m(ctx, name, opts)
}

func Test_destroyNamespace(t *testing.T) {
	tests := map[string]struct {
		api         func(*testing.T) mockK8sNSClient
		errExpected error
	}{
		"calls api": {
			api: func(t *testing.T) mockK8sNSClient {
				return func(ctx context.Context, name string, do apimetav1.DeleteOptions) error {
					assert.Equal(t, t.Name(), name)
					assert.NotNil(t, do)
					return nil
				}
			},
		},

		"wraps error": {
			api: func(t *testing.T) mockK8sNSClient {
				return func(ctx context.Context, name string, do apimetav1.DeleteOptions) error {
					assert.Equal(t, t.Name(), name)
					assert.NotNil(t, do)
					return errOops
				}
			},
			errExpected: errOops,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c := &nsDestroyer{}
			api := test.api(t)
			err := c.destroyNamespace(context.Background(), api, t.Name())
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)
		})
	}
}

type testNamespaceDestroyer struct {
	fn func(ctx context.Context, dna destroyNamespaceAPI, s string) error
}

func (t testNamespaceDestroyer) destroyNamespace(ctx context.Context, dna destroyNamespaceAPI, s string) error {
	return t.fn(ctx, dna, s)
}

func TestDestroyNamespace(t *testing.T) {
	tests := map[string]struct {
		request     DestroyNamespaceRequest
		expected    DestroyNamespaceResponse
		nsDestroyer func(*testing.T) testNamespaceDestroyer
		errExpected error
	}{
		"wraps client error": {
			request: DestroyNamespaceRequest{
				NamespaceName: "test",
			},
			expected:    DestroyNamespaceResponse{},
			errExpected: errOops,
			nsDestroyer: func(t *testing.T) testNamespaceDestroyer {
				return testNamespaceDestroyer{
					fn: func(ctx context.Context, dna destroyNamespaceAPI, s string) error {
						assert.NotNil(t, dna)
						assert.Equal(t, "test", s)
						return errOops
					},
				}
			},
		},

		"does not error with valid request": {
			request: DestroyNamespaceRequest{
				NamespaceName: "test",
			},
			expected: DestroyNamespaceResponse{},
			nsDestroyer: func(t *testing.T) testNamespaceDestroyer {
				return testNamespaceDestroyer{
					fn: func(ctx context.Context, dna destroyNamespaceAPI, s string) error {
						assert.NotNil(t, dna)
						assert.Equal(t, "test", s)
						return nil
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()
			a := &Activities{}
			a.Kubeconfig = &rest.Config{}
			env.RegisterActivity(a)

			if test.nsDestroyer != nil {
				a.namespaceDestroyer = test.nsDestroyer(t)
			}
			enc, err := env.ExecuteActivity(a.DestroyNamespace, test.request)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			resp := DestroyNamespaceResponse{}
			err = enc.Get(&resp)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, resp)
		})
	}
}
