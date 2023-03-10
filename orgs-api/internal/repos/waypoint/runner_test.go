package waypoint

//import (
//"context"
//"fmt"
//"testing"

//gomock "github.com/golang/mock/gomock"
//"github.com/google/uuid"
//waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
//"github.com/stretchr/testify/assert"
//"google.golang.org/protobuf/proto"
//)

//func Test_repo_GetRunner(t *testing.T) {
//errGetRunner := fmt.Errorf("error getting runner")
//getRunnerResp := &waypointv1.Runner{Id: uuid.NewString()}
//runnerID := uuid.NewString()

//tests := map[string]struct {
//clientGetter func(*gomock.Controller) clientGetter
//assertFn     func(*testing.T, *waypointv1.Runner)
//errExpected  error
//}{
//"happy path": {
//clientGetter: func(mockCtl *gomock.Controller) clientGetter {
//mock := NewMockwaypointClient(mockCtl)
//expectedReq := &waypointv1.GetRunnerRequest{RunnerId: runnerID}
//mock.EXPECT().GetRunner(gomock.Any(), expectedReq, gomock.Any()).
//Return(getRunnerResp, nil)

//return func(context.Context) (waypointClient, error) {
//return mock, nil
//}
//},
//assertFn: func(t *testing.T, resp *waypointv1.Runner) {
//assert.True(t, proto.Equal(resp, getRunnerResp))
//},
//},
//"unable to get client err": {
//clientGetter: func(mockCtl *gomock.Controller) clientGetter {
//return func(context.Context) (waypointClient, error) {
//return nil, errGetRunner
//}
//},
//errExpected: errGetRunner,
//},
//"client error": {
//clientGetter: func(mockCtl *gomock.Controller) clientGetter {
//mock := NewMockwaypointClient(mockCtl)
//mock.EXPECT().GetRunner(gomock.Any(), gomock.Any(), gomock.Any()).
//Return(nil, errGetRunner)

//return func(context.Context) (waypointClient, error) {
//return mock, nil
//}
//},
//errExpected: errGetRunner,
//},
//}

//for name, test := range tests {
//t.Run(name, func(t *testing.T) {
//ctx := context.Background()
//mockCtl := gomock.NewController(t)
//repo := &repo{
//ClientGetter: test.clientGetter(mockCtl),
//}

//resp, err := repo.GetRunner(ctx, runnerID)
//if test.errExpected != nil {
//assert.ErrorContains(t, err, test.errExpected.Error())
//return
//}

//assert.NoError(t, err)
//test.assertFn(t, resp)
//})
//}
//}
