package token

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

	v := validator.New()
	tests := map[string]struct {
		v           *validator.Validate
		opts        []k8sTokenGetterOption
		errExpected error
	}{
		"happy path": {
			v:    v,
			opts: []k8sTokenGetterOption{WithNamespace("happy-path"), WithName("happy-path")},
		},
		"missing validator": {
			v:           nil,
			opts:        []k8sTokenGetterOption{WithNamespace("happy-path"), WithName("happy-path")},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing namespace": {
			v:           v,
			opts:        []k8sTokenGetterOption{WithName("happy-path")},
			errExpected: fmt.Errorf("Field validation for 'Namespace' failed on the 'required' tag"),
		},
		"missing name": {
			v:           v,
			opts:        []k8sTokenGetterOption{WithNamespace("happy-path")},
			errExpected: fmt.Errorf("Field validation for 'Name' failed on the 'required' tag"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			k, err := New(test.v, test.opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, k)
		})
	}
}

func TestK8sTokenGetter_GetOrgToken(t *testing.T) {
	t.Parallel()
	namespace := "default"
	secretName := uuid.NewString()

	tests := map[string]struct {
		kubeClientFn  func(*testing.T) kubeClientSecretGetter
		assertFn      func(*testing.T, kubeClientSecretGetter)
		errExpected   error
		expectedToken string
	}{
		"happy path": {
			expectedToken: "valid-test-token",
			kubeClientFn: func(t *testing.T) kubeClientSecretGetter {
				client := &testKubeClientSecretGetter{}
				client.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Secret{
					Data: map[string][]byte{
						"token": []byte("valid-test-token"),
					},
				}, nil)
				return client
			},
			assertFn: func(t *testing.T, client kubeClientSecretGetter) {
			},
		},
		"unable to get secret": {
			kubeClientFn: func(t *testing.T) kubeClientSecretGetter {
				connector := &testKubeClientSecretGetter{}
				connector.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil, errTokenTest)
				return connector
			},
			errExpected: errTokenTest,
		},
		"secret found no token": {
			kubeClientFn: func(t *testing.T) kubeClientSecretGetter {
				connector := &testKubeClientSecretGetter{}
				connector.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Secret{
					Data: map[string][]byte{},
				}, nil)
				return connector
			},
			errExpected: fmt.Errorf("token not found"),
		},
		"token is empty": {
			kubeClientFn: func(t *testing.T) kubeClientSecretGetter {
				connector := &testKubeClientSecretGetter{}
				connector.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Secret{
					Data: map[string][]byte{
						"token": []byte(nil),
					},
				}, nil)
				return connector
			},
			errExpected: fmt.Errorf("token was empty"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := test.kubeClientFn(t)
			tokenGetter := &k8sTokenGetter{
				Namespace: namespace,
				Name:      secretName,
				client:    client,
			}

			token, err := tokenGetter.GetOrgToken(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedToken, token)
			}

			if test.assertFn != nil {
				test.assertFn(t, client)
			}
		})
	}
}
