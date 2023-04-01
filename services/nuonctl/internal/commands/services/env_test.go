package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_commands_getServiceEnv(t *testing.T) {
	svcName := uuid.NewString()
	errGetEnv := fmt.Errorf("unable to get environment")
	cfgMap := &corev1.ConfigMap{
		Data: map[string]string{
			"KEY": "VALUE",
		},
	}

	tests := map[string]struct {
		kubeClient  func(*gomock.Controller) k8sConfigMapGetter
		errExpected error
		assertFn    func(*testing.T, map[string]string)
	}{
		"happy path": {
			kubeClient: func(mockCtl *gomock.Controller) k8sConfigMapGetter {
				client := NewMockk8sConfigMapGetter(mockCtl)

				client.EXPECT().Get(gomock.Any(), svcName, metav1.GetOptions{}).
					Return(cfgMap, nil)

				return client
			},
			assertFn: func(t *testing.T, env map[string]string) {
				assert.Equal(t, env, cfgMap.Data)
			},
		},
		"error": {
			kubeClient: func(mockCtl *gomock.Controller) k8sConfigMapGetter {
				client := NewMockk8sConfigMapGetter(mockCtl)

				client.EXPECT().Get(gomock.Any(), svcName, metav1.GetOptions{}).
					Return(nil, errGetEnv)

				return client

			},
			errExpected: errGetEnv,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			kubeClient := test.kubeClient(mockCtl)

			cmds := &commands{}
			env, err := cmds.getServiceEnv(ctx, kubeClient, svcName)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, env)
		})
	}
}
