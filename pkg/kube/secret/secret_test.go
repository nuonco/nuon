package secret

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
)

var errTokenTest = fmt.Errorf("token-test-err")

type testKubeClientSecretGetter struct {
	mock.Mock
}

func (t *testKubeClientSecretGetter) Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Secret, error) {
	args := t.Called(ctx, name, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*corev1.Secret), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestNew(t *testing.T) {
	t.Parallel()

	name := uuid.NewString()
	key := uuid.NewString()
	namespace := uuid.NewString()
	clusterInfo := generics.GetFakeObj[*kube.ClusterInfo]()

	v := validator.New()
	tests := map[string]struct {
		optsFn      func() []k8sSecretManagerOption
		assertFn    func(*testing.T, *k8sSecretManager)
		errExpected error
	}{
		"happy path": {
			optsFn: func() []k8sSecretManagerOption {
				return []k8sSecretManagerOption{
					WithKey(key),
					WithName(name),
					WithNamespace(namespace),
				}
			},
			assertFn: func(t *testing.T, sg *k8sSecretManager) {
				assert.Equal(t, name, sg.Name)
				assert.Equal(t, namespace, sg.Namespace)
				assert.Equal(t, key, sg.Key)
			},
		},
		"custom cluster info": {
			optsFn: func() []k8sSecretManagerOption {
				return []k8sSecretManagerOption{
					WithKey(key),
					WithName(name),
					WithNamespace(namespace),
					WithCluster(clusterInfo),
				}
			},
			assertFn: func(t *testing.T, sg *k8sSecretManager) {
				assert.Equal(t, name, sg.Name)
				assert.Equal(t, namespace, sg.Namespace)
				assert.Equal(t, key, sg.Key)
				assert.Equal(t, clusterInfo, sg.ClusterInfo)
			},
		},
		"missing namespace": {
			optsFn: func() []k8sSecretManagerOption {
				return []k8sSecretManagerOption{
					WithKey(key),
					WithName(name),
				}
			},
			errExpected: fmt.Errorf("Namespace"),
		},
		"missing name": {
			optsFn: func() []k8sSecretManagerOption {
				return []k8sSecretManagerOption{
					WithKey(key),
					WithNamespace(namespace),
				}
			},
			errExpected: fmt.Errorf("Name"),
		},
		"missing key": {
			optsFn: func() []k8sSecretManagerOption {
				return []k8sSecretManagerOption{
					WithName(name),
					WithNamespace(namespace),
				}
			},
			errExpected: fmt.Errorf("Key"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			k, err := New(v, test.optsFn()...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, k)
		})
	}
}
