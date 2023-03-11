package provision

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/sender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type testWaypointClientHostnameGetter struct {
	mock.Mock
}

func (t *testWaypointClientHostnameGetter) ListHostnames(ctx context.Context, req *gen.ListHostnamesRequest, opts ...grpc.CallOption) (*gen.ListHostnamesResponse, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.ListHostnamesResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

var _ waypointClientHostnameGetter = (*testWaypointClientHostnameGetter)(nil)

func Test_hostnameNotificationSenderImpl_getHostname(t *testing.T) {
	testErr := fmt.Errorf("test-error")

	req := generics.GetFakeObj[SendHostnameNotificationRequest]()

	tests := map[string]struct {
		clientFn    func() waypointClientHostnameGetter
		assertFn    func(*testing.T, waypointClientHostnameGetter, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func() waypointClientHostnameGetter {
				client := &testWaypointClientHostnameGetter{}
				resp := &gen.ListHostnamesResponse{
					Hostnames: []*gen.Hostname{
						{
							Fqdn: "https://initially-central-monkey.waypoint.run",
						},
					},
				}

				client.On("ListHostnames", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientHostnameGetter, hostname string) {
				assert.Equal(t, hostname, "https://initially-central-monkey.waypoint.run")

				obj := client.(*testWaypointClientHostnameGetter)
				obj.AssertNumberOfCalls(t, "ListHostnames", 1)
				wpReq := obj.Calls[0].Arguments[1].(*gen.ListHostnamesRequest)

				target := wpReq.Target.Target.(*gen.Hostname_Target_Application)
				assert.Equal(t, req.ComponentID, target.Application.Application.Application)
				assert.Equal(t, req.InstallID, target.Application.Application.Project)
				assert.Equal(t, req.InstallID, target.Application.Workspace.Workspace)
			},
		},
		"no hostnames found": {
			clientFn: func() waypointClientHostnameGetter {
				client := &testWaypointClientHostnameGetter{}
				client.On("ListHostnames", mock.Anything, mock.Anything, mock.Anything).Return(&gen.ListHostnamesResponse{}, nil)
				return client
			},
			errExpected: errNoHostnamesFound,
		},
		"error": {
			clientFn: func() waypointClientHostnameGetter {
				client := &testWaypointClientHostnameGetter{}
				client.On("ListHostnames", mock.Anything, mock.Anything, mock.Anything).Return(nil, testErr)
				return client
			},
			errExpected: testErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			impl := hostnameNotificationSenderImpl{}
			client := test.clientFn()

			hostname, err := impl.getHostname(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client, hostname)
		})
	}
}

type testNotificationSender struct {
	mock.Mock
}

func (t *testNotificationSender) Send(ctx context.Context, msg string) error {
	args := t.Called(ctx, msg)
	return args.Error(0)
}

var _ sender.NotificationSender = (*testNotificationSender)(nil)

func Test_hostnameNotificationSenderImpl_sendHostnameNotification(t *testing.T) {
	testErr := fmt.Errorf("test-error")

	req := generics.GetFakeObj[SendHostnameNotificationRequest]()
	hostname := "https://initially-central-monkey.waypoint.run"

	tests := map[string]struct {
		clientFn    func() sender.NotificationSender
		assertFn    func(*testing.T, sender.NotificationSender)
		errExpected error
	}{
		"happy path": {
			clientFn: func() sender.NotificationSender {
				client := &testNotificationSender{}
				client.On("Send", mock.Anything, mock.Anything).Return(nil)
				return client
			},
			assertFn: func(t *testing.T, client sender.NotificationSender) {
				obj := client.(*testNotificationSender)
				obj.AssertNumberOfCalls(t, "Send", 1)
				msg := obj.Calls[0].Arguments[1].(string)

				assert.Contains(t, msg, req.InstallID)
				assert.Contains(t, msg, req.OrgID)
				assert.Contains(t, msg, hostname)
			},
		},
		"error": {
			clientFn: func() sender.NotificationSender {
				client := &testNotificationSender{}
				client.On("Send", mock.Anything, mock.Anything).Return(testErr)
				return client
			},
			errExpected: testErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			impl := hostnameNotificationSenderImpl{
				sender: client,
			}

			err := impl.sendHostnameNotification(context.Background(), hostname, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
