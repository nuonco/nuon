package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

type fakeStackPlanner struct {
	outputs   []*cloudformation.DescribeChangeSetOutput
	describes int
	created   *cloudformation.CreateChangeSetInput
}

func (f *fakeStackPlanner) DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return &cloudformation.DescribeStacksOutput{Stacks: []cloudformationtypes.Stack{{Parameters: []cloudformationtypes.Parameter{{ParameterKey: aws.String("Region")}}}}}, nil
}

func (f *fakeStackPlanner) GetTemplate(context.Context, *cloudformation.GetTemplateInput, ...func(*cloudformation.Options)) (*cloudformation.GetTemplateOutput, error) {
	return &cloudformation.GetTemplateOutput{TemplateBody: aws.String(`{"Resources":{"Runner":{"Properties":{"ImageId":"ami-old"}}}}`)}, nil
}

func (f *fakeStackPlanner) CreateChangeSet(_ context.Context, input *cloudformation.CreateChangeSetInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
	f.created = input
	return &cloudformation.CreateChangeSetOutput{Id: aws.String("change-set-id")}, nil
}

func (f *fakeStackPlanner) DescribeChangeSet(context.Context, *cloudformation.DescribeChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {
	output := f.outputs[f.describes]
	f.describes++
	return output, nil
}

func TestCreateStackPlanBuildsChangeSetOnDemand(t *testing.T) {
	next := "next"
	planner := &fakeStackPlanner{outputs: []*cloudformation.DescribeChangeSetOutput{
		{Status: cloudformationtypes.ChangeSetStatusCreateComplete, ExecutionStatus: cloudformationtypes.ExecutionStatusAvailable, NextToken: &next, Changes: []cloudformationtypes.Change{stackPlanChange("Modify", "Runner")}},
		{Status: cloudformationtypes.ChangeSetStatusCreateComplete, ExecutionStatus: cloudformationtypes.ExecutionStatusAvailable, Changes: []cloudformationtypes.Change{stackPlanChange("Add", "Portal")}},
	}}
	candidate := day2.BundleCandidate{
		Bundle: day2.BundleInfo{BundleDigest: "sha256:0123456789abcdef"},
		Deployment: &day2.BundleDeploymentAssets{
			StackTemplateURL: "https://bucket.s3.us-west-2.amazonaws.com/template.json", CandidateBundleKey: "candidate.tar.zst", TargetBundleKey: "bundle.tar.zst",
		},
	}

	result, err := createStackPlan(context.Background(), planner, "install-stack", candidate, []byte(`{"Resources":{"Runner":{"Properties":{"ImageId":"ami-new"}},"Portal":{"Properties":{}}}}`))
	require.NoError(t, err)
	require.Contains(t, result.ChangeSetName, "nuon-bundle-0123456789ab-")
	require.Equal(t, "candidate.tar.zst", result.CandidateBundleKey)
	require.Len(t, result.Changes, 2)
	require.Equal(t, []day2.StackPropertyChange{{Path: "Properties.ImageId", Before: "ami-old", After: "ami-new"}}, result.Changes[0].PropertyChanges)
	require.Equal(t, candidate.Deployment.StackTemplateURL, aws.ToString(planner.created.TemplateURL))
	require.True(t, aws.ToBool(planner.created.Parameters[0].UsePreviousValue))
}

func stackPlanChange(action cloudformationtypes.ChangeAction, logicalID string) cloudformationtypes.Change {
	return cloudformationtypes.Change{ResourceChange: &cloudformationtypes.ResourceChange{
		Action: action, LogicalResourceId: &logicalID, ResourceType: aws.String("AWS::Example::Resource"),
	}}
}
