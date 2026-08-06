package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	airgapLeaseObject       = "LEASE"
	airgapLeaseTTL          = 90 * time.Second
	airgapLeaseInterval     = 30 * time.Second
	airgapLeaseRenewTimeout = 15 * time.Second
	// airgapLeaseFailStopGrace is how long a runner that lost its lease may
	// spend shutting down before the process force-exits. Loss is declared
	// ttl-interval-renewTimeout (45s) after the last confirmed write and a
	// successor's takeover additionally waits out an interval+renewTimeout
	// ETag-stability watch, so grace ends before a takeover can complete.
	airgapLeaseFailStopGrace = 60 * time.Second
)

func airgapLeaseOwner() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "runner"
	}
	return host + "-" + uuid.NewString()[:8]
}

type airgapLeaseRecord struct {
	Owner      string    `json:"owner"`
	AcquiredAt time.Time `json:"acquired_at"`
	RenewedAt  time.Time `json:"renewed_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Released   bool      `json:"released,omitempty"`
}

// airgapLease is a deployment-scoped mutual-exclusion lease held as one S3
// object under the state prefix, guarded by S3 conditional writes
// (If-None-Match on create, If-Match on every renewal). It keeps a second
// runner — a replaced ASG instance whose predecessor is still alive, or an
// operator running the CLI against the same state — from executing the same
// deployment concurrently.
type airgapLease struct {
	client        airgapS3Client
	bucket        string
	key           string
	owner         string
	ttl           time.Duration
	interval      time.Duration
	renewTimeout  time.Duration
	takeoverWatch time.Duration

	mu         sync.Mutex
	etag       string
	writeStart time.Time
}

func newAirgapLease(client airgapS3Client, bucket, key, owner string, ttl time.Duration) *airgapLease {
	return &airgapLease{
		client:       client,
		bucket:       bucket,
		key:          key,
		owner:        owner,
		ttl:          ttl,
		interval:     airgapLeaseInterval,
		renewTimeout: airgapLeaseRenewTimeout,
		// The watch must outlast a still-running holder's worst-case loss
		// detection (up to ttl after its last confirmed write) plus its
		// fail-stop grace plus one in-flight renewal, so any live holder
		// either renews (aborting the takeover) or force-exits before the
		// takeover can complete. Tolerates ~105s of contender clock skew.
		takeoverWatch: airgapLeaseTTL + airgapLeaseFailStopGrace + airgapLeaseRenewTimeout,
	}
}

// Acquire takes the lease if it is absent, expired, or already ours. A live
// lease held by another owner is an error: the caller should exit and let its
// supervisor retry after the holder dies or lets the lease lapse.
func (l *airgapLease) Acquire(ctx context.Context) error {
	existing, etag, lastModified, found, err := l.read(ctx)
	if err != nil {
		return err
	}
	if !found {
		if err := l.put(ctx, "*", ""); err != nil {
			if isS3ConditionFailed(err) {
				return fmt.Errorf("deployment lease was acquired by another runner concurrently")
			}
			return fmt.Errorf("create deployment lease: %w", err)
		}
		return nil
	}
	if existing.Owner != l.owner {
		if !leaseExpired(existing, lastModified, l.ttl) {
			return fmt.Errorf("deployment is locked by runner %q until %s; a second runner must not execute the same deployment", existing.Owner, existing.ExpiresAt.Format(time.RFC3339))
		}
		// An expired-looking record only proves this process's clock says so.
		// A live holder renews every interval, so an ETag that stays put for
		// the takeover watch is clock-independent proof the holder is dead.
		// An explicit release needs no confirmation.
		if !existing.Released {
			if err := l.confirmHolderDead(ctx, etag); err != nil {
				return err
			}
		}
	}
	if err := l.put(ctx, "", etag); err != nil {
		if isS3ConditionFailed(err) {
			return fmt.Errorf("deployment lease was taken over by another runner concurrently")
		}
		return fmt.Errorf("take over deployment lease: %w", err)
	}
	return nil
}

func (l *airgapLease) read(ctx context.Context) (airgapLeaseRecord, string, *time.Time, bool, error) {
	getCtx, cancel := context.WithTimeout(ctx, l.renewTimeout)
	defer cancel()
	out, err := l.client.GetObject(getCtx, &s3.GetObjectInput{Bucket: &l.bucket, Key: &l.key})
	if err != nil {
		if isS3NotFound(err) {
			return airgapLeaseRecord{}, "", nil, false, nil
		}
		return airgapLeaseRecord{}, "", nil, false, fmt.Errorf("read deployment lease: %w", err)
	}
	body, readErr := io.ReadAll(out.Body)
	closeErr := out.Body.Close()
	if readErr != nil {
		return airgapLeaseRecord{}, "", nil, false, fmt.Errorf("read deployment lease: %w", readErr)
	}
	if closeErr != nil {
		return airgapLeaseRecord{}, "", nil, false, closeErr
	}
	var record airgapLeaseRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return airgapLeaseRecord{}, "", nil, false, fmt.Errorf("decode deployment lease: %w", err)
	}
	return record, aws.ToString(out.ETag), out.LastModified, true, nil
}

// confirmHolderDead watches the lease object for takeoverWatch. If the ETag
// changes the holder is alive and this contender's clock was fast; if it
// stays put, no live holder exists regardless of anyone's clock, and any
// holder that was merely partitioned has force-exited by now.
func (l *airgapLease) confirmHolderDead(ctx context.Context, etag string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(l.takeoverWatch):
	}
	current, currentETag, _, found, err := l.read(ctx)
	if err != nil {
		return err
	}
	if found && currentETag != etag {
		return fmt.Errorf("deployment lease is still being renewed by runner %q; a second runner must not execute the same deployment", current.Owner)
	}
	return nil
}

// Renew extends the lease with a conditional write against the ETag of our
// last write, so a takeover by another runner surfaces as a condition failure.
func (l *airgapLease) Renew(ctx context.Context) error {
	etag := l.currentETag()
	if etag == "" {
		return fmt.Errorf("lease was never acquired")
	}
	return l.put(ctx, "", etag)
}

// Release expires the lease in place (best-effort) so a successor can acquire
// immediately instead of waiting out the TTL.
func (l *airgapLease) Release(ctx context.Context) error {
	etag := l.currentETag()
	if etag == "" {
		return nil
	}
	record := airgapLeaseRecord{Owner: l.owner, RenewedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC(), Released: true}
	return l.putRecord(ctx, record, "", etag)
}

// leaseExpired decides takeover eligibility. An explicit release always
// permits takeover. Otherwise the holder-written ExpiresAt is extended by the
// server-side write time plus TTL when available, so a holder with a slow
// clock is not evicted early: clock skew delays takeover instead of allowing
// two owners.
func leaseExpired(record airgapLeaseRecord, lastModified *time.Time, ttl time.Duration) bool {
	if record.Released {
		return true
	}
	expiry := record.ExpiresAt
	if lastModified != nil {
		if serverExpiry := lastModified.Add(ttl); serverExpiry.After(expiry) {
			expiry = serverExpiry
		}
	}
	return !time.Now().UTC().Before(expiry)
}

func (l *airgapLease) currentETag() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.etag
}

func (l *airgapLease) lastWriteStart() time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.writeStart
}

func (l *airgapLease) put(ctx context.Context, ifNoneMatch, ifMatch string) error {
	now := time.Now().UTC()
	record := airgapLeaseRecord{Owner: l.owner, AcquiredAt: now, RenewedAt: now, ExpiresAt: now.Add(l.ttl)}
	return l.putRecord(ctx, record, ifNoneMatch, ifMatch)
}

// putRecord bounds the write with renewTimeout and, on success, records the
// attempt START time as the freshness anchor, so a response delayed by S3
// never makes this process believe its lease is fresher than the server does.
// A commit whose response is lost fails closed: the old ETag makes the next
// conditional write fail.
func (l *airgapLease) putRecord(ctx context.Context, record airgapLeaseRecord, ifNoneMatch, ifMatch string) error {
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}
	input := &s3.PutObjectInput{Bucket: &l.bucket, Key: &l.key, Body: strings.NewReader(string(body))}
	if ifNoneMatch != "" {
		input.IfNoneMatch = &ifNoneMatch
	}
	if ifMatch != "" {
		input.IfMatch = &ifMatch
	}
	attemptStart := time.Now()
	putCtx, cancel := context.WithTimeout(ctx, l.renewTimeout)
	defer cancel()
	out, err := l.client.PutObject(putCtx, input)
	if err != nil {
		return err
	}
	l.mu.Lock()
	l.etag = aws.ToString(out.ETag)
	l.writeStart = attemptStart
	l.mu.Unlock()
	return nil
}

// renewLoop renews until ctx is done. onLost fires when another runner took
// the lease, or when renewals have failed long enough that the lease could
// expire before the next attempt succeeds; the caller must stop executing the
// deployment. Loss is declared interval+renewTimeout before nominal expiry so
// shutdown starts while the lease is still held with real margin (a blocked
// renewal can consume up to renewTimeout past the check). Freshness is
// measured from write attempt starts — seeded by the acquisition write — not
// from responses, so S3 latency never makes this process believe its lease is
// fresher than the server does.
func (l *airgapLease) renewLoop(ctx context.Context, logger *zap.Logger, onLost func()) {
	lastRenewed := l.lastWriteStart()
	if lastRenewed.IsZero() {
		lastRenewed = time.Now()
	}
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			attemptStart := time.Now()
			renewCtx, cancel := context.WithTimeout(ctx, l.renewTimeout)
			err := l.Renew(renewCtx)
			cancel()
			if err == nil {
				lastRenewed = attemptStart
				continue
			}
			if isS3ConditionFailed(err) || time.Since(lastRenewed) >= l.ttl-l.interval-l.renewTimeout {
				logger.Error("deployment lease lost; stopping", zap.Error(err))
				onLost()
				return
			}
			logger.Warn("deployment lease renewal failed; retrying", zap.Error(err))
		}
	}
}

func isS3NotFound(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "NoSuchKey" || code == "NotFound"
	}
	return false
}

func isS3ConditionFailed(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "PreconditionFailed" || code == "ConditionalRequestConflict"
	}
	return false
}
