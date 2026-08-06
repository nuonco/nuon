package day2run

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

	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
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
func (f *fakeExecutor) Execute(_ context.Context, req day2.Request, runID string) (*day2.RunStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.runs = append(f.runs, runID)
	return &day2.RunStatus{RunID: runID, DispatchID: req.DispatchID, Status: day2.RunStatusFinished}, nil
}

func testDispatcher(t *testing.T) (*dispatcher, *Mailbox, *fakeS3, *fakeExecutor) {
	t.Helper()
	s := newFakeS3()
	m := NewMailbox(s, "b", "p")
	ex := &fakeExecutor{}
	env := &airgap.Envelope{InstallID: "dep", Actions: []airgap.ActionTemplate{{ID: "act", Name: "Action"}}}
	d := &dispatcher{mailbox: m, envelope: env, digest: "digest", deploymentID: "dep", owner: "owner", executor: ex, logger: zap.NewNop(), now: time.Now}
	return d, m, s, ex
}
func validRequest(id string) day2.Request {
	return day2.Request{SchemaVersion: 1, DeploymentID: "dep", BundleDigest: "digest", RefID: "act", DispatchID: id, Source: day2.SourceCLI, CreatedAt: time.Now()}
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
	require.Equal(t, day2.ReceiptStatusFinished, receipt.Status)
	require.Len(t, ex.runs, 1)
	require.Equal(t, ex.runs[0], receipt.RunID)
}
func TestRejectedRequests(t *testing.T) {
	for _, mutate := range []func(*day2.Request){func(r *day2.Request) { r.BundleDigest = "wrong" }, func(r *day2.Request) { r.RefID = "unknown" }} {
		d, m, _, ex := testDispatcher(t)
		req := validRequest("bad")
		mutate(&req)
		require.NoError(t, d.handle(context.Background(), req.DispatchID, req))
		receipt, found, err := m.GetReceipt(context.Background(), req.DispatchID)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, day2.ReceiptStatusRejected, receipt.Status)
		require.NotEmpty(t, receipt.Reason)
		require.Empty(t, ex.runs)
	}
}
func TestClaimTakeoverAndLiveSkip(t *testing.T) {
	d, m, _, ex := testDispatcher(t)
	req := validRequest("takeover")
	require.NoError(t, m.ClaimNew(context.Background(), day2.Claim{DispatchID: req.DispatchID, Owner: "old", RunID: "old", Attempt: 1, ExpiresAt: time.Now().Add(-claimGrace - time.Minute)}))
	require.NoError(t, d.handle(context.Background(), req.DispatchID, req))
	claim, _, _, err := m.GetClaim(context.Background(), req.DispatchID)
	require.NoError(t, err)
	require.Equal(t, 2, claim.Attempt)
	require.Len(t, ex.runs, 1)
	d2, m2, _, ex2 := testDispatcher(t)
	req2 := validRequest("live")
	require.NoError(t, m2.ClaimNew(context.Background(), day2.Claim{DispatchID: req2.DispatchID, Owner: "other", RunID: "live", Attempt: 1, ExpiresAt: time.Now().Add(time.Minute)}))
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
	id := day2.OccurrenceID("dep", "digest", "act", tick)
	_, ok := s.objects["p/"+day2.RequestKey(id)]
	require.True(t, ok)
	require.Len(t, ex.runs, 1)
	ex.busy = true
	scheduler.tick(context.Background(), action, tick.Add(time.Minute))
	require.NotEmpty(t, local[day2.ScheduleCursorKey("act")])
	var cursor day2.ScheduleCursor
	require.NoError(t, json.Unmarshal(local[day2.ScheduleCursorKey("act")], &cursor))
	require.Equal(t, 1, cursor.Skipped)
}
func TestCatalogPublication(t *testing.T) {
	d, m, _, ex := testDispatcher(t)
	env := d.envelope
	env.Actions[0].CronSchedule = "0 * * * *"
	env.Drift = []airgap.DriftTemplate{{ID: "drift", ComponentName: "api"}}
	env.Runbooks = []airgap.RunbookTemplate{{ID: "book", Name: "Book", Steps: []airgap.RunbookStep{{Kind: airgap.RunbookStepKindHealthGate}}}}
	var local []byte
	c, err := NewController(ControllerConfig{Mailbox: m, Envelope: env, Digest: "digest", DeploymentID: "dep", Owner: "owner", Executor: ex, WriteLocal: func(_ string, b []byte) error { local = b; return nil }})
	require.NoError(t, err)
	require.NoError(t, c.publishCatalog(context.Background()))
	var catalog day2.Catalog
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
