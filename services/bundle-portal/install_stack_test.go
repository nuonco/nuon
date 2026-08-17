package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2state"
)

type fakeCloudFormationClient struct {
	stack    *cloudformation.DescribeStacksOutput
	events   *cloudformation.DescribeStackEventsOutput
	template *cloudformation.GetTemplateOutput
	err      error
}

func (f *fakeCloudFormationClient) DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return f.stack, f.err
}

func (f *fakeCloudFormationClient) DescribeStackEvents(context.Context, *cloudformation.DescribeStackEventsInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
	return f.events, f.err
}

func (f *fakeCloudFormationClient) GetTemplate(context.Context, *cloudformation.GetTemplateInput, ...func(*cloudformation.Options)) (*cloudformation.GetTemplateOutput, error) {
	return f.template, f.err
}

func TestInstallStackReaderReturnsOrderedFailureTimeline(t *testing.T) {
	started := time.Date(2026, time.August, 18, 10, 0, 0, 0, time.UTC)
	failed := started.Add(time.Minute)
	name := "nuon-airgap-install"
	reason := "resource creation failed"
	client := &fakeCloudFormationClient{
		stack: &cloudformation.DescribeStacksOutput{Stacks: []types.Stack{{
			StackName: &name, StackStatus: types.StackStatusRollbackComplete, CreationTime: &started,
		}}},
		template: &cloudformation.GetTemplateOutput{TemplateBody: aws.String(`{"Resources":{"Cluster":{"Type":"AWS::EKS::Cluster","Properties":{"Version":"1.32"}}}}`)},
		events: &cloudformation.DescribeStackEventsOutput{StackEvents: []types.StackEvent{
			{EventId: aws.String("failed"), Timestamp: &failed, LogicalResourceId: aws.String("Cluster"), ResourceType: aws.String("AWS::EKS::Cluster"), ResourceStatus: types.ResourceStatusCreateFailed, ResourceStatusReason: &reason},
			{EventId: aws.String("started"), Timestamp: &started, LogicalResourceId: &name, ResourceType: aws.String("AWS::CloudFormation::Stack"), ResourceStatus: types.ResourceStatusCreateInProgress},
		}},
	}

	status, err := (&awsInstallStackReader{client: client}).Read(context.Background(), name)
	require.NoError(t, err)
	require.Equal(t, "failed", status.Phase)
	require.Equal(t, reason, status.StatusReason)
	require.Equal(t, []string{"started", "failed"}, []string{status.Events[0].ID, status.Events[1].ID})
	require.Equal(t, "AWS::EKS::Cluster", status.Resources["Cluster"].Type)
	require.Equal(t, "1.32", status.Resources["Cluster"].Properties["Version"])
}

func TestInstallStackReaderTreatsMissingStackAsPending(t *testing.T) {
	reader := &awsInstallStackReader{client: &fakeCloudFormationClient{err: &smithy.GenericAPIError{
		Code: "ValidationError", Message: "Stack with id nuon-airgap-install does not exist",
	}}}

	status, err := reader.Read(context.Background(), "nuon-airgap-install")
	require.NoError(t, err)
	require.Equal(t, "NOT_CREATED", status.Status)
	require.Equal(t, "pending", status.Phase)
}

func TestInstallStackPhase(t *testing.T) {
	require.Equal(t, "in-progress", installStackPhase("CREATE_IN_PROGRESS"))
	require.Equal(t, "finished", installStackPhase("CREATE_COMPLETE"))
	require.Equal(t, "failed", installStackPhase("UPDATE_ROLLBACK_COMPLETE"))
	require.Equal(t, "failed", installStackPhase("CREATE_FAILED"))
}

type fakeInstallStackReader struct {
	status *installStackStatus
}

func (f *fakeInstallStackReader) Read(context.Context, string) (*installStackStatus, error) {
	return f.status, nil
}

func TestInstallStackEndpoint(t *testing.T) {
	portal, err := newPortalServer(day2state.NewLocal(t.TempDir()), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	portal.installStackName = "nuon-airgap-install"
	portal.installStackReader = &fakeInstallStackReader{status: &installStackStatus{
		Name: "nuon-airgap-install", Status: "CREATE_IN_PROGRESS", Phase: "in-progress", Events: []installStackEvent{},
	}}

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/install-stack", nil)
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var status installStackStatus
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &status))
	require.Equal(t, "in-progress", status.Phase)
}

func TestInstallStackEndpointReturnsNullWhenUnconfigured(t *testing.T) {
	portal, err := newPortalServer(day2state.NewLocal(t.TempDir()), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/install-stack", nil)
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, "null", response.Body.String())
}
