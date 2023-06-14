package client

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/generics"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/stretchr/testify/assert"
	tclient "go.temporal.io/sdk/client"
)

func Test_TriggerCanaryProvision(t *testing.T) {
	errProvision := fmt.Errorf("error provision")
	req := generics.GetFakeObj[*canaryv1.ProvisionRequest]()
	tests := map[string]struct {
		tclient     func(*testing.T, *gomock.Controller) temporal.Client
		errExpected error
	}{
		"happy path": {
			tclient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "canary", gomock.Any(), "Provision", gomock.Any()).Return(nil, nil)
				return client
			},
		},
		"error": {
			tclient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
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

			tclient := test.tclient(t, mockCtl)
			wfClient := &workflowsClient{
				TemporalClient: tclient,
			}

			err := wfClient.TriggerCanaryProvision(ctx, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_TriggerCanaryDeprovision(t *testing.T) {
	errDeprovision := fmt.Errorf("error deprovision")
	req := generics.GetFakeObj[*canaryv1.DeprovisionRequest]()
	tests := map[string]struct {
		tclient     func(*testing.T, *gomock.Controller) temporal.Client
		errExpected error
	}{
		"happy path": {
			tclient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "canary", gomock.Any(), "Deprovision", gomock.Any()).Return(nil, nil)
				return client
			},
		},
		"error": {
			tclient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(), "canary", gomock.Any(), "Deprovision", req).Return(nil, errDeprovision)
				return client
			},
			errExpected: errDeprovision,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			tclient := test.tclient(t, mockCtl)
			wfClient := &workflowsClient{
				TemporalClient: tclient,
			}

			err := wfClient.TriggerCanaryDeprovision(ctx, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_ScheduleCanaryProvision(t *testing.T) {
	errScheduleCanaryProvision := fmt.Errorf("error scheduling")

	req := generics.GetFakeObj[*canaryv1.ProvisionRequest]()
	schedule := "* * * * *"
	workflowID := generics.GetFakeObj[string]()
	tests := map[string]struct {
		newTClient  func(*testing.T, *gomock.Controller) temporal.Client
		errExpected error
	}{
		"happy path": {
			newTClient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				expectedOpts := tclient.StartWorkflowOptions{
					ID:           workflowID,
					CronSchedule: schedule,
					TaskQueue:    DefaultTaskQueue,
					Memo: map[string]interface{}{
						"canary-id":  req.CanaryId,
						"started-by": defaultAgent,
					},
				}
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(),
					"canary",
					expectedOpts,
					"Provision",
					req).Return(nil, nil)
				return client
			},
		},
		"error path": {
			newTClient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				client := temporal.NewMockClient(mockCtl)
				client.EXPECT().ExecuteWorkflowInNamespace(gomock.Any(),
					"canary",
					gomock.Any(),
					"Provision",
					req).Return(nil, errScheduleCanaryProvision)
				return client
			},
			errExpected: errScheduleCanaryProvision,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			client := test.newTClient(t, mockCtl)
			wfClient := &workflowsClient{
				TemporalClient: client,
				Agent:          defaultAgent,
			}

			err := wfClient.ScheduleCanaryProvision(ctx, workflowID, schedule, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_TriggerCanaryUnschedule(t *testing.T) {
	errCancel := fmt.Errorf("error canceling")
	workflowID := generics.GetFakeObj[string]()

	tests := map[string]struct {
		newTClient  func(*testing.T, *gomock.Controller) temporal.Client
		errExpected error
	}{
		"happy path": {
			newTClient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				tclient := temporal.NewMockClient(mockCtl)
				tclient.EXPECT().CancelWorkflowInNamespace(gomock.Any(), "canary", workflowID, "").Return(nil)
				return tclient
			},
		},
		"error": {
			newTClient: func(t *testing.T, mockCtl *gomock.Controller) temporal.Client {
				tclient := temporal.NewMockClient(mockCtl)
				tclient.EXPECT().CancelWorkflowInNamespace(gomock.Any(), "canary", workflowID, "").Return(errCancel)
				return tclient
			},
			errExpected: errCancel,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			tclient := test.newTClient(t, mockCtl)
			wfClient := &workflowsClient{
				TemporalClient: tclient,
				Agent:          defaultAgent,
			}

			err := wfClient.UnscheduleCanaryProvision(ctx, workflowID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
