package secret

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

func TestK8sTokenGetter_Get(t *testing.T) {
	t.Parallel()
	namespace := "default"
	name := uuid.NewString()
	key := uuid.NewString()

	tests := map[string]struct {
		kubeClientFn func(*testing.T) kubeClientSecretGetter
		assertFn     func(*testing.T, kubeClientSecretGetter)
		errExpected  error
		expectedData []byte
	}{
		"happy path": {
			expectedData: []byte("valid"),
			kubeClientFn: func(t *testing.T) kubeClientSecretGetter {
				client := &testKubeClientSecretGetter{}
				client.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Secret{
					Data: map[string][]byte{
						key: []byte("valid"),
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
		"secret found no key": {
			kubeClientFn: func(t *testing.T) kubeClientSecretGetter {
				connector := &testKubeClientSecretGetter{}
				connector.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Secret{
					Data: map[string][]byte{},
				}, nil)
				return connector
			},
			errExpected: fmt.Errorf("key not found"),
		},
		"token is empty": {
			kubeClientFn: func(t *testing.T) kubeClientSecretGetter {
				connector := &testKubeClientSecretGetter{}
				connector.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Secret{
					Data: map[string][]byte{
						key: []byte(nil),
					},
				}, nil)
				return connector
			},
			errExpected: fmt.Errorf("key was empty"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := test.kubeClientFn(t)
			secretGetter := &k8sSecretGetter{
				Namespace: namespace,
				Name:      name,
				Key:       key,
				client:    client,
			}

			token, err := secretGetter.Get(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedData, token)
			}

			if test.assertFn != nil {
				test.assertFn(t, client)
			}
		})
	}
}
