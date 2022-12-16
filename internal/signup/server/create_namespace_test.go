package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	corev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/rest"
)

type mockK8sNSClient func(context.Context, *coreapplyv1.NamespaceApplyConfiguration, apimetav1.ApplyOptions) (*corev1.Namespace, error)

func (m mockK8sNSClient) Apply(ctx context.Context, cfg *coreapplyv1.NamespaceApplyConfiguration, opts apimetav1.ApplyOptions) (*corev1.Namespace, error) {
	return m(ctx, cfg, opts)
}

func Test_createNamespace(t *testing.T) {
	tests := map[string]struct {
		api         func(*testing.T) mockK8sNSClient
		errExpected error
	}{
		"calls api": {
			api: func(t *testing.T) mockK8sNSClient {
				return func(ctx context.Context, nac *coreapplyv1.NamespaceApplyConfiguration, ao apimetav1.ApplyOptions) (*corev1.Namespace, error) {
					assert.Equal(t, t.Name(), *nac.Name)
					assert.NotNil(t, ao)
					return &corev1.Namespace{ObjectMeta: apimetav1.ObjectMeta{Name: t.Name()}}, nil
				}
			},
		},

		"wraps error": {
			api: func(t *testing.T) mockK8sNSClient {
				return func(ctx context.Context, nac *coreapplyv1.NamespaceApplyConfiguration, ao apimetav1.ApplyOptions) (*corev1.Namespace, error) {
					assert.Equal(t, t.Name(), *nac.Name)
					assert.NotNil(t, ao)
					return nil, errOops
				}
			},
			errExpected: errOops,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c := &nsCreator{}
			api := test.api(t)
			ns, err := c.createNamespace(context.Background(), api, t.Name())
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, ns)
		})
	}
}

type testNamespaceCreator struct {
	fn func(ctx context.Context, cna createNamespaceAPI, s string) (*corev1.Namespace, error)
}

func (t testNamespaceCreator) createNamespace(ctx context.Context, cna createNamespaceAPI, s string) (*corev1.Namespace, error) {
	return t.fn(ctx, cna, s)
}

func TestCreateNamespace(t *testing.T) {
	tests := map[string]struct {
		requestFn   func() CreateNamespaceRequest
		expected    CreateNamespaceResponse
		nsCreator   func(*testing.T) testNamespaceCreator
		errExpected error
	}{
		"errors if no namespace name": {
			requestFn: func() CreateNamespaceRequest {
				req := getFakeObj[CreateNamespaceRequest]()
				req.NamespaceName = ""
				return req
			},
			expected:    CreateNamespaceResponse{},
			errExpected: fmt.Errorf("NamespaceName"),
		},

		"wraps client error": {
			requestFn: func() CreateNamespaceRequest {
				req := getFakeObj[CreateNamespaceRequest]()
				req.NamespaceName = "test"
				return req
			},
			expected:    CreateNamespaceResponse{NamespaceName: "test"},
			errExpected: errOops,
			nsCreator: func(t *testing.T) testNamespaceCreator {
				return testNamespaceCreator{
					fn: func(ctx context.Context, cna createNamespaceAPI, s string) (*corev1.Namespace, error) {
						assert.NotNil(t, cna)
						assert.Equal(t, "test", s)
						return nil, errOops
					},
				}
			},
		},

		"does not error with valid request": {
			requestFn: func() CreateNamespaceRequest {
				req := getFakeObj[CreateNamespaceRequest]()
				req.NamespaceName = "test"
				return req
			},
			expected: CreateNamespaceResponse{NamespaceName: "test"},
			nsCreator: func(t *testing.T) testNamespaceCreator {
				return testNamespaceCreator{
					fn: func(ctx context.Context, cna createNamespaceAPI, s string) (*corev1.Namespace, error) {
						assert.NotNil(t, cna)
						assert.Equal(t, "test", s)

						return &corev1.Namespace{ObjectMeta: apimetav1.ObjectMeta{Name: "test"}}, nil
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

			if test.nsCreator != nil {
				a.namespaceCreator = test.nsCreator(t)
			}
			enc, err := env.ExecuteActivity(a.CreateNamespace, test.requestFn())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			resp := CreateNamespaceResponse{}
			err = enc.Get(&resp)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, resp)
		})
	}
}
