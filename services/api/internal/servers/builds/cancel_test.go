package builds

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	temporal "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/stretchr/testify/assert"
)

func TestCancelBuild(t *testing.T) {
	errCancelBuild := fmt.Errorf("cancel build failed")
	tests := map[string]struct {
		clientFn    func(*gomock.Controller) *temporal.MockClient
		errExpected error
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) *temporal.MockClient {
				mock := temporal.NewMockClient(mockCtl)
				mock.EXPECT().CancelWorkflowInNamespace(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, namespace string, workflowID string, runID string) error {
						assert.Equal(t, "builds", namespace)
						assert.Equal(t, "is anything", workflowID)
						return nil
					})
				return mock
			},
		},
		"error": {
			errExpected: errCancelBuild,
			clientFn: func(mockCtl *gomock.Controller) *temporal.MockClient {
				mock := temporal.NewMockClient(mockCtl)
				mock.EXPECT().CancelWorkflowInNamespace(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errCancelBuild)
				return mock
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)
			client := test.clientFn(mockCtl)

			srv := &server{
				v:              validator.New(),
				temporalClient: client,
			}

			err := srv.cancelWorkflow(ctx, gomock.Any().String())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
