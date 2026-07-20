package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/require"
)

type fakeCloudFormationCreateStackClient struct {
	createErr *cloudformationtypes.AlreadyExistsException
	events    []cloudformationtypes.StackEvent
}

func (f *fakeCloudFormationCreateStackClient) CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &cloudformation.CreateStackOutput{}, nil
}

func (f *fakeCloudFormationCreateStackClient) DescribeStackEvents(context.Context, *cloudformation.DescribeStackEventsInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
	return &cloudformation.DescribeStackEventsOutput{StackEvents: f.events}, nil
}

func TestManagedCreateStackInput(t *testing.T) {
	input := managedCreateStackInput("nuon-install", "https://templates.example/stack.json", "isv_test")

	require.Equal(t, "nuon-install", aws.ToString(input.StackName))
	require.Equal(t, "https://templates.example/stack.json", aws.ToString(input.TemplateURL))
	require.Equal(t, "isv_test", aws.ToString(input.ClientRequestToken))
	require.Equal(t, []cloudformationtypes.Capability{cloudformationtypes.CapabilityCapabilityNamedIam}, input.Capabilities)
}

func TestCreateManagedStack(t *testing.T) {
	input := managedCreateStackInput("nuon-install", "https://templates.example/stack.json", "isv_test")

	t.Run("create succeeds", func(t *testing.T) {
		require.NoError(t, createManagedStack(context.Background(), &fakeCloudFormationCreateStackClient{}, input))
	})

	t.Run("accepted request is reconciled after a lost response", func(t *testing.T) {
		client := &fakeCloudFormationCreateStackClient{
			createErr: &cloudformationtypes.AlreadyExistsException{},
			events: []cloudformationtypes.StackEvent{
				{ClientRequestToken: aws.String("isv_test")},
			},
		}
		require.NoError(t, createManagedStack(context.Background(), client, input))
	})

	t.Run("unrelated existing stack remains an error", func(t *testing.T) {
		client := &fakeCloudFormationCreateStackClient{
			createErr: &cloudformationtypes.AlreadyExistsException{},
			events: []cloudformationtypes.StackEvent{
				{ClientRequestToken: aws.String("isv_other")},
			},
		}
		var alreadyExists *cloudformationtypes.AlreadyExistsException
		require.True(t, errors.As(createManagedStack(context.Background(), client, input), &alreadyExists))
	})
}
