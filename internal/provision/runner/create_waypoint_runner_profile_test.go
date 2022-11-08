package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type testWaypointClientRunnerProfileCreator struct {
	mock.Mock
}

func (t *testWaypointClientRunnerProfileCreator) UpsertOnDemandRunnerConfig(ctx context.Context, req *gen.UpsertOnDemandRunnerConfigRequest, _ ...grpc.CallOption) (*gen.UpsertOnDemandRunnerConfigResponse, error) {
	args := t.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.UpsertOnDemandRunnerConfigResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func getFakeCreateWaypointRunnerProfileRequest() CreateWaypointRunnerProfileRequest {
	id := uuid.NewString()
	return CreateWaypointRunnerProfileRequest{
		InstallID:            id,
		OrgID:                id,
		TokenSecretNamespace: "default",
		OrgServerAddr:        fmt.Sprintf("%s.nuon.co", uuid.NewString()),
	}
}

func Test_wpRunnerProfileCreator_createWaypointRunnerProfile(t *testing.T) {
	errCreateWpRunnerProfile := fmt.Errorf("error creating wp runner profile")
	req := getFakeCreateWaypointRunnerProfileRequest()

	tests := map[string]struct {
		clientFn    func() waypointClientODRConfigUpserter
		assertFn    func(*testing.T, waypointClientODRConfigUpserter)
		errExpected error
	}{
		"happy path": {
			clientFn: func() waypointClientODRConfigUpserter {
				client := &testWaypointClientRunnerProfileCreator{}
				client.On("UpsertOnDemandRunnerConfig", mock.Anything, mock.Anything).Return(&gen.UpsertOnDemandRunnerConfigResponse{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientODRConfigUpserter) {
				obj := client.(*testWaypointClientRunnerProfileCreator)
				obj.AssertNumberOfCalls(t, "UpsertOnDemandRunnerConfig", 1)

				r := obj.Calls[0].Arguments[1].(*gen.UpsertOnDemandRunnerConfigRequest)
				assert.Equal(t, req.InstallID, r.Config.Name)
				assert.Equal(t, defaultODRImageURL, r.Config.OciUrl)
				assert.Equal(t, "kubernetes", r.Config.PluginType)
				assert.False(t, r.Config.Default)
				assert.Equal(t, gen.Hcl_JSON, r.Config.ConfigFormat)

				var vals map[string]string
				assert.Nil(t, json.Unmarshal(r.Config.PluginConfig, &vals))
				assert.Equal(t, runnerOdrServiceAccountName(req.InstallID), vals["service_account"])
				assert.Equal(t, defaultODRImagePullPolicy, vals["image_pull_policy"])
			},
			errExpected: nil,
		},
		"error returned": {
			clientFn: func() waypointClientODRConfigUpserter {
				client := &testWaypointClientRunnerProfileCreator{}
				client.On("UpsertOnDemandRunnerConfig", mock.Anything, mock.Anything).Return(nil, errCreateWpRunnerProfile)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientODRConfigUpserter) {
				obj := client.(*testWaypointClientRunnerProfileCreator)
				obj.AssertNumberOfCalls(t, "UpsertOnDemandRunnerConfig", 1)
				assert.NotNil(t, req)
			},
			errExpected: errCreateWpRunnerProfile,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			r := wpRunnerProfileCreator{}

			err := r.createWaypointRunnerProfile(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(t, client)
		})
	}
}
