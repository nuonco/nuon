package deprovision

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_finisherImpl_sendSuccessNotification(t *testing.T) {
	errUnableToSend := fmt.Errorf("unableToSend")
	req := FinishRequest{
		DeprovisionRequest:  getFakeDeprovisionRequest(),
		InstallationsBucket: "nuon-installations-stage",
		Success:             true,
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
				assert.Contains(t, notif, "success")
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
			s := &finisherImpl{}
			sender := test.senderFn()

			err := s.sendSuccessNotification(context.Background(), req, sender)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(sender)
		})
	}
}

func Test_finisherImpl_sendErrorNotification(t *testing.T) {
	errUnableToSend := fmt.Errorf("unableToSend")
	req := FinishRequest{
		DeprovisionRequest:  getFakeDeprovisionRequest(),
		InstallationsBucket: "nuon-installations-stage",
		Success:             true,
		ErrorStep:           "destroy_step",
		ErrorMessage:        "failed to destroy",
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
				assert.Contains(t, notif, "error")
				assert.Contains(t, notif, req.ErrorMessage)
				assert.Contains(t, notif, req.ErrorStep)
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
			s := &finisherImpl{}
			sender := test.senderFn()

			err := s.sendErrorNotification(context.Background(), req, sender)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(sender)
		})
	}
}
