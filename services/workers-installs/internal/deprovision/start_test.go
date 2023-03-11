package deprovision

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testNotificationSender struct {
	mock.Mock
}

func (t *testNotificationSender) Send(ctx context.Context, notif string) error {
	resp := t.Called(ctx, notif)
	return resp.Error(0)
}

func Test_starterImpl_sendStartNotification(t *testing.T) {
	errUnableToSend := fmt.Errorf("unableToSend")
	req := StartRequest{
		DeprovisionRequest:  generics.GetFakeObj[*installsv1.DeprovisionRequest](),
		InstallationsBucket: "nuon-installations-stage",
	}
	assert.Nil(t, req.validate())

	tests := map[string]struct {
		senderFn    func() notificationSender
		assertFn    func(notificationSender)
		errExpected error
	}{
		"happy path": {
			senderFn: func() notificationSender {
				s := &testNotificationSender{}
				s.On("Send", mock.Anything, mock.Anything).Return(nil)
				return s
			},
			assertFn: func(sender notificationSender) {
				obj := sender.(*testNotificationSender)
				obj.AssertNumberOfCalls(t, "Send", 1)
				notif := obj.Calls[0].Arguments[1].(string)
				assert.NotEmpty(t, notif)
			},
			errExpected: nil,
		},
		"error": {
			senderFn: func() notificationSender {
				s := &testNotificationSender{}
				s.On("Send", mock.Anything, mock.Anything).Return(errUnableToSend)
				return s
			},
			assertFn: func(sender notificationSender) {
				obj := sender.(*testNotificationSender)
				obj.AssertNumberOfCalls(t, "Send", 1)
				notif := obj.Calls[0].Arguments[1].(string)
				assert.NotEmpty(t, notif)
			},
			errExpected: errUnableToSend,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := &starterImpl{}
			sender := test.senderFn()

			err := s.sendStartNotification(context.Background(), req, sender)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(sender)
		})
	}
}
