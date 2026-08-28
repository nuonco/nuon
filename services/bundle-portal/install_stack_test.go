package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationstate"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
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
	name := "nuon-customer-managed-install"
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
		Code: "ValidationError", Message: "Stack with id nuon-customer-managed-install does not exist",
	}}}

	status, err := reader.Read(context.Background(), "nuon-customer-managed-install")
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
	portal, err := newPortalServer(operationstate.NewLocal(t.TempDir()), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	portal.installStackName = "nuon-customer-managed-install"
	portal.installStackReader = &fakeInstallStackReader{status: &installStackStatus{
		Name: "nuon-customer-managed-install", Status: "CREATE_IN_PROGRESS", Phase: "in-progress", Events: []installStackEvent{},
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
	portal, err := newPortalServer(operationstate.NewLocal(t.TempDir()), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/install-stack", nil)
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, "null", response.Body.String())
}

func TestInstallationRegistrationEndpoint(t *testing.T) {
	portal, dir := testPortal(t)
	finishedAt := time.Now().UTC()
	writeStateObject(t, dir, operation.BundleKey, operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion, DeploymentID: "vinst1234-prod",
		Release:      &operation.BundleReleaseIdentity{ID: "release-1", Digest: "sha256:" + strings.Repeat("c", 64)},
		Package:      &operation.BundlePackageIdentity{ID: "package-1", Digest: "sha256:" + strings.Repeat("d", 64), Format: "portable-oci", Target: "linux/amd64"},
		BundleDigest: "sha256:" + strings.Repeat("a", 64), ArchiveDigest: "sha256:" + strings.Repeat("b", 64), ActivatedAt: finishedAt,
	})
	writeStateObject(t, dir, "status.json", statestore.Status{
		InstallID: "vinst1234-prod", RunID: "install-run-1", Status: statestore.RunStatusFinished, FinishedAt: &finishedAt,
	})
	portal.deploymentID = "prod"
	portal.cloudProvider = "aws"
	portal.cloudAccountID = "123456789012"
	portal.cloudRegion = "us-east-1"
	portal.installStackName = "install-stack"
	portal.installStackReader = &fakeInstallStackReader{status: &installStackStatus{
		ID: "stack-id", Name: "install-stack", Status: "CREATE_COMPLETE", Phase: "finished",
	}}

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/installation-registration", nil)
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Header().Get("Content-Disposition"), "nuon-installation-registration-prod.json")
	var registration customermanaged.InstallationRegistration
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &registration))
	require.NoError(t, registration.Validate())
	require.Equal(t, "prod", registration.DeploymentID)
	require.Equal(t, "install-run-1", registration.OperationID)
	require.Equal(t, "123456789012", registration.Cloud.AccountID)
}

func TestInstallationRegistrationRemainsBoundToInitialBundle(t *testing.T) {
	portal, dir := testPortal(t)
	installedAt := time.Now().UTC().Add(-time.Hour)
	upgradedAt := time.Now().UTC()
	initial := operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion, DeploymentID: "vinst1234-prod",
		Release:      &operation.BundleReleaseIdentity{ID: "release-1", Digest: "sha256:" + strings.Repeat("c", 64)},
		Package:      &operation.BundlePackageIdentity{ID: "package-1", Digest: "sha256:" + strings.Repeat("d", 64), Format: "portable-oci", Target: "linux/amd64"},
		BundleDigest: "sha256:" + strings.Repeat("a", 64), ArchiveDigest: "sha256:" + strings.Repeat("b", 64), ActivatedAt: installedAt,
	}
	upgraded := operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion, DeploymentID: "vinst1234-prod",
		Release:      &operation.BundleReleaseIdentity{ID: "release-2", Digest: "sha256:" + strings.Repeat("e", 64)},
		Package:      &operation.BundlePackageIdentity{ID: "package-2", Digest: "sha256:" + strings.Repeat("f", 64), Format: "portable-oci", Target: "linux/amd64"},
		BundleDigest: "sha256:" + strings.Repeat("1", 64), ArchiveDigest: "sha256:" + strings.Repeat("2", 64), ActivatedAt: upgradedAt,
	}
	writeStateObject(t, dir, operation.BundleKey, upgraded)
	writeStateObject(t, dir, operation.BundleHistoryKey(initial.BundleDigest), initial)
	writeStateObject(t, dir, operation.BundleHistoryKey(upgraded.BundleDigest), upgraded)
	writeStateObject(t, dir, "status.json", statestore.Status{
		InstallID: "vinst1234-prod", Status: statestore.RunStatusFinished, FinishedAt: &upgradedAt,
	})
	writeStateObject(t, dir, statestore.InstallRunStatusKey("install-1"), statestore.Status{
		InstallID: "vinst1234-prod", BundleDigest: initial.BundleDigest, RunID: "install-1", RunType: statestore.RunTypeInstall,
		Status: statestore.RunStatusFinished, FinishedAt: &installedAt,
	})
	portal.deploymentID = "prod"
	portal.cloudProvider = "aws"
	portal.cloudAccountID = "123456789012"
	portal.cloudRegion = "us-east-1"
	portal.installStackName = "install-stack"
	portal.installStackReader = &fakeInstallStackReader{status: &installStackStatus{
		ID: "stack-id", Name: "install-stack", Status: "UPDATE_COMPLETE", Phase: "finished",
	}}

	registration, err := portal.buildInstallationRegistration(context.Background())
	require.NoError(t, err)
	require.Equal(t, initial.Release.ID, registration.ReleaseID)
	require.Equal(t, initial.Package.ID, registration.PackageID)
	require.Equal(t, installedAt, registration.InstalledAt)

	writeStateObject(t, dir, operation.BundleHistoryKey(initial.BundleDigest), upgraded)
	persisted, err := portal.buildInstallationRegistration(context.Background())
	require.NoError(t, err)
	require.Equal(t, registration, persisted)
}
