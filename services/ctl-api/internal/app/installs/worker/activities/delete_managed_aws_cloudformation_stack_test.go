package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
)

type fakeCloudFormationDeleteStackClient struct {
	deleteErr   error
	describeErr error
	events      []cloudformationtypes.StackEvent
	input       *cloudformation.DeleteStackInput
	described   string
}

func (f *fakeCloudFormationDeleteStackClient) DeleteStack(_ context.Context, input *cloudformation.DeleteStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	f.input = input
	return &cloudformation.DeleteStackOutput{}, f.deleteErr
}

func (f *fakeCloudFormationDeleteStackClient) DescribeStackEvents(_ context.Context, input *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
	f.described = aws.ToString(input.StackName)
	return &cloudformation.DescribeStackEventsOutput{StackEvents: f.events}, f.describeErr
}

func TestDeleteManagedStack(t *testing.T) {
	duplicate := &cloudformationtypes.TokenAlreadyExistsException{}
	denied := &smithy.GenericAPIError{Code: "AccessDenied", Message: "not allowed"}
	missing := &smithy.GenericAPIError{Code: "ValidationError", Message: "Stack with id nuon-install does not exist"}
	invalid := &smithy.GenericAPIError{Code: "ValidationError", Message: "Stack cannot be deleted"}
	for name, tt := range map[string]struct {
		deleteErr   error
		describeErr error
		token       string
		status      cloudformationtypes.ResourceStatus
		wantErr     error
		wantLookup  bool
	}{
		"delete succeeds":                {},
		"duplicate deletion in progress": {deleteErr: duplicate, token: "isv-test-delete", status: cloudformationtypes.ResourceStatusDeleteInProgress, wantLookup: true},
		"duplicate deletion completed":   {deleteErr: duplicate, token: "isv-test-delete", status: cloudformationtypes.ResourceStatusDeleteComplete, wantLookup: true},
		"different token":                {deleteErr: duplicate, token: "isv-other-delete", status: cloudformationtypes.ResourceStatusDeleteComplete, wantErr: duplicate, wantLookup: true},
		"non-delete event":               {deleteErr: duplicate, token: "isv-test-delete", status: cloudformationtypes.ResourceStatusUpdateComplete, wantErr: duplicate, wantLookup: true},
		"failed delete event":            {deleteErr: duplicate, token: "isv-test-delete", status: cloudformationtypes.ResourceStatusDeleteFailed, wantErr: duplicate, wantLookup: true},
		"event lookup fails":             {deleteErr: duplicate, describeErr: errors.New("unavailable"), wantErr: duplicate, wantLookup: true},
		"stack already gone":             {deleteErr: missing},
		"other validation error":         {deleteErr: invalid, wantErr: invalid},
		"access denied":                  {deleteErr: denied, wantErr: denied},
	} {
		t.Run(name, func(t *testing.T) {
			client := &fakeCloudFormationDeleteStackClient{deleteErr: tt.deleteErr, describeErr: tt.describeErr}
			if tt.token != "" {
				client.events = []cloudformationtypes.StackEvent{{ClientRequestToken: aws.String(tt.token), ResourceStatus: tt.status}}
			}
			err := deleteManagedStack(context.Background(), client, "nuon-install", "isv-test")
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, "nuon-install", aws.ToString(client.input.StackName))
			require.Equal(t, "isv-test-delete", aws.ToString(client.input.ClientRequestToken))
			if tt.wantLookup {
				require.Equal(t, "nuon-install", client.described)
			} else {
				require.Empty(t, client.described)
			}
		})
	}
}
