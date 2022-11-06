package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func getFakeCreateRunnerProfileRequest() CreateRunnerProfileRequest {
	fkr := faker.New()
	var req CreateRunnerProfileRequest
	fkr.Struct().Fill(&req)
	return req
}

func TestCreateRunnerProfile_validateRequest(t *testing.T) {
	tests := map[string]struct {
		reqFn       func() CreateRunnerProfileRequest
		errExpected error
	}{
		"happy path": {
			reqFn: getFakeCreateRunnerProfileRequest,
		},
		"no-org-id": {
			reqFn: func() CreateRunnerProfileRequest {
				req := getFakeCreateRunnerProfileRequest()
				req.OrgID = ""
				return req
			},
			errExpected: fmt.Errorf("CreateRunnerProfileRequest.OrgID"),
		},
		"no-namespace": {
			reqFn: func() CreateRunnerProfileRequest {
				req := getFakeCreateRunnerProfileRequest()
				req.TokenSecretNamespace = ""
				return req
			},
			errExpected: fmt.Errorf("CreateRunnerProfileRequest.TokenSecretNamespace"),
		},
		"no-server-addr": {
			reqFn: func() CreateRunnerProfileRequest {
				req := getFakeCreateRunnerProfileRequest()
				req.OrgServerAddr = ""
				return req
			},
			errExpected: fmt.Errorf("CreateRunnerProfileRequest.OrgServerAddr"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.reqFn()
			err := req.validate()
			if test.errExpected == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, test.errExpected.Error())
			}
		})
	}
}

type testClientRunnerProfileCreator struct {
	mock.Mock
}

func (t *testClientRunnerProfileCreator) UpsertOnDemandRunnerConfig(
	ctx context.Context,
	req *gen.UpsertOnDemandRunnerConfigRequest,
	opts ...grpc.CallOption,
) (*gen.UpsertOnDemandRunnerConfigResponse, error) {
	args := t.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.UpsertOnDemandRunnerConfigResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestRunnerProfileCreatorCreateRunnerProfile(t *testing.T) {
	errCreateWpRunnerProfile := fmt.Errorf("error creating wp runner profile")
	req := getFakeCreateRunnerProfileRequest()

	tests := map[string]struct {
		clientFn    func() clientODRConfigUpserter
		assertFn    func(*testing.T, clientODRConfigUpserter)
		errExpected error
	}{
		"happy path": {
			clientFn: func() clientODRConfigUpserter {
				client := &testClientRunnerProfileCreator{}
				client.On("UpsertOnDemandRunnerConfig", mock.Anything, mock.Anything).
					Return(&gen.UpsertOnDemandRunnerConfigResponse{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client clientODRConfigUpserter) {
				obj := client.(*testClientRunnerProfileCreator)
				obj.AssertNumberOfCalls(t, "UpsertOnDemandRunnerConfig", 1)

				r := obj.Calls[0].Arguments[1].(*gen.UpsertOnDemandRunnerConfigRequest)
				assert.Equal(t, defaultODRImageURL, r.Config.OciUrl)
				assert.Equal(t, "kubernetes", r.Config.PluginType)
				assert.True(t, r.Config.Default)
				assert.Equal(t, gen.Hcl_JSON, r.Config.ConfigFormat)

				var vals map[string]string
				assert.Nil(t, json.Unmarshal(r.Config.PluginConfig, &vals))
				assert.Equal(t, runnerOdrServiceAccountName(req.OrgID), vals["service_account"])
				assert.Equal(t, defaultODRImagePullPolicy, vals["image_pull_policy"])
			},
			errExpected: nil,
		},
		"error returned": {
			clientFn: func() clientODRConfigUpserter {
				client := &testClientRunnerProfileCreator{}
				client.On("UpsertOnDemandRunnerConfig", mock.Anything, mock.Anything).Return(nil, errCreateWpRunnerProfile)
				return client
			},
			assertFn: func(t *testing.T, client clientODRConfigUpserter) {
				obj := client.(*testClientRunnerProfileCreator)
				obj.AssertNumberOfCalls(t, "UpsertOnDemandRunnerConfig", 1)
				assert.NotNil(t, req)
			},
			errExpected: errCreateWpRunnerProfile,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()

			err := createRunnerProfile(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			} else {
				assert.NoError(t, err)
			}

			test.assertFn(t, client)
		})
	}
}
