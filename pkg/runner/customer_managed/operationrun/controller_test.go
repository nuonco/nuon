package operationrun

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

type fakeObject struct {
	body []byte
	etag string
}
type fakeS3 struct {
	mu      sync.Mutex
	objects map[string]fakeObject
	seq     int
}

func newFakeS3() *fakeS3 { return &fakeS3{objects: map[string]fakeObject{}} }
func (f *fakeS3) GetObject(_ context.Context, in *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	o, ok := f.objects[aws.ToString(in.Key)]
	if !ok {
		return nil, &smithy.GenericAPIError{Code: "NoSuchKey"}
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(string(o.body))), ETag: aws.String(o.etag)}, nil
}
func (f *fakeS3) ListObjectsV2(_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listOutput(f.objects, aws.ToString(in.Prefix)), nil
}
func listOutput(objects map[string]fakeObject, prefix string) *s3.ListObjectsV2Output {
	out := &s3.ListObjectsV2Output{}
	for key := range objects {
		if strings.HasPrefix(key, prefix) {
			k := key
			out.Contents = append(out.Contents, types.Object{Key: &k})
		}
	}
	return out
}
func (f *fakeS3) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := aws.ToString(in.Key)
	old, exists := f.objects[key]
	if aws.ToString(in.IfNoneMatch) == "*" && exists {
		return nil, &smithy.GenericAPIError{Code: "PreconditionFailed"}
	}
	if match := aws.ToString(in.IfMatch); match != "" && (!exists || old.etag != match) {
		return nil, &smithy.GenericAPIError{Code: "PreconditionFailed"}
	}
	b, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	f.seq++
	etag := fmt.Sprintf("e%d", f.seq)
	f.objects[key] = fakeObject{b, etag}
	return &s3.PutObjectOutput{ETag: &etag}, nil
}

type fakeExecutor struct {
	mu   sync.Mutex
	runs []string
	busy bool
}

func (f *fakeExecutor) Busy() bool { f.mu.Lock(); defer f.mu.Unlock(); return f.busy }
func (f *fakeExecutor) Execute(_ context.Context, req operation.Request, runID string) (*operation.RunStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.runs = append(f.runs, runID)
	return &operation.RunStatus{RunID: runID, DispatchID: req.DispatchID, Status: operation.RunStatusFinished}, nil
}

func testDispatcher(t *testing.T) (*dispatcher, *Mailbox, *fakeS3, *fakeExecutor) {
	t.Helper()
	s := newFakeS3()
	m := NewMailbox(s, "b", "p")
	ex := &fakeExecutor{}
	env := &customermanaged.Envelope{InstallID: "dep", Actions: []customermanaged.ActionTemplate{{ID: "act", Name: "Action"}}}
	d := &dispatcher{mailbox: m, envelope: env, digest: "digest", deploymentID: "dep", owner: "owner", executor: ex, logger: zap.NewNop(), now: time.Now}
	return d, m, s, ex
}
func validRequest(id string) operation.Request {
	return operation.Request{SchemaVersion: 1, DeploymentID: "dep", BundleDigest: "digest", RefID: "act", DispatchID: id, Source: operation.SourceCLI, CreatedAt: time.Now()}
}

func TestNamespacedMailboxSeparatesControlAndRunnerWrites(t *testing.T) {
	ctx := context.Background()
	s := newFakeS3()
	m := NewNamespacedMailbox(s, "b", "state")
	req := validRequest("dispatch-1")

	require.NoError(t, m.PutRequest(ctx, req))
	require.NoError(t, m.ClaimNew(ctx, operation.Claim{DispatchID: req.DispatchID}))
	require.NoError(t, m.PutReceipt(ctx, operation.Receipt{DispatchID: req.DispatchID}))
	require.NoError(t, m.PutCatalog(ctx, operation.Catalog{DeploymentID: "dep"}))

	s.mu.Lock()
	defer s.mu.Unlock()
	require.Contains(t, s.objects, "state/control/dispatch/requests/dispatch-1.json")
	require.Contains(t, s.objects, "state/runner/dispatch/claims/dispatch-1.json")
	require.Contains(t, s.objects, "state/runner/dispatch/receipts/dispatch-1.json")
	require.Contains(t, s.objects, "state/runner/operations/catalog.json")
}

func TestNamespacedMailboxHonorsAndTakesOverLegacyClaim(t *testing.T) {
	ctx := context.Background()
	s := newFakeS3()
	legacy := NewMailbox(s, "b", "state")
	claim := operation.Claim{DispatchID: "dispatch-1", Owner: "old"}
	require.NoError(t, legacy.ClaimNew(ctx, claim))
	m := NewNamespacedMailbox(s, "b", "state")

	require.ErrorIs(t, m.ClaimNew(ctx, claim), ErrAlreadyClaimed)
	got, etag, found, err := m.GetClaim(ctx, claim.DispatchID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "old", got.Owner)
	require.True(t, strings.HasPrefix(etag, legacyETagPrefix))

	claim.Owner = "new"
	require.NoError(t, m.TakeOverClaim(ctx, claim, etag))
	_, namespaced := s.objects["state/runner/"+operation.ClaimKey(claim.DispatchID)]
	require.False(t, namespaced)
	var persisted operation.Claim
	require.NoError(t, json.Unmarshal(s.objects["state/"+operation.ClaimKey(claim.DispatchID)].body, &persisted))
	require.Equal(t, "new", persisted.Owner)
}

func TestDispatchAndDuplicate(t *testing.T) {
	d, m, _, ex := testDispatcher(t)
	req := validRequest("dispatch-1")
	require.NoError(t, m.PutRequest(context.Background(), req))
	p := Poller{dispatcher: d, interval: time.Hour}
	p.poll(context.Background())
	p.poll(context.Background())
	receipt, found, err := m.GetReceipt(context.Background(), req.DispatchID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, operation.ReceiptStatusFinished, receipt.Status)
	require.Len(t, ex.runs, 1)
	require.Equal(t, ex.runs[0], receipt.RunID)
}
func TestRejectedRequests(t *testing.T) {
	for _, mutate := range []func(*operation.Request){func(r *operation.Request) { r.BundleDigest = "wrong" }, func(r *operation.Request) { r.RefID = "unknown" }} {
		d, m, _, ex := testDispatcher(t)
		req := validRequest("bad")
		mutate(&req)
		require.NoError(t, d.handle(context.Background(), req.DispatchID, req))
		receipt, found, err := m.GetReceipt(context.Background(), req.DispatchID)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, operation.ReceiptStatusRejected, receipt.Status)
		require.NotEmpty(t, receipt.Reason)
		require.Empty(t, ex.runs)
	}
}

func TestBundlePlanAcceptsCandidateDigestAndUsesRequestedRunID(t *testing.T) {
	d, m, _, ex := testDispatcher(t)
	req := validRequest("candidate-plan")
	req.RefKind = operation.RefKindBundlePlan
	req.RunID = "portal-run"
	req.BundleDigest = "sha256:candidate"
	req.CandidateArchiveKey = "candidate.tar.zst"
	req.CandidateRecordKey = "candidate.json"
	req.PlanStepIDs = []string{"sandbox-plan"}

	require.NoError(t, d.handle(context.Background(), req.DispatchID, req))
	receipt, found, err := m.GetReceipt(context.Background(), req.DispatchID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, operation.ReceiptStatusFinished, receipt.Status)
	require.Equal(t, req.RunID, receipt.RunID)
	require.Equal(t, []string{req.RunID}, ex.runs)
}

type staticCandidateLoader struct {
	envelope *customermanaged.Envelope
}

func (l staticCandidateLoader) Load(context.Context, operation.Request) (*CandidateBundle, error) {
	return &CandidateBundle{Envelope: l.envelope}, nil
}

func TestBundlePlanRejectsNonPlanOperations(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	require.NoError(t, err)
	envelope := &customermanaged.Envelope{InstallID: "dep", Steps: []customermanaged.Step{{ID: "sandbox-apply", JobOperation: "apply-plan"}}}
	executor := &Executor{envelope: &customermanaged.Envelope{InstallID: "dep"}, store: store, loader: staticCandidateLoader{envelope: envelope}}
	req := operation.Request{RefKind: operation.RefKindBundlePlan, DeploymentID: "dep", BundleDigest: "sha256:candidate", PlanStepIDs: []string{"sandbox-apply"}}

	run, err := executor.Execute(context.Background(), req, "run")
	require.Nil(t, run)
	require.ErrorContains(t, err, "is not create-apply-plan")
}
func TestClaimTakeoverAndLiveSkip(t *testing.T) {
	d, m, _, ex := testDispatcher(t)
	req := validRequest("takeover")
	require.NoError(t, m.ClaimNew(context.Background(), operation.Claim{DispatchID: req.DispatchID, Owner: "old", RunID: "old", Attempt: 1, ExpiresAt: time.Now().Add(-claimGrace - time.Minute)}))
	require.NoError(t, d.handle(context.Background(), req.DispatchID, req))
	claim, _, _, err := m.GetClaim(context.Background(), req.DispatchID)
	require.NoError(t, err)
	require.Equal(t, 2, claim.Attempt)
	require.Len(t, ex.runs, 1)
	d2, m2, _, ex2 := testDispatcher(t)
	req2 := validRequest("live")
	require.NoError(t, m2.ClaimNew(context.Background(), operation.Claim{DispatchID: req2.DispatchID, Owner: "other", RunID: "live", Attempt: 1, ExpiresAt: time.Now().Add(time.Minute)}))
	require.NoError(t, d2.handle(context.Background(), req2.DispatchID, req2))
	require.Empty(t, ex2.runs)
}
func TestSchedulerDuplicateAndBusy(t *testing.T) {
	d, _, s, ex := testDispatcher(t)
	tick := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	local := map[string][]byte{}
	scheduler := Scheduler{dispatcher: d, writeLocal: func(k string, b []byte) error { local[k] = b; return nil }}
	action := d.envelope.Actions[0]
	scheduler.tick(context.Background(), action, tick)
	scheduler.tick(context.Background(), action, tick)
	id := operation.OccurrenceID("dep", "digest", "act", tick)
	_, ok := s.objects["p/"+operation.RequestKey(id)]
	require.True(t, ok)
	require.Len(t, ex.runs, 1)
	ex.busy = true
	scheduler.tick(context.Background(), action, tick.Add(time.Minute))
	require.NotEmpty(t, local[operation.ScheduleCursorKey("act")])
	var cursor operation.ScheduleCursor
	require.NoError(t, json.Unmarshal(local[operation.ScheduleCursorKey("act")], &cursor))
	require.Equal(t, 1, cursor.Skipped)
}
func TestCatalogPublication(t *testing.T) {
	d, m, _, ex := testDispatcher(t)
	env := d.envelope
	env.Actions[0].CronSchedule = "0 * * * *"
	env.Drift = []customermanaged.DriftTemplate{{ID: "drift", ComponentName: "api"}}
	env.Runbooks = []customermanaged.RunbookTemplate{{ID: "book", Name: "Book", Steps: []customermanaged.RunbookStep{{Kind: customermanaged.RunbookStepKindHealthGate}}}}
	var local []byte
	c, err := NewController(ControllerConfig{Mailbox: m, Envelope: env, Digest: "digest", DeploymentID: "dep", Owner: "owner", Executor: ex, WriteLocal: func(_ string, b []byte) error { local = b; return nil }})
	require.NoError(t, err)
	require.NoError(t, c.publishCatalog(context.Background()))
	var catalog operation.Catalog
	require.NoError(t, json.Unmarshal(local, &catalog))
	require.Len(t, catalog.Refs, 3)
	require.Equal(t, "api", catalog.Refs[1].Component)
	require.Equal(t, 1, catalog.Refs[2].Steps)
}

func TestClassifyDriftIgnoresNoopsAndNullEmptyNormalization(t *testing.T) {
	plan := []byte(`{
		"format_version":"1.2",
		"terraform_version":"1.11.3",
		"resource_changes":[{"change":{"actions":["no-op"]}}],
		"output_changes":{"arn":{"actions":["no-op"]}},
		"resource_drift":[{"change":{"actions":["update"],"before":{"tags":null},"after":{"tags":{}}}}]
	}`)
	result, err := ClassifyDrift(plan)
	require.NoError(t, err)
	require.False(t, result.Drifted)
	require.Zero(t, result.ResourceChanges)
	require.Zero(t, result.OutputChanges)
	require.Zero(t, result.ResourceDrift)

	plan = []byte(`{
		"format_version":"1.2",
		"terraform_version":"1.11.3",
		"resource_changes":[{"address":"resource.demo","change":{"actions":["update"]}}],
		"resource_drift":[{"address":"resource.demo","change":{"actions":["update"],"before":{"tags":{"demo":"changed"}},"after":{"tags":{}}}}]
	}`)
	result, err = ClassifyDrift(plan)
	require.NoError(t, err)
	require.True(t, result.Drifted)
	require.Equal(t, 1, result.ResourceChanges)
	require.Equal(t, 1, result.ResourceDrift)
}

func TestClassifyDriftResources(t *testing.T) {
	plan := []byte(`{
		"format_version":"1.2",
		"terraform_version":"1.11.3",
		"resource_changes":[
			{"address":"aws_acm_certificate.demo","change":{"actions":["update"]}},
			{"address":"aws_s3_bucket.new","change":{"actions":["create"]}},
			{"address":"aws_iam_role.replaced","change":{"actions":["delete","create"]}},
			{"address":"aws_vpc.untouched","change":{"actions":["no-op"]}}
		],
		"resource_drift":[
			{"address":"aws_acm_certificate.demo","change":{"actions":["update"]}}
		]
	}`)
	result, err := ClassifyDrift(plan)
	require.NoError(t, err)
	require.True(t, result.Drifted)
	require.False(t, result.ResourcesTruncated)
	require.Equal(t, []operation.DriftResourceChange{
		{Address: "aws_acm_certificate.demo", Action: "update", Drifted: true},
		{Address: "aws_s3_bucket.new", Action: "create"},
		{Address: "aws_iam_role.replaced", Action: "replace"},
		{Address: "aws_vpc.untouched", Action: "noop"},
	}, result.Resources)
}

func TestClassifyDriftResourcesTruncated(t *testing.T) {
	changes := make([]string, 0, operation.MaxDriftResources+10)
	for i := 0; i < operation.MaxDriftResources+10; i++ {
		changes = append(changes, fmt.Sprintf(`{"address":"aws_vpc.r%d","change":{"actions":["no-op"]}}`, i))
	}
	plan := []byte(`{
		"format_version":"1.2",
		"terraform_version":"1.11.3",
		"resource_changes":[` + strings.Join(changes, ",") + `]
	}`)
	result, err := ClassifyDrift(plan)
	require.NoError(t, err)
	require.True(t, result.ResourcesTruncated)
	require.Len(t, result.Resources, operation.MaxDriftResources)
}
