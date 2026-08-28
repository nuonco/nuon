package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/plans"
	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/supportsnapshot"
)

func TestSupportSnapshotRoundTripIncludesRunsAndRedactedLogs(t *testing.T) {
	p, dir := testPortal(t)
	finishedAt := time.Now().UTC().Truncate(time.Second)
	writeStateObject(t, dir, operation.BundleKey, operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion, DeploymentID: "prod",
		Release:      &operation.BundleReleaseIdentity{ID: "release-1", Digest: "sha256:" + strings.Repeat("c", 64)},
		Package:      &operation.BundlePackageIdentity{ID: "package-1", Digest: "sha256:" + strings.Repeat("d", 64), Format: "portable-oci", Target: "linux/amd64"},
		BundleDigest: "sha256:" + strings.Repeat("a", 64), ArchiveDigest: "sha256:" + strings.Repeat("b", 64), ActivatedAt: finishedAt,
	})
	staged := operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:" + strings.Repeat("a", 64), StagedAt: finishedAt,
		Bundle:  operation.BundleInfo{BundleDigest: "sha256:" + strings.Repeat("c", 64)},
		Changes: []operation.BundleChange{{Kind: operation.BundleContentKindComponent, Name: "api", Change: operation.BundleChangeChanged}},
	}
	writeStateObject(t, dir, operation.CandidateStageKey(staged.Bundle.BundleDigest, staged.StagedAt), staged)
	writeStateObject(t, dir, "status.json", statestore.Status{
		RunID: "install-run", InstallID: "inst-1", Status: statestore.RunStatusFinished,
		StartedAt: finishedAt.Add(-time.Minute), FinishedAt: &finishedAt,
		Steps: []statestore.StepStatus{{ID: "job-1", Name: "Install API", Status: "finished"}},
	})
	writeStateObject(t, dir, "report.json", map[string]any{"status": "finished", "customer_value": "raw-state-value"})
	writeStateObject(t, dir, operation.RunStatusKey("operation-run"), operation.RunStatus{
		RunID: "operation-run", RefID: "restart-api", RefName: "Restart API", Status: "finished", StartedAt: finishedAt,
		Steps: []operation.RunStep{{ID: "step-1", Name: "Restart", JobID: "job-1", Status: "finished"}},
	})
	writeStateObject(t, dir, operation.RunStatusKey("bundle-plan"), operation.RunStatus{
		RunID: "bundle-plan", RefID: "candidate", RefKind: operation.RefKindBundlePlan, RefName: "Bundle deployment plan", Status: "finished", StartedAt: finishedAt,
		Steps: []operation.RunStep{{ID: "install-stack-plan", Name: "Plan install stack", Status: "finished"}},
	})
	stackPlan := operation.StackCandidate{StackName: "install-stack", ChangeSetName: "candidate", Changes: []operation.StackChange{{Action: "Modify", LogicalResourceID: "Runner"}}}
	writeStateObject(t, dir, operation.RunStepResultKey("bundle-plan", "install-stack-plan"), stackPlan)
	plan := `{"terraform_version":"1.9.0","resource_changes":[{"address":"aws_lambda_function.demo","change":{"actions":["create"]}}]}`
	compressedPlan, err := plans.CompressPlan([]byte(plan))
	require.NoError(t, err)
	writeStateObject(t, dir, statestore.StepResultKey("job-1"), map[string]any{
		"success":                     true,
		"contents_display_compressed": compressedPlan,
	})
	writeJobLog(t, dir, "job-1", strings.Join([]string{
		`{"level":"info","ts":1700000000,"msg":"starting","component":"api","password":"secret"}`,
		`malformed secret line`,
	}, "\n"))
	writeStateObject(t, dir, customermanaged.CapturedInputsKey, customermanaged.CapturedInputs{
		ObservedAt: finishedAt,
		Inputs: []customermanaged.CapturedInput{
			{Name: "region", Value: stringPtr("us-west-2"), ValueAvailable: true, ValueStatus: "provided"},
			{Name: "token", Secret: true, ValueStatus: "redacted"},
		},
	})
	writeStateObject(t, dir, customermanaged.CapturedRolesKey, customermanaged.CapturedRoles{
		ObservedAt: finishedAt,
		Roles:      []customermanaged.CapturedRole{{Name: "Provision", Type: "provision", CloudID: "arn:aws:iam::123:role/provision", Provisioned: true}},
	})
	p.deploymentID = "prod"
	p.cloudProvider = "aws"
	p.cloudAccountID = "123456789012"
	p.cloudRegion = "us-west-2"
	p.installStackName = "install-stack"
	p.installStackReader = &fakeInstallStackReader{status: &installStackStatus{ID: "stack-id", Name: "install-stack", Phase: "finished"}}

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/support-snapshot", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, supportsnapshot.ArchiveContentType, response.Header().Get("Content-Type"))
	require.Contains(t, response.Header().Get("Content-Disposition"), "nuon-support-prod-")

	archive, err := supportsnapshot.Read(bytes.NewReader(response.Body.Bytes()))
	require.NoError(t, err)
	require.NotEmpty(t, archive.Snapshot.Runs)
	require.NotNil(t, archive.Snapshot.StagedBundle)
	require.Equal(t, "sha256:"+strings.Repeat("c", 64), archive.Snapshot.StagedBundle.Bundle.BundleDigest)
	require.Contains(t, archive.Snapshot.Runs, supportsnapshot.Run{RunID: "operation-run", RefID: "restart-api", RefName: "Restart API", Status: "finished", StartedAt: finishedAt, Steps: []supportsnapshot.RunStep{{ID: "step-1", Name: "Restart", JobID: "job-1", Status: "finished", Plan: &supportsnapshot.StepPlan{Kind: "terraform", Content: json.RawMessage(plan)}}}})
	var capturedStackPlan *supportsnapshot.StepPlan
	for _, run := range archive.Snapshot.Runs {
		if run.RunID == "bundle-plan" {
			capturedStackPlan = run.Steps[0].Plan
		}
	}
	require.NotNil(t, capturedStackPlan)
	require.Equal(t, "cloudformation", capturedStackPlan.Kind)
	require.Len(t, archive.Snapshot.Logs, 1)
	require.Len(t, archive.Snapshot.Logs[0].Entries, 1)
	require.Equal(t, "us-west-2", *archive.Snapshot.CurrentInputs.Inputs[0].Value)
	require.Equal(t, "redacted", archive.Snapshot.CurrentInputs.Inputs[1].ValueStatus)
	require.Equal(t, "arn:aws:iam::123:role/provision", archive.Snapshot.Roles.Roles[0].CloudID)
	require.Equal(t, map[string]any{"component": "api"}, archive.Snapshot.Logs[0].Entries[0].Fields)
	require.False(t, archive.Snapshot.IncludeState)
	require.Nil(t, archive.Snapshot.State)
	require.NotContains(t, string(response.Body.Bytes()), "password")
	require.NotContains(t, string(response.Body.Bytes()), "malformed secret line")

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/support-snapshot?include_state=true", nil)
	response = httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	archive, err = supportsnapshot.Read(bytes.NewReader(response.Body.Bytes()))
	require.NoError(t, err)
	require.True(t, archive.Snapshot.IncludeState)
	require.JSONEq(t, `{"status":"finished","customer_value":"raw-state-value"}`, string(archive.Snapshot.State.Report))
	require.Contains(t, archive.Snapshot.Collection.Included, "state")
}

func TestSupportSnapshotRejectsInvalidIncludeState(t *testing.T) {
	p, _ := testPortal(t)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/support-snapshot?include_state=sometimes", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func stringPtr(value string) *string { return &value }
