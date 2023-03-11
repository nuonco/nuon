package provision

//
// TODO(jm): working on pulling this into `go-workflows-meta` to send notifications into slack channels, when provided
// as part of the message.
//

//func Test_notifierImpl_sendSuccessNotification(t *testing.T) {
//errUnableToSend := fmt.Errorf("unableToSend")
//req := generics.GetFakeObj[FinishRequest]()
//req.Success = true
//assert.Nil(t, req.validate())

//tests := map[string]struct {
//senderFn    func() sender.NotificationSender
//assertFn    func(sender.NotificationSender)
//errExpected error
//}{
//"happy path": {
//senderFn: func() sender.NotificationSender {
//s := &mockSender{}
//s.On("Send", mock.Anything, mock.Anything).Return(nil)
//return s
//},
//assertFn: func(sender sender.NotificationSender) {
//obj := sender.(*mockSender)
//obj.AssertNumberOfCalls(t, "Send", 1)
//notif := obj.Calls[0].Arguments[1].(string)
//assert.NotEmpty(t, notif)
//assert.Contains(t, notif, "success")
//},
//errExpected: nil,
//},
//"error": {
//senderFn: func() sender.NotificationSender {
//s := &mockSender{}
//s.On("Send", mock.Anything, mock.Anything).Return(errUnableToSend)
//return s
//},
//assertFn: func(sender sender.NotificationSender) {
//obj := sender.(*mockSender)
//obj.AssertNumberOfCalls(t, "Send", 1)
//notif := obj.Calls[0].Arguments[1].(string)
//assert.NotEmpty(t, notif)
//},
//errExpected: errUnableToSend,
//},
//}
//for name, test := range tests {
//t.Run(name, func(t *testing.T) {
//s := &notifierImpl{}
//sender := test.senderFn()

//err := s.sendSuccessNotification(context.Background(), req, sender)
//if test.errExpected != nil {
//assert.ErrorContains(t, err, test.errExpected.Error())
//} else {
//assert.Nil(t, err)
//}

//test.assertFn(sender)
//})
//}
//}

//func Test_notifierImpl_sendErrorNotification(t *testing.T) {
//errUnableToSend := fmt.Errorf("unableToSend")
//req := FinishRequest{
//ProvisionRequest:	       generics.GetFakeObj[*installsv1.ProvisionRequest](),
//InstallationsBucket:	       "nuon-installations-stage",
//Success:		       false,
//ErrorStep:		       "destroy_step",
//ErrorMessage:		       "failed to destroy",
//InstallationsAccessIAMRoleARN: "role-arn",
//}
//assert.Nil(t, req.validate())

//tests := map[string]struct {
//senderFn    func() sender.NotificationSender
//assertFn    func(sender.NotificationSender)
//errExpected error
//}{
//"happy path": {
//senderFn: func() sender.NotificationSender {
//s := &mockSender{}
//s.On("Send", mock.Anything, mock.Anything).Return(nil)
//return s
//},
//assertFn: func(sender sender.NotificationSender) {
//obj := sender.(*mockSender)
//obj.AssertNumberOfCalls(t, "Send", 1)
//notif := obj.Calls[0].Arguments[1].(string)
//assert.NotEmpty(t, notif)
//assert.Contains(t, notif, "error")
//assert.Contains(t, notif, req.ErrorMessage)
//assert.Contains(t, notif, req.ErrorStep)
//},
//errExpected: nil,
//},
//"error": {
//senderFn: func() sender.NotificationSender {
//s := &mockSender{}
//s.On("Send", mock.Anything, mock.Anything).Return(errUnableToSend)
//return s
//},
//assertFn: func(sender sender.NotificationSender) {
//obj := sender.(*mockSender)
//obj.AssertNumberOfCalls(t, "Send", 1)
//notif := obj.Calls[0].Arguments[1].(string)
//assert.NotEmpty(t, notif)
//},
//errExpected: errUnableToSend,
//},
//}
//for name, test := range tests {
//t.Run(name, func(t *testing.T) {
//s := &notifierImpl{}
//sender := test.senderFn()

//err := s.sendErrorNotification(context.Background(), req, sender)
//if test.errExpected != nil {
//assert.ErrorContains(t, err, test.errExpected.Error())
//} else {
//assert.Nil(t, err)
//}

//test.assertFn(sender)
//})
//}
//}

//func Test_sendStartNotification(t *testing.T) {
//tests := map[string]struct {
//fn	    func(*testing.T, func(string) bool) sender.NotificationSender
//req	    StartWorkflowRequest
//errExpected error
//}{
//"happy path": {
//req: generics.GetFakeObj[StartWorkflowRequest](),
//fn: func(t *testing.T, matcher func(string) bool) sender.NotificationSender {
//ms := &mockSender{}
//ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(nil).Once()

//return ms
//},
//},

//"error on send": {
//req:	     generics.GetFakeObj[StartWorkflowRequest](),
//errExpected: fmt.Errorf("send error"),
//fn: func(t *testing.T, matcher func(string) bool) sender.NotificationSender {
//ms := &mockSender{}
//ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(fmt.Errorf("send error")).Once()

//return ms
//},
//},

//"error without sender": {
//req:	     generics.GetFakeObj[StartWorkflowRequest](),
//errExpected: errNoValidSender,
//fn: func(t *testing.T, matcher func(string) bool) sender.NotificationSender {
//return nil
//},
//},
//}

//for name, test := range tests {
//t.Run(name, func(t *testing.T) {
//matcher := func(s string) bool {
//var accum []bool
//for _, v := range []string{test.req.AppID, test.req.InstallID, test.req.OrgID} {
//accum = append(accum, assert.Contains(t, s, v))
//}
//accum = append(accum, assert.Contains(t, s, "started provisioning sandbox"))
//return !slices.Contains(accum, false)
//}

//s := test.fn(t, matcher)
//n := &starterImpl{sender: s}

//err := n.sendStartNotification(context.Background(), test.req)
//if test.errExpected != nil {
//assert.ErrorContains(t, err, test.errExpected.Error())
//return
//}
//assert.NoError(t, err)

//if s, ok := s.(*mockSender); ok {
//s.AssertExpectations(t)
//}
//})
//}
//}
