package waypoint

//import (
//"context"
//"fmt"
//"testing"

//gomock "github.com/golang/mock/gomock"
//waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
//"github.com/powertoolsdev/go-generics"
//"github.com/stretchr/testify/assert"
//"google.golang.org/protobuf/proto"
//)

//func Test_repo_GetVersionInfo(t *testing.T) {
//errGetVersionInfo := fmt.Errorf("error getting version info")
//versionInfoResp := generics.GetFakeObj[*waypointv1.GetVersionInfoResponse]()

//tests := map[string]struct {
//clientGetter func(*gomock.Controller) clientGetter
//assertFn     func(*testing.T, *waypointv1.GetVersionInfoResponse)
//errExpected  error
//}{
//"happy path": {
//clientGetter: func(mockCtl *gomock.Controller) clientGetter {
//mock := NewMockwaypointClient(mockCtl)
//mock.EXPECT().GetVersionInfo(gomock.Any(), gomock.Any(), gomock.Any()).
//Return(versionInfoResp, nil)

//return func(context.Context) (waypointClient, error) {
//return mock, nil
//}
//},
//assertFn: func(t *testing.T, resp *waypointv1.GetVersionInfoResponse) {
//assert.True(t, proto.Equal(resp, versionInfoResp))
//},
//},
//"unable to get client err": {
//clientGetter: func(mockCtl *gomock.Controller) clientGetter {
//return func(context.Context) (waypointClient, error) {
//return nil, errGetVersionInfo
//}
//},
//errExpected: errGetVersionInfo,
//},
//"client error": {
//clientGetter: func(mockCtl *gomock.Controller) clientGetter {
//mock := NewMockwaypointClient(mockCtl)
//mock.EXPECT().GetVersionInfo(gomock.Any(), gomock.Any(), gomock.Any()).
//Return(nil, errGetVersionInfo)

//return func(context.Context) (waypointClient, error) {
//return mock, nil
//}
//},
//errExpected: errGetVersionInfo,
//},
//}

//for name, test := range tests {
//t.Run(name, func(t *testing.T) {
//ctx := context.Background()
//mockCtl := gomock.NewController(t)
//repo := &repo{
//ClientGetter: test.clientGetter(mockCtl),
//}

//resp, err := repo.GetVersionInfo(ctx)
//if test.errExpected != nil {
//assert.ErrorContains(t, err, test.errExpected.Error())
//return
//}

//assert.NoError(t, err)
//test.assertFn(t, resp)
//})
//}
//}
