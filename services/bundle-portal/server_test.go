package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/nuonco/nuon/pkg/plans"
	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/bundleupgrade"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationstate"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

func writeStateObject(t *testing.T, dir, key string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	path := filepath.Join(dir, filepath.FromSlash(key))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, raw, 0o600))
}

func testPortal(t *testing.T) (*portalServer, string) {
	t.Helper()
	dir := t.TempDir()
	now := time.Now().UTC()
	writeStateObject(t, dir, operation.CatalogKey, operation.Catalog{
		SchemaVersion: operation.SchemaVersion,
		DeploymentID:  "dep-1",
		BundleDigest:  "sha256:bundle",
		Refs: []operation.CatalogRef{{
			ID:        "restart-api",
			Kind:      operation.RefKindAction,
			Name:      "Restart API",
			Component: "api",
		}},
	})
	writeStateObject(t, dir, "health/latest.json", customermanaged.HealthSnapshot{
		ObservedAt: now,
		Components: []customermanaged.ComponentHealth{{ComponentID: "cmp-1", ComponentName: "api", ComponentType: "helm", Health: "healthy"}},
	})
	writeStateObject(t, dir, customermanaged.RunnerHeartbeatKey, customermanaged.RunnerHeartbeat{
		RunnerID: "customer-managed-dep-1", SessionID: "session-1", Version: "v1", BundleDigest: "sha256:bundle",
		Capabilities: []string{customermanaged.RunnerCapabilityCandidateArtifactPlans}, StartedAt: now.Add(-time.Minute), ObservedAt: now,
	})
	writeStateObject(t, dir, operation.RunStatusKey("run-1"), operation.RunStatus{RunID: "run-1", RefID: "restart-api", StartedAt: now})
	writeStateObject(t, dir, operation.BundleKey, operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion,
		DeploymentID:  "dep-1",
		BundleDigest:  "sha256:bundle",
		ActivatedAt:   now,
		Target:        &operation.BundleTarget{OS: "linux", Architecture: "arm64"},
		Verification:  operation.BundleVerification{BlobsVerified: true, EnvelopeParsed: true},
		TotalSize:     300,
		Contents: []operation.BundleContent{
			{Kind: operation.BundleContentKindComponent, Name: "api", Detail: "helm_chart", Digest: "sha256:a", Size: 100},
			{Kind: operation.BundleContentKindImage, Name: "nginx", Detail: "docker.io/nginx", Digest: "sha256:i", Size: 200},
		},
	})
	writeStateObject(t, dir, operation.BundleHistoryKey("sha256:bundle"), operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:bundle", ActivatedAt: now,
		Contents: []operation.BundleContent{{Kind: operation.BundleContentKindComponent, Name: "api", Digest: "sha256:a", ConfigDigest: "sha256:config-a"}},
	})
	writeStateObject(t, dir, operation.BundleHistoryKey("sha256:old"), operation.BundleInfo{
		SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:old", ActivatedAt: now.Add(-time.Hour),
		Contents: []operation.BundleContent{{Kind: operation.BundleContentKindComponent, Name: "api", Digest: "sha256:old-a", ConfigDigest: "sha256:old-config-a"}},
	})
	p, err := newPortalServer(operationstate.NewLocal(dir), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	return p, dir
}

func TestBundleHistoryIncludesInventoryComparisons(t *testing.T) {
	p, _ := testPortal(t)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var body struct {
		Comparisons []bundleHistoryComparison `json:"comparisons"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Len(t, body.Comparisons, 1)
	require.Equal(t, "sha256:old", body.Comparisons[0].PreviousDigest)
	require.Equal(t, "sha256:bundle", body.Comparisons[0].BundleDigest)
	require.True(t, body.Comparisons[0].Available)
	require.Equal(t, []operation.BundleChange{{
		Kind: operation.BundleContentKindComponent, Name: "api", Change: operation.BundleChangeChanged,
		PreviousDigest: "sha256:old-a", CandidateDigest: "sha256:a",
		PreviousConfig: "sha256:old-config-a", CandidateConfig: "sha256:config-a",
		PlanStepID: "deploy-api-plan", ApplyStepID: "deploy-api-apply",
	}}, body.Comparisons[0].Changes)
}

func TestUploadBundleCandidateStreamsArchiveToStager(t *testing.T) {
	p, _ := testPortal(t)
	var stagedPath string
	var stagedName string
	p.stageBundle = func(_ context.Context, path, name string, progress func(bundleupgrade.Progress)) (*bundleupgrade.Result, error) {
		stagedPath = path
		stagedName = name
		progress(bundleupgrade.Progress{Phase: "verifying", Detail: "Verifying bundle content"})
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, []byte("bundle archive"), raw)
		return &bundleupgrade.Result{Candidate: operation.BundleCandidate{Bundle: operation.BundleInfo{BundleDigest: "sha256:next"}}}, nil
	}
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate", strings.NewReader("bundle archive"))
	request.Header.Set("X-CSRF-Token", "secret")
	request.Header.Set("X-Nuon-Bundle-Filename", "release%202026-08-20.tar.zst")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var candidate operation.BundleCandidate
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &candidate))
	require.Equal(t, "sha256:next", candidate.Bundle.BundleDigest)
	require.Equal(t, "release 2026-08-20.tar.zst", stagedName)
	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle-candidate/upload-status", nil)
	response = httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"state":"complete","phase":"complete","detail":"Bundle staged and ready for review","updated_at":"`+p.bundleUploadStatus.UpdatedAt.Format(time.RFC3339Nano)+`"}`, response.Body.String())
	_, err := os.Stat(stagedPath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestUploadBundleCandidateRejectsUnavailableAndConcurrentUploads(t *testing.T) {
	p, _ := testPortal(t)
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate", strings.NewReader("bundle"))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code)

	p.stageBundle = func(context.Context, string, string, func(bundleupgrade.Progress)) (*bundleupgrade.Result, error) {
		return nil, nil
	}
	p.bundleUploadMu.Lock()
	defer p.bundleUploadMu.Unlock()
	request = httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate", strings.NewReader("bundle"))
	request.Header.Set("X-CSRF-Token", "secret")
	response = httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code)
}

func TestPlanBundleCandidateStackPersistsOnlyAfterRequest(t *testing.T) {
	p, dir := testPortal(t)
	activeBefore, found, err := p.store.Get(context.Background(), operation.BundleKey)
	require.NoError(t, err)
	require.True(t, found)
	digest := "sha256:0123456789abcdef"
	writeStateObject(t, dir, operation.StagedCandidateKey, operation.BundleCandidate{
		SchemaVersion: 1, PreviousDigest: "sha256:bundle", StagedAt: time.Now().UTC(),
		Bundle: operation.BundleInfo{BundleDigest: digest},
		Deployment: &operation.BundleDeploymentAssets{
			StackTemplateURL: "https://bucket.s3.us-west-2.amazonaws.com/template.json", CandidateBundleKey: "candidate.tar.zst", TargetBundleKey: "bundle.tar.zst",
		},
	})
	writeStateObject(t, dir, stackCandidateTemplateKey(digest), json.RawMessage(`{"Resources":{"Runner":{"Properties":{"ImageId":"ami-new"}}}}`))
	_, found, err = p.store.Get(context.Background(), operation.StackCandidateKey)
	require.NoError(t, err)
	require.False(t, found)
	p.installStackName = "install-stack"
	p.stackPlanner = &fakeStackPlanner{outputs: []*cloudformation.DescribeChangeSetOutput{{
		Status: cloudformationtypes.ChangeSetStatusCreateComplete, ExecutionStatus: cloudformationtypes.ExecutionStatusAvailable,
		Changes: []cloudformationtypes.Change{stackPlanChange("Modify", "Runner")},
	}}}

	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/plan-stack", strings.NewReader(`{"bundle_digest":"`+digest+`"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Eventually(t, func() bool {
		raw, found, err := p.store.Get(context.Background(), operation.StackCandidateKey)
		if err != nil || !found {
			return false
		}
		var planned operation.StackCandidate
		return json.Unmarshal(raw, &planned) == nil && planned.BundleDigest == digest && planned.Status == string(cloudformationtypes.ChangeSetStatusCreateComplete)
	}, time.Second, 10*time.Millisecond)
	require.Eventually(t, func() bool {
		runs, err := p.listRuns(httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/runs", nil))
		if err != nil {
			return false
		}
		for _, run := range runs {
			if run.RefKind == "bundle-plan" && run.BundleDigest == digest {
				return run.Status == operation.RunStatusFinished && run.FinishedAt != nil && len(run.Steps) == 1
			}
		}
		return false
	}, time.Second, 10*time.Millisecond)
	activeAfter, found, err := p.store.Get(context.Background(), operation.BundleKey)
	require.NoError(t, err)
	require.True(t, found)
	require.JSONEq(t, string(activeBefore), string(activeAfter))
}

func TestPlanBundleCandidateStackRejectsConcurrentPlan(t *testing.T) {
	p, dir := testPortal(t)
	digest := "sha256:0123456789abcdef"
	writeStateObject(t, dir, operation.StagedCandidateKey, operation.BundleCandidate{
		SchemaVersion: 1, PreviousDigest: "sha256:bundle", StagedAt: time.Now().UTC(),
		Bundle: operation.BundleInfo{BundleDigest: digest},
		Deployment: &operation.BundleDeploymentAssets{
			StackTemplateURL: "https://bucket.s3.us-west-2.amazonaws.com/template.json", CandidateBundleKey: "candidate.tar.zst", TargetBundleKey: "bundle.tar.zst",
		},
	})
	writeStateObject(t, dir, stackCandidateTemplateKey(digest), json.RawMessage(`{"Resources":{}}`))
	p.installStackName = "install-stack"
	p.stackPlanner = &fakeStackPlanner{}
	p.stackPlanMu.Lock()
	defer p.stackPlanMu.Unlock()

	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/plan-stack", strings.NewReader(`{"bundle_digest":"`+digest+`"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code)
	require.Contains(t, response.Body.String(), "another bundle plan is already running")
}

func TestPlanBundleCandidateStackDispatchesCompleteCandidatePlanRun(t *testing.T) {
	p, dir := testPortal(t)
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	recordKey := operation.CandidateStageKey(digest, time.Now().UTC())
	candidate := operation.BundleCandidate{
		SchemaVersion: 1, PreviousDigest: "sha256:bundle", StagedAt: time.Now().UTC(),
		Bundle: operation.BundleInfo{DeploymentID: "dep-1", BundleDigest: digest},
		Changes: []operation.BundleChange{
			{Kind: operation.BundleContentKindSandbox, Name: "terraform", Change: operation.BundleChangeChanged, PlanStepID: "sandbox-plan"},
			{Kind: operation.BundleContentKindComponent, Name: "api", Change: operation.BundleChangeChanged, PlanStepID: "deploy-api-plan"},
		},
		Deployment: &operation.BundleDeploymentAssets{
			StackTemplateURL:   "https://bucket.s3.us-west-2.amazonaws.com/template.json",
			CandidateBundleKey: "state/bundle/candidates/candidate.tar.zst", TargetBundleKey: "state/bundle.tar.zst",
		},
	}
	writeStateObject(t, dir, recordKey, candidate)
	writeStateObject(t, dir, stackCandidateTemplateKey(digest), json.RawMessage(`{"Resources":{}}`))
	p.installStackName = "install-stack"
	p.stackPlanner = &fakeStackPlanner{outputs: []*cloudformation.DescribeChangeSetOutput{{
		Status: cloudformationtypes.ChangeSetStatusCreateComplete, ExecutionStatus: cloudformationtypes.ExecutionStatusAvailable,
	}}}

	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/plan-stack", strings.NewReader(`{"bundle_digest":"`+digest+`"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	require.Eventually(t, func() bool {
		keys, err := p.store.List(context.Background(), operation.RequestsPrefix)
		if err != nil || len(keys) != 1 {
			return false
		}
		raw, found, err := p.store.Get(context.Background(), keys[0])
		if err != nil || !found {
			return false
		}
		var planRequest operation.Request
		if json.Unmarshal(raw, &planRequest) != nil {
			return false
		}
		if planRequest.RefKind != operation.RefKindBundlePlan || planRequest.RefID != recordKey || planRequest.RunID == "" || planRequest.CandidateRecordKey != recordKey {
			return false
		}
		require.Equal(t, []string{"sandbox-plan", "deploy-api-plan"}, planRequest.PlanStepIDs)
		runRaw, found, err := p.store.Get(context.Background(), operation.RunStatusKey(planRequest.RunID))
		if err != nil || !found {
			return false
		}
		var run operation.RunStatus
		return json.Unmarshal(runRaw, &run) == nil && run.RefID == recordKey && run.Status == operation.RunStatusInProgress && len(run.Steps) == 3 && run.Steps[0].Status == operation.RunStatusFinished
	}, time.Second, 10*time.Millisecond)
}

func TestBundleUsesAppendOnlyStageWhenMutablePointerIsStale(t *testing.T) {
	p, dir := testPortal(t)
	stagedAt := time.Now().UTC()
	writeStateObject(t, dir, operation.StagedCandidateKey, operation.BundleCandidate{
		Bundle: operation.BundleInfo{BundleDigest: "sha256:stale"}, StagedAt: stagedAt.Add(-time.Minute),
	})
	writeStateObject(t, dir, operation.CandidateStageKey("sha256:new", stagedAt), operation.BundleCandidate{
		Bundle: operation.BundleInfo{BundleDigest: "sha256:new"}, StagedAt: stagedAt,
	})

	candidate, recordKey, found, err := p.latestBundleCandidateRecord(context.Background())
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "sha256:new", candidate.Bundle.BundleDigest)
	require.Equal(t, operation.CandidateStageKey("sha256:new", stagedAt), recordKey)
}

func TestPortalSecurityAndAPI(t *testing.T) {
	p, dir := testPortal(t)
	h := p.handler()

	request := httptest.NewRequest(http.MethodGet, "http://evil.example/api/catalog", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusForbidden, response.Code)

	request = httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/dispatch", strings.NewReader(`{"ref_id":"restart-api"}`))
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusForbidden, response.Code)

	request = httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/dispatch", strings.NewReader(`{"ref_id":"restart-api"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var dispatched map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &dispatched))
	require.NotEmpty(t, dispatched["dispatch_id"])
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(operation.RequestKey(dispatched["dispatch_id"]))))
	require.NoError(t, err)
	var saved operation.Request
	require.NoError(t, json.Unmarshal(raw, &saved))
	require.Equal(t, operation.SourcePortal, saved.Source)
	require.Equal(t, "operator", saved.RequestedBy)
	require.Equal(t, "sha256:bundle", saved.BundleDigest)

	for _, path := range []string{"/api/catalog", "/api/bundle", "/api/health", "/api/runner-heartbeat", "/api/runs", "/api/runs/run-1"} {
		request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234"+path, nil)
		response = httptest.NewRecorder()
		h.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code, path)
		require.True(t, json.Valid(response.Body.Bytes()), path)
		require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
		require.Equal(t, "DENY", response.Header().Get("X-Frame-Options"))
		require.Contains(t, response.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
	}
}

func TestPortalRunnerHeartbeatUnknownWhenAbsent(t *testing.T) {
	p, err := newPortalServer(operationstate.NewLocal(t.TempDir()), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/runner-heartbeat", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, "null", response.Body.String())
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
}

func TestPortalRejectsCandidatePlansFromIncompatibleRunner(t *testing.T) {
	p, dir := testPortal(t)
	writeStateObject(t, dir, customermanaged.RunnerHeartbeatKey, customermanaged.RunnerHeartbeat{RunnerID: "old-runner"})

	err := p.requireRunnerCapability(context.Background(), customermanaged.RunnerCapabilityCandidateArtifactPlans)
	require.ErrorContains(t, err, "update the runner")
}

func TestPortalPlanEndpoint(t *testing.T) {
	p, dir := testPortal(t)
	h := p.handler()
	planJSON := []byte(`{"format_version":"1.2"}`)
	path := filepath.Join(dir, filepath.FromSlash(operation.JobPlanKey("run-1--drift-1")))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, planJSON, 0o600))

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/plans/run-1--drift-1", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, planJSON, response.Body.Bytes())

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/plans/..%2Fstatus", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/plans/missing-job", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestPortalStepPlanEndpoint(t *testing.T) {
	p, dir := testPortal(t)
	h := p.handler()
	planJSON := []byte(`{"deploy_plan":{"install_id":"in1"}}`)
	path := filepath.Join(dir, filepath.FromSlash(statestore.StepPlanKey("job-deploy-1")))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, planJSON, 0o600))

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-plans/job-deploy-1", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, planJSON, response.Body.Bytes())

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-plans/..%2Fstatus", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-plans/not-rendered-yet", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestPortalStepResultEndpoint(t *testing.T) {
	p, dir := testPortal(t)
	h := p.handler()

	tfPlan := `{"terraform_version":"1.9.0","resource_changes":[{"address":"aws_acm_certificate.this","change":{"actions":["create"],"after":{"domain_name":"demo.example.com"}}}]}`
	compressed, err := plans.CompressPlan([]byte(tfPlan))
	require.NoError(t, err)
	writeStateObject(t, dir, statestore.StepResultKey("job-tf-1"), map[string]any{
		"success":                     true,
		"contents_display_compressed": compressed,
	})

	k8sPlan := `{"plan":"1 resource changed","op":"update","k8s_content_diff":[{"kind":"Deployment","name":"guestbook-ui","namespace":"guestbook","op":"update","entries":[{"path":"spec.replicas","original":1,"applied":3,"type":"changed"}]}]}`
	k8sCompressed, err := plans.CompressPlan([]byte(k8sPlan))
	require.NoError(t, err)
	writeStateObject(t, dir, statestore.StepResultKey("job-k8s-1"), map[string]any{
		"success":             true,
		"contents_compressed": k8sCompressed,
	})

	writeStateObject(t, dir, statestore.StepResultKey("job-plain-1"), map[string]any{
		"success":  false,
		"contents": "helm upgrade failed",
	})

	var payload struct {
		Success bool            `json:"success"`
		Kind    string          `json:"kind"`
		Content json.RawMessage `json:"content"`
	}

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-results/job-tf-1", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	require.Equal(t, "terraform", payload.Kind)
	require.JSONEq(t, tfPlan, string(payload.Content))

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-results/job-k8s-1", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, "kubernetes_manifest", payload.Kind)
	require.JSONEq(t, k8sPlan, string(payload.Content))

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-results/job-plain-1", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.False(t, payload.Success)
	require.Equal(t, "unknown", payload.Kind)
	require.Equal(t, `"helm upgrade failed"`, string(payload.Content))

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-results/..%2Fstatus", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/step-results/no-result-yet", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestPortalBundleEndpoint(t *testing.T) {
	p, dir := testPortal(t)
	writeStateObject(t, dir, operation.CandidateKey, operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:bundle",
		Bundle: operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:next"},
	})
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		Active    *operation.BundleInfo      `json:"active"`
		Candidate *operation.BundleCandidate `json:"candidate"`
		History   []operation.BundleInfo     `json:"history"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.NotNil(t, payload.Active)
	require.Equal(t, "sha256:bundle", payload.Active.BundleDigest)
	require.Len(t, payload.Active.Contents, 2)
	require.Equal(t, int64(300), payload.Active.TotalSize)
	require.Equal(t, "sha256:next", payload.Candidate.Bundle.BundleDigest)
	require.Len(t, payload.History, 2)
	require.Equal(t, "sha256:bundle", payload.History[0].BundleDigest)
	require.Equal(t, "sha256:old", payload.History[1].BundleDigest)
}

func TestPortalBundleEndpointEmptyState(t *testing.T) {
	p, err := newPortalServer(operationstate.NewLocal(t.TempDir()), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"active":null,"candidate":null,"stack_candidate":null,"history":[],"comparisons":[]}`, response.Body.String())
}

func TestPortalBundleEndpointShowsStagedCandidateAndStackChanges(t *testing.T) {
	p, dir := testPortal(t)
	writeStateObject(t, dir, operation.CandidateKey, operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:old",
		Bundle: operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:bundle"},
	})
	writeStateObject(t, dir, operation.StagedCandidateKey, operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:bundle",
		Bundle: operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:next"},
	})
	writeStateObject(t, dir, operation.StackCandidateKey, operation.StackCandidate{
		SchemaVersion: operation.SchemaVersion, BundleDigest: "sha256:next", StackName: "install-stack", ChangeSetName: "candidate-next",
		Status: "CREATE_COMPLETE", ExecutionStatus: "AVAILABLE",
		Changes: []operation.StackChange{{Action: "Modify", LogicalResourceID: "Runner", ResourceType: "AWS::AutoScaling::LaunchConfiguration", Replacement: "True"}},
	})
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		CandidateRecordKey string                     `json:"candidate_record_key"`
		Candidate          *operation.BundleCandidate `json:"candidate"`
		StackCandidate     *operation.StackCandidate  `json:"stack_candidate"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, operation.StagedCandidateKey, payload.CandidateRecordKey)
	require.Equal(t, "sha256:next", payload.Candidate.Bundle.BundleDigest)
	require.Equal(t, "Runner", payload.StackCandidate.Changes[0].LogicalResourceID)
}

func TestPortalBundleEndpointHidesActivatedCandidate(t *testing.T) {
	p, dir := testPortal(t)
	writeStateObject(t, dir, operation.CandidateKey, operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:old",
		Bundle: operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:bundle"},
	})
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var payload struct {
		Candidate *operation.BundleCandidate `json:"candidate"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Nil(t, payload.Candidate)
}

func TestPortalClearsBundleCandidateWithoutDeletingHistory(t *testing.T) {
	p, dir := testPortal(t)
	digest := "sha256:next"
	stagedAt := time.Now().UTC().Add(-time.Minute)
	candidate := operation.BundleCandidate{
		SchemaVersion:  operation.SchemaVersion,
		PreviousDigest: "sha256:bundle",
		StagedAt:       stagedAt,
		Bundle:         operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: digest},
	}
	stageKey := operation.CandidateStageKey(digest, stagedAt)
	writeStateObject(t, dir, operation.StagedCandidateKey, candidate)
	writeStateObject(t, dir, stageKey, candidate)

	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/clear", strings.NewReader(`{"bundle_digest":"`+digest+`"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(stageKey)))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, filepath.FromSlash(operation.StagedCandidateKey)))
	require.NoError(t, err)

	keys, err := p.store.List(context.Background(), operation.CandidateStagesPrefix)
	require.NoError(t, err)
	var dismissal operation.BundleCandidateDismissal
	dismissalCount := 0
	for _, key := range keys {
		if !strings.Contains(key, "/dismissed/") {
			continue
		}
		raw, found, getErr := p.store.Get(context.Background(), key)
		require.NoError(t, getErr)
		require.True(t, found)
		require.NoError(t, json.Unmarshal(raw, &dismissal))
		dismissalCount++
	}
	require.Equal(t, 1, dismissalCount)
	require.Equal(t, digest, dismissal.BundleDigest)
	require.Equal(t, "operator", dismissal.RequestedBy)

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle", nil)
	response = httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var payload struct {
		Candidate *operation.BundleCandidate `json:"candidate"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Nil(t, payload.Candidate)

	restaged := candidate
	restaged.StagedAt = dismissal.DismissedAt.Add(time.Second)
	writeStateObject(t, dir, operation.CandidateStageKey(digest, restaged.StagedAt), restaged)
	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/bundle", nil)
	response = httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.NotNil(t, payload.Candidate)
	require.Equal(t, restaged.StagedAt, payload.Candidate.StagedAt)
}

func TestPortalRejectsClearingStaleBundleCandidate(t *testing.T) {
	p, dir := testPortal(t)
	writeStateObject(t, dir, operation.StagedCandidateKey, operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion,
		StagedAt:      time.Now().UTC(),
		Bundle:        operation.BundleInfo{BundleDigest: "sha256:next"},
	})

	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/clear", strings.NewReader(`{"bundle_digest":"sha256:stale"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code, response.Body.String())

	keys, err := p.store.List(context.Background(), operation.CandidateStagesPrefix)
	require.NoError(t, err)
	for _, key := range keys {
		require.NotContains(t, key, "/dismissed/")
	}
}

func TestPortalApprovesMatchingPlannedBundleCandidate(t *testing.T) {
	p, dir := testPortal(t)
	candidate := operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:bundle", StagedAt: time.Now().UTC(),
		Bundle:  operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:next"},
		Changes: []operation.BundleChange{{Kind: operation.BundleContentKindComponent, Name: "api", Change: operation.BundleChangeChanged, PlanStepID: "deploy-api-plan", ApplyStepID: "deploy-api-apply"}},
	}
	writeStateObject(t, dir, operation.StagedCandidateKey, candidate)
	writeStateObject(t, dir, "status.json", statestore.Status{
		InstallID: "dep-1", BundleDigest: "sha256:next", ApprovalRequired: true,
		Steps: []statestore.StepStatus{{ID: "deploy-api-plan", Status: "finished"}},
	})

	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/approve", strings.NewReader(`{"bundle_digest":"sha256:next"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(operation.CandidateApprovalKey("sha256:next"))))
	require.NoError(t, err)
}

func TestPortalApprovesTerraformSandboxPlanSeparately(t *testing.T) {
	p, dir := testPortal(t)
	writeStateObject(t, dir, operation.CandidateKey, operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:bundle",
		Bundle:  operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:next"},
		Changes: []operation.BundleChange{{Kind: operation.BundleContentKindSandbox, Name: "terraform", Change: operation.BundleChangeChanged, PlanStepID: "sandbox-plan", ApplyStepID: "sandbox-apply"}},
	})
	writeStateObject(t, dir, "status.json", statestore.Status{
		InstallID: "dep-1", BundleDigest: "sha256:next", ApprovalRequired: true, ApprovalPhase: "sandbox",
		Steps: []statestore.StepStatus{{ID: "sandbox-plan", Status: "finished"}},
	})
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/approve", strings.NewReader(`{"bundle_digest":"sha256:next"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(operation.CandidateSandboxApprovalKey("sha256:next"))))
	require.NoError(t, err)
}

func TestPortalRejectsStaleOrUnplannedBundleCandidateApproval(t *testing.T) {
	p, dir := testPortal(t)
	writeStateObject(t, dir, operation.CandidateKey, operation.BundleCandidate{
		SchemaVersion: operation.SchemaVersion, PreviousDigest: "sha256:bundle",
		Bundle:  operation.BundleInfo{SchemaVersion: operation.SchemaVersion, DeploymentID: "dep-1", BundleDigest: "sha256:next"},
		Changes: []operation.BundleChange{{Kind: operation.BundleContentKindComponent, Name: "api", Change: operation.BundleChangeChanged, PlanStepID: "deploy-api-plan"}},
	})
	writeStateObject(t, dir, "status.json", statestore.Status{InstallID: "dep-1", BundleDigest: "sha256:next", ApprovalRequired: true})

	for _, digest := range []string{"sha256:stale", "sha256:next"} {
		request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/bundle-candidate/approve", strings.NewReader(`{"bundle_digest":"`+digest+`"}`))
		request.Header.Set("X-CSRF-Token", "secret")
		response := httptest.NewRecorder()
		p.handler().ServeHTTP(response, request)
		require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
	}
}

func TestPortalReportAndStackOutputs(t *testing.T) {
	dir := t.TempDir()
	stackDir := t.TempDir()
	writeStateObject(t, dir, "report.json", map[string]any{"install_id": "inst-1", "run_id": "run-1", "status": "finished"})
	writeStateObject(t, stackDir, "stack-outputs/outputs.json", map[string]any{"vpc_id": "vpc-123"})
	p, err := newPortalServer(operationstate.NewLocal(dir), operationstate.NewLocal(stackDir), "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	h := p.handler()

	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/report", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"install_id":"inst-1","run_id":"run-1","status":"finished"}`, response.Body.String())

	response = httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/stack-outputs", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"vpc_id":"vpc-123"}`, response.Body.String())
}

func TestPortalReportAndStackOutputsAbsent(t *testing.T) {
	p, err := newPortalServer(operationstate.NewLocal(t.TempDir()), nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	h := p.handler()
	for _, path := range []string{"/api/report", "/api/stack-outputs"} {
		response := httptest.NewRecorder()
		h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234"+path, nil))
		require.Equal(t, http.StatusNotFound, response.Code, path)
	}
}

func TestPortalEmbedsSPAAndCSRFToken(t *testing.T) {
	p, _ := testPortal(t)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/runs", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `content="secret"`)
	require.NotContains(t, response.Body.String(), "{{CSRF_TOKEN}}")
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
}

func TestPortalServesRuntimeBranding(t *testing.T) {
	p, _ := testPortal(t)
	p.branding = portalBranding{
		Name:         "Acme cloud",
		LogoURL:      "https://assets.example.com/logo.svg",
		FaviconURL:   "/favicon.svg",
		PrimaryColor: "#123456",
		SupportURL:   "https://support.example.com",
	}
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/branding", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"name":"Acme cloud","logo_url":"https://assets.example.com/logo.svg","favicon_url":"/favicon.svg","primary_color":"#123456","support_url":"https://support.example.com"}`, response.Body.String())
}

type failingPutState struct {
	operationstate.State
}

func (f failingPutState) PutIfAbsent(context.Context, string, []byte) error {
	return operationstate.ErrObjectExists
}

func TestPortalDispatchUsesConditionalWrite(t *testing.T) {
	p, _ := testPortal(t)
	p.controlStore = failingPutState{State: p.controlStore}
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/dispatch", strings.NewReader(`{"ref_id":"restart-api"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), operationstate.ErrObjectExists.Error())
}

func TestPortalNamespacedStateOwnership(t *testing.T) {
	ctx := context.Background()
	base := operationstate.NewLocal(t.TempDir())
	runner := operationstate.WithPrefix(base, operationstate.RunnerNamespace)
	control := operationstate.WithPrefix(base, operationstate.ControlNamespace)
	read := operationstate.ReadOverlay(runner, control, operationstate.Legacy(base))
	catalog, err := json.Marshal(operation.Catalog{
		SchemaVersion: operation.SchemaVersion,
		DeploymentID:  "dep-1",
		BundleDigest:  "sha256:bundle",
		Refs:          []operation.CatalogRef{{ID: "restart-api", Kind: operation.RefKindAction, Name: "Restart API"}},
	})
	require.NoError(t, err)
	require.NoError(t, runner.Put(ctx, operation.CatalogKey, catalog))
	p, err := newPortalServer(read, nil, "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	p.controlStore = control

	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/dispatch", strings.NewReader(`{"ref_id":"restart-api"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	controlKeys, err := control.List(ctx, operation.RequestsPrefix)
	require.NoError(t, err)
	require.Len(t, controlKeys, 1)
	runnerKeys, err := runner.List(ctx, operation.RequestsPrefix)
	require.NoError(t, err)
	require.Empty(t, runnerKeys)
}
