package temporal

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	"github.com/powertoolsdev/mono/pkg/generics"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/stretchr/testify/assert"
)

func Test_repo_TriggerCanaryProvision(t *testing.T) {
	errProvision := fmt.Errorf("error provision")
	req := generics.GetFakeObj[*canaryv1.ProvisionRequest]()
	tests := map[string]struct {
		client      func(*testing.T, *gomock.Controller) temporal.Client
		errExpected error
	}{
		"happy path": {
			client: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "canary", gomock.Any(), "Provision", gomock.Any()).Return(nil, nil)
				return client
			},
		},
		"error": {
			client: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "canary", gomock.Any(), "Provision", req).Return(nil, errProvision)
				return client
			},
			errExpected: errProvision,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			client := test.client(t, mockCtl)
			repo := &repo{
				Client: client,
			}

			err := repo.TriggerCanaryProvision(ctx, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
