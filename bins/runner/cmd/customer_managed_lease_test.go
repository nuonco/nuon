package cmd

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
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeS3Object struct {
	body         []byte
	etag         string
	lastModified time.Time
}

// fakeLeaseS3 implements customerManagedS3Client with S3's conditional-write semantics
// (If-None-Match: * and If-Match: <etag>).
type fakeLeaseS3 struct {
	mu      sync.Mutex
	objects map[string]fakeS3Object
	seq     int
}

func newFakeLeaseS3() *fakeLeaseS3 {
	return &fakeLeaseS3{objects: map[string]fakeS3Object{}}
}

func (f *fakeLeaseS3) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	object, ok := f.objects[aws.ToString(input.Key)]
	if !ok {
		return nil, &smithy.GenericAPIError{Code: "NoSuchKey", Message: "not found"}
	}
	out := &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(string(object.body))), ETag: aws.String(object.etag)}
	if !object.lastModified.IsZero() {
		out.LastModified = aws.Time(object.lastModified)
	}
	return out, nil
}

func (f *fakeLeaseS3) ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return &s3.ListObjectsV2Output{}, nil
}

func (f *fakeLeaseS3) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := aws.ToString(input.Key)
	existing, exists := f.objects[key]
	if aws.ToString(input.IfNoneMatch) == "*" && exists {
		return nil, &smithy.GenericAPIError{Code: "PreconditionFailed", Message: "object exists"}
	}
	if ifMatch := aws.ToString(input.IfMatch); ifMatch != "" {
		if !exists || existing.etag != ifMatch {
			return nil, &smithy.GenericAPIError{Code: "PreconditionFailed", Message: "etag mismatch"}
		}
	}
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	f.seq++
	object := fakeS3Object{body: body, etag: fmt.Sprintf("etag-%d", f.seq), lastModified: time.Now()}
	f.objects[key] = object
	return &s3.PutObjectOutput{ETag: aws.String(object.etag)}, nil
}

func (f *fakeLeaseS3) backdate(t *testing.T, key string, d time.Duration) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	object, ok := f.objects[key]
	require.True(t, ok, "lease object missing")
	object.lastModified = object.lastModified.Add(-d)
	f.objects[key] = object
}

func (f *fakeLeaseS3) record(t *testing.T, key string) customerManagedLeaseRecord {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	object, ok := f.objects[key]
	require.True(t, ok, "lease object missing")
	var record customerManagedLeaseRecord
	require.NoError(t, json.Unmarshal(object.body, &record))
	return record
}

func TestCustomerManagedLeaseAcquireAbsent(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Acquire(context.Background()))
	record := fake.record(t, "state/LEASE")
	require.Equal(t, "runner-a", record.Owner)
	require.True(t, record.ExpiresAt.After(time.Now()))
}

func TestCustomerManagedLeaseAcquireContention(t *testing.T) {
	fake := newFakeLeaseS3()
	holder := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, holder.Acquire(context.Background()))

	contender := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-b", customerManagedLeaseTTL)
	err := contender.Acquire(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "locked by runner")
	require.Equal(t, "runner-a", fake.record(t, "state/LEASE").Owner)
}

// fastWatch shrinks the ETag-stability watch a contender performs before
// taking over an expired-looking lease, so tests do not sleep for real.
func fastWatch(l *customerManagedLease) *customerManagedLease {
	l.interval = 20 * time.Millisecond
	l.renewTimeout = 10 * time.Millisecond
	l.takeoverWatch = 30 * time.Millisecond
	return l
}

func TestCustomerManagedLeaseTakeoverExpired(t *testing.T) {
	fake := newFakeLeaseS3()
	holder := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", -time.Minute)
	require.NoError(t, holder.Acquire(context.Background()))
	fake.backdate(t, "state/LEASE", 2*customerManagedLeaseTTL)

	successor := fastWatch(newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-b", customerManagedLeaseTTL))
	require.NoError(t, successor.Acquire(context.Background()))
	require.Equal(t, "runner-b", fake.record(t, "state/LEASE").Owner)

	require.Error(t, holder.Renew(context.Background()), "stale holder must not renew over the successor")
}

func TestCustomerManagedLeaseTakeoverAbortedWhenHolderRenews(t *testing.T) {
	fake := newFakeLeaseS3()
	holder := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", -time.Minute)
	require.NoError(t, holder.Acquire(context.Background()))
	fake.backdate(t, "state/LEASE", 2*customerManagedLeaseTTL)

	successor := fastWatch(newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-b", customerManagedLeaseTTL))
	successor.takeoverWatch = 100 * time.Millisecond

	renewed := make(chan error, 1)
	go func() {
		time.Sleep(30 * time.Millisecond)
		renewed <- holder.Renew(context.Background())
	}()

	err := successor.Acquire(context.Background())
	require.Error(t, err, "a holder that renews during the stability watch is alive; the contender's clock was wrong")
	require.Contains(t, err.Error(), "still being renewed")
	require.NoError(t, <-renewed)
	require.Equal(t, "runner-a", fake.record(t, "state/LEASE").Owner)
}

func TestCustomerManagedLeaseNoTakeoverWithFreshServerWrite(t *testing.T) {
	fake := newFakeLeaseS3()
	holder := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", -time.Minute)
	require.NoError(t, holder.Acquire(context.Background()))

	successor := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-b", customerManagedLeaseTTL)
	err := successor.Acquire(context.Background())
	require.Error(t, err, "a record just written to S3 is not expired even if the holder's clock produced a past ExpiresAt")
	require.Equal(t, "runner-a", fake.record(t, "state/LEASE").Owner)
}

func TestCustomerManagedLeaseReacquireOwn(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Acquire(context.Background()))
	require.NoError(t, lease.Acquire(context.Background()), "an owner may re-acquire its own live lease")
}

func TestCustomerManagedLeaseRenewExtends(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Acquire(context.Background()))
	first := fake.record(t, "state/LEASE")
	require.NoError(t, lease.Renew(context.Background()))
	second := fake.record(t, "state/LEASE")
	require.False(t, second.ExpiresAt.Before(first.ExpiresAt))
}

func TestCustomerManagedLeaseRenewFailsWhenTakenOver(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Acquire(context.Background()))

	usurper := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-b", customerManagedLeaseTTL)
	fake.mu.Lock()
	fake.objects["state/LEASE"] = fakeS3Object{
		body: mustJSON(t, customerManagedLeaseRecord{Owner: "runner-b", ExpiresAt: time.Now().Add(time.Minute)}),
		etag: "etag-usurped",
	}
	fake.mu.Unlock()
	_ = usurper

	err := lease.Renew(context.Background())
	require.Error(t, err)
	require.True(t, isS3ConditionFailed(err))
}

func TestCustomerManagedLeaseReleaseExpiresInPlace(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Acquire(context.Background()))
	require.NoError(t, lease.Release(context.Background()))

	record := fake.record(t, "state/LEASE")
	require.True(t, record.Released)
	require.False(t, record.ExpiresAt.After(time.Now()))

	successor := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-b", customerManagedLeaseTTL)
	require.NoError(t, successor.Acquire(context.Background()), "released lease should be immediately acquirable despite a fresh LastModified")
}

func TestCustomerManagedLeaseStaleReleaseDoesNotDisturbSuccessor(t *testing.T) {
	fake := newFakeLeaseS3()
	holder := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", -time.Minute)
	require.NoError(t, holder.Acquire(context.Background()))
	fake.backdate(t, "state/LEASE", 2*customerManagedLeaseTTL)

	successor := fastWatch(newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-b", customerManagedLeaseTTL))
	require.NoError(t, successor.Acquire(context.Background()))

	err := holder.Release(context.Background())
	require.Error(t, err)
	require.True(t, isS3ConditionFailed(err))
	record := fake.record(t, "state/LEASE")
	require.Equal(t, "runner-b", record.Owner)
	require.False(t, record.Released)
}

func TestCustomerManagedLeaseReleaseWithoutAcquireIsNoop(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Release(context.Background()))
	fake.mu.Lock()
	defer fake.mu.Unlock()
	require.Empty(t, fake.objects)
}

func TestLeaseExpired(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-time.Minute)
	future := now.Add(time.Minute)

	require.True(t, leaseExpired(customerManagedLeaseRecord{Released: true, ExpiresAt: future}, &now, customerManagedLeaseTTL), "explicit release always permits takeover")
	require.False(t, leaseExpired(customerManagedLeaseRecord{ExpiresAt: future}, nil, customerManagedLeaseTTL))
	require.True(t, leaseExpired(customerManagedLeaseRecord{ExpiresAt: past}, nil, customerManagedLeaseTTL))
	require.False(t, leaseExpired(customerManagedLeaseRecord{ExpiresAt: past}, &now, customerManagedLeaseTTL), "fresh server write extends a stale holder-clock expiry")
	stale := now.Add(-2 * customerManagedLeaseTTL)
	require.True(t, leaseExpired(customerManagedLeaseRecord{ExpiresAt: past}, &stale, customerManagedLeaseTTL))
	require.False(t, leaseExpired(customerManagedLeaseRecord{ExpiresAt: future}, &stale, customerManagedLeaseTTL), "holder expiry wins when it is later than server write + TTL")
}

// blockingS3 blocks every PutObject until its context is done, simulating an
// S3 outage or partition during renewal.
type blockingS3 struct {
	customerManagedS3Client
}

func (b *blockingS3) PutObject(ctx context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func waitForLost(t *testing.T, lost chan struct{}, msg string) {
	t.Helper()
	select {
	case <-lost:
	case <-time.After(5 * time.Second):
		t.Fatal(msg)
	}
}

func TestCustomerManagedRenewLoopLostOnConditionFailure(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Acquire(context.Background()))
	lease.interval = 10 * time.Millisecond

	fake.mu.Lock()
	fake.objects["state/LEASE"] = fakeS3Object{
		body:         mustJSON(t, customerManagedLeaseRecord{Owner: "runner-b", ExpiresAt: time.Now().Add(time.Minute)}),
		etag:         "etag-usurped",
		lastModified: time.Now(),
	}
	fake.mu.Unlock()

	lost := make(chan struct{})
	go lease.renewLoop(context.Background(), zap.NewNop(), func() { close(lost) })
	waitForLost(t, lost, "renewLoop did not report loss after takeover")
}

func TestCustomerManagedRenewLoopLostAfterBlockedRenewals(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", 100*time.Millisecond)
	require.NoError(t, lease.Acquire(context.Background()))
	lease.client = &blockingS3{}
	lease.interval = 20 * time.Millisecond
	lease.renewTimeout = 10 * time.Millisecond

	lost := make(chan struct{})
	go lease.renewLoop(context.Background(), zap.NewNop(), func() { close(lost) })
	waitForLost(t, lost, "renewLoop did not report loss while renewals were blocked")
}

func TestCustomerManagedRenewLoopStopsOnCancel(t *testing.T) {
	fake := newFakeLeaseS3()
	lease := newCustomerManagedLease(fake, "bucket", "state/LEASE", "runner-a", customerManagedLeaseTTL)
	require.NoError(t, lease.Acquire(context.Background()))
	lease.interval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		lease.renewLoop(ctx, zap.NewNop(), func() { t.Error("onLost must not fire on clean cancellation") })
	}()
	time.Sleep(30 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("renewLoop did not stop after cancellation")
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
