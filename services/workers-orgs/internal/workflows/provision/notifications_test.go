package provision

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/slices"
)

func getFakeSendNotificationRequest() SendNotificationRequest {
	var obj SendNotificationRequest
	err := faker.FakeData(&obj)
	if err != nil {
		log.Fatalf("unable to create fake obj: %s", err)
	}
	return obj
}

type mockSender struct {
	mock.Mock
}

func (m *mockSender) Send(ctx context.Context, msg string) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

var _ NotificationSender = (*mockSender)(nil)

func Test_sendStartNotification(t *testing.T) {
	tests := map[string]struct {
		fn          func(*testing.T, func(string) bool) NotificationSender
		req         func() SendNotificationRequest
		errExpected error
	}{
		"happy path": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Started = true
				return req
			},
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(nil).Once()

				return ms
			},
		},

		"error on send": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Started = true
				return req
			},
			errExpected: fmt.Errorf("send error"),
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(fmt.Errorf("send error")).Once()

				return ms
			},
		},

		"error without sender": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Started = true
				return req
			},
			errExpected: errNoValidSender,
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				return nil
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.req()

			matcher := func(s string) bool {
				var accum []bool

				accum = append(accum, assert.Contains(t, s, req.ID))
				accum = append(accum, assert.Contains(t, s, "started provisioning a new org"))
				return !slices.Contains(accum, false)
			}

			s := test.fn(t, matcher)
			n := &notifierImpl{sender: s}

			err := n.sendStartNotification(context.Background(), req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			if s, ok := s.(*mockSender); ok {
				s.AssertExpectations(t)
			}
		})
	}
}

func Test_sendSuccessNotification(t *testing.T) {
	tests := map[string]struct {
		fn          func(*testing.T, func(string) bool) NotificationSender
		req         func() SendNotificationRequest
		errExpected error
	}{
		"happy path": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Finished = true
				return req
			},
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(nil).Once()

				return ms
			},
		},

		"error on send": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Finished = true
				return req
			},
			errExpected: fmt.Errorf("send error"),
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(fmt.Errorf("send error")).Once()

				return ms
			},
		},

		"error without sender": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Finished = true
				return req
			},
			errExpected: errNoValidSender,
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				return nil
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.req()
			matcher := func(s string) bool {
				var accum []bool
				accum = append(accum, assert.Contains(t, s, req.ID))
				accum = append(accum, assert.Contains(t, s, "successfully provisioned org"))
				return !slices.Contains(accum, false)
			}

			s := test.fn(t, matcher)
			n := &notifierImpl{sender: s}

			err := n.sendSuccessNotification(context.Background(), req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			if s, ok := s.(*mockSender); ok {
				s.AssertExpectations(t)
			}
		})
	}
}

func Test_sendErrorNotification(t *testing.T) {
	tests := map[string]struct {
		fn          func(*testing.T, func(string) bool) NotificationSender
		req         func() SendNotificationRequest
		errExpected error
	}{
		"happy path": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Erred = true
				return req
			},
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(nil).Once()

				return ms
			},
		},

		"error on send": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Erred = true
				return req
			},
			errExpected: fmt.Errorf("send error"),
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(fmt.Errorf("send error")).Once()

				return ms
			},
		},

		"error without sender": {
			req: func() SendNotificationRequest {
				req := getFakeSendNotificationRequest()
				req.Erred = true
				return req
			},
			errExpected: errNoValidSender,
			fn: func(t *testing.T, matcher func(string) bool) NotificationSender {
				return nil
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.req()
			matcher := func(s string) bool {
				var accum []bool
				accum = append(accum, assert.Contains(t, s, "error occurred provisioning org"))
				return !slices.Contains(accum, false)
			}

			s := test.fn(t, matcher)
			n := &notifierImpl{sender: s}

			err := n.sendErrorNotification(context.Background(), req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			if s, ok := s.(*mockSender); ok {
				s.AssertExpectations(t)
			}
		})
	}
}
