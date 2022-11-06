package server

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
	corev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/rest"
)

type mockK8sSvcClient func(context.Context, *coreapplyv1.ServiceApplyConfiguration, apimetav1.ApplyOptions) (*corev1.Service, error)

func (m mockK8sSvcClient) Apply(ctx context.Context, cfg *coreapplyv1.ServiceApplyConfiguration, opts apimetav1.ApplyOptions) (*corev1.Service, error) {
	return m(ctx, cfg, opts)
}

func Test_createService(t *testing.T) {
	tests := map[string]struct {
		api         func(*testing.T) mockK8sSvcClient
		errExpected error
	}{
		"calls api": {
			api: func(t *testing.T) mockK8sSvcClient {
				return func(ctx context.Context, sac *coreapplyv1.ServiceApplyConfiguration, ao apimetav1.ApplyOptions) (*corev1.Service, error) {
					assert.NotNil(t, ao)
					return &corev1.Service{ObjectMeta: apimetav1.ObjectMeta{Name: t.Name()}}, nil
				}
			},
		},

		"wraps error": {
			api: func(t *testing.T) mockK8sSvcClient {
				return func(ctx context.Context, nac *coreapplyv1.ServiceApplyConfiguration, ao apimetav1.ApplyOptions) (*corev1.Service, error) {
					return nil, errOops
				}
			},
			errExpected: errOops,
		},
		"properly configures service": {
			api: func(t *testing.T) mockK8sSvcClient {
				return func(ctx context.Context, svc *coreapplyv1.ServiceApplyConfiguration, ao apimetav1.ApplyOptions) (*corev1.Service, error) {
					assert.Equal(t, "wp-"+t.Name()+"-waypoint-server-public", *svc.Name)
					assert.Equal(t, t.Name(), *svc.Namespace)

					// NOTE(jm): there's not a great way to test the actual schema itself without
					// copy/pasting all the different fields here. This covers some basics but not
					// all the fields
					assert.NotEmpty(t, svc.Spec)
					assert.Equal(t, 1, len(svc.Spec.Ports))
					assert.NotEmpty(t, svc.Annotations)
					assert.NotEmpty(t, svc.Labels)
					return nil, nil
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c := &svcCreator{}
			k8sAPI := test.api(t)
			req := ExposeWaypointServerRequest{
				NamespaceName: t.Name(),
				ShortID:       t.Name(),
				RootDomain:    "test.nuon.co",
			}
			_, err := c.createService(context.Background(), k8sAPI, req)
			if test.errExpected != nil {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

type testServiceCreator struct {
	fn func(ctx context.Context, api k8sServiceCreator, req ExposeWaypointServerRequest) (*corev1.Service, error)
}

func (t testServiceCreator) createService(ctx context.Context, api k8sServiceCreator, req ExposeWaypointServerRequest) (*corev1.Service, error) {
	return t.fn(ctx, api, req)
}

func getFakeExposeWaypointServerRequest() ExposeWaypointServerRequest {
	fkr := faker.New()
	var req ExposeWaypointServerRequest
	fkr.Struct().Fill(&req)
	return req
}

func TestCreateService(t *testing.T) {
	tests := map[string]struct {
		requestFn      func() ExposeWaypointServerRequest
		svcCreator     func(*testing.T) testServiceCreator
		errExpectedMsg string
	}{
		"errors if no root domain": {
			requestFn: func() ExposeWaypointServerRequest {
				req := getFakeExposeWaypointServerRequest()
				req.RootDomain = ""
				return req
			},
			errExpectedMsg: "RootDomain",
		},
		"errors if no Short ID": {
			requestFn: func() ExposeWaypointServerRequest {
				req := getFakeExposeWaypointServerRequest()
				req.ShortID = ""
				return req
			},
			errExpectedMsg: "ShortID",
		},
		"errors if no namespace name": {
			requestFn: func() ExposeWaypointServerRequest {
				req := getFakeExposeWaypointServerRequest()
				req.NamespaceName = ""
				return req
			},
			errExpectedMsg: "Namespace",
		},

		"wraps client error": {
			requestFn: func() ExposeWaypointServerRequest {
				return getFakeExposeWaypointServerRequest()
			},
			errExpectedMsg: errOops.Error(),
			svcCreator: func(t *testing.T) testServiceCreator {
				return testServiceCreator{
					fn: func(ctx context.Context, csa k8sServiceCreator, req ExposeWaypointServerRequest) (*corev1.Service, error) {
						assert.NotNil(t, csa)
						return nil, errOops
					},
				}
			},
		},

		"does not error with valid request": {
			requestFn: func() ExposeWaypointServerRequest {
				return getFakeExposeWaypointServerRequest()
			},
			svcCreator: func(t *testing.T) testServiceCreator {
				return testServiceCreator{
					fn: func(ctx context.Context, csa k8sServiceCreator, req ExposeWaypointServerRequest) (*corev1.Service, error) {
						assert.NotNil(t, csa)
						return &corev1.Service{ObjectMeta: apimetav1.ObjectMeta{Name: "test"}}, nil
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

			if test.svcCreator != nil {
				a.serviceCreator = test.svcCreator(t)
			}
			_, err := env.ExecuteActivity(a.ExposeWaypointServer, test.requestFn())
			if test.errExpectedMsg != "" {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, test.errExpectedMsg)
				return
			}
			assert.NoError(t, err)
		})
	}
}
