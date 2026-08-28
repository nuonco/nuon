package operationrun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationstate"
)

const mailboxTimeout = 15 * time.Second

const legacyETagPrefix = "legacy:"

var ErrAlreadyClaimed = errors.New("dispatch already claimed")

type S3Client interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type ListedRequest struct {
	DispatchID string
	Request    operation.Request
}

type Mailbox struct {
	client         S3Client
	bucket         string
	legacyPrefix   string
	runnerPrefix   string
	controlPrefix  string
	legacyFallback bool
	logger         *zap.Logger
}

func NewMailbox(client S3Client, bucket, prefix string, logger ...*zap.Logger) *Mailbox {
	l := zap.NewNop()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}
	prefix = strings.Trim(prefix, "/")
	return &Mailbox{client: client, bucket: bucket, legacyPrefix: prefix, runnerPrefix: prefix, controlPrefix: prefix, logger: l}
}

func NewNamespacedMailbox(client S3Client, bucket, prefix string, logger ...*zap.Logger) *Mailbox {
	m := NewMailbox(client, bucket, prefix, logger...)
	m.runnerPrefix = joinMailboxKey(m.legacyPrefix, operationstate.RunnerNamespace)
	m.controlPrefix = joinMailboxKey(m.legacyPrefix, operationstate.ControlNamespace)
	m.legacyFallback = true
	return m
}

func joinMailboxKey(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(name, "/")
}

func (m *Mailbox) ListRequests(ctx context.Context) ([]ListedRequest, error) {
	requests, err := m.listRequests(ctx, m.controlPrefix)
	if err != nil || !m.legacyFallback {
		return requests, err
	}
	legacy, err := m.listRequests(ctx, m.legacyPrefix)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(requests))
	for _, request := range requests {
		seen[request.DispatchID] = true
	}
	for _, request := range legacy {
		if !seen[request.DispatchID] {
			requests = append(requests, request)
		}
	}
	return requests, nil
}

func (m *Mailbox) listRequests(ctx context.Context, statePrefix string) ([]ListedRequest, error) {
	var requests []ListedRequest
	var token *string
	prefix := joinMailboxKey(statePrefix, operation.RequestsPrefix)
	for {
		callCtx, cancel := context.WithTimeout(ctx, mailboxTimeout)
		out, err := m.client.ListObjectsV2(callCtx, &s3.ListObjectsV2Input{Bucket: &m.bucket, Prefix: &prefix, ContinuationToken: token})
		cancel()
		if err != nil {
			return nil, fmt.Errorf("list dispatch requests: %w", err)
		}
		for _, object := range out.Contents {
			key := aws.ToString(object.Key)
			id := strings.TrimSuffix(strings.TrimPrefix(key, prefix), ".json")
			if id == "" || !strings.HasSuffix(key, ".json") {
				m.logger.Warn("skipping malformed dispatch request key", zap.String("key", key))
				continue
			}
			var req operation.Request
			found, err := m.get(ctx, key, &req)
			if err != nil || !found {
				m.logger.Warn("skipping malformed dispatch request", zap.String("key", key), zap.Error(err))
				continue
			}
			requests = append(requests, ListedRequest{DispatchID: id, Request: req})
		}
		if !aws.ToBool(out.IsTruncated) || out.NextContinuationToken == nil {
			return requests, nil
		}
		token = out.NextContinuationToken
	}
}

func (m *Mailbox) GetReceipt(ctx context.Context, id string) (*operation.Receipt, bool, error) {
	var receipt operation.Receipt
	found, err := m.getOwned(ctx, operation.ReceiptKey(id), &receipt)
	return &receipt, found, err
}

func (m *Mailbox) PutReceipt(ctx context.Context, receipt operation.Receipt) error {
	return m.put(ctx, m.runnerPrefix, operation.ReceiptKey(receipt.DispatchID), receipt, "", "")
}
func (m *Mailbox) ClaimNew(ctx context.Context, claim operation.Claim) error {
	if m.legacyFallback {
		var legacy operation.Claim
		_, found, err := m.getETag(ctx, joinMailboxKey(m.legacyPrefix, operation.ClaimKey(claim.DispatchID)), &legacy)
		if err != nil {
			return err
		}
		if found {
			return ErrAlreadyClaimed
		}
	}
	err := m.put(ctx, m.runnerPrefix, operation.ClaimKey(claim.DispatchID), claim, "*", "")
	if isConditionFailed(err) {
		return ErrAlreadyClaimed
	}
	return err
}
func (m *Mailbox) GetClaim(ctx context.Context, id string) (*operation.Claim, string, bool, error) {
	var claim operation.Claim
	etag, found, err := m.getETag(ctx, joinMailboxKey(m.runnerPrefix, operation.ClaimKey(id)), &claim)
	if err == nil && !found && m.legacyFallback {
		etag, found, err = m.getETag(ctx, joinMailboxKey(m.legacyPrefix, operation.ClaimKey(id)), &claim)
		if found {
			etag = legacyETagPrefix + etag
		}
	}
	return &claim, etag, found, err
}
func (m *Mailbox) TakeOverClaim(ctx context.Context, claim operation.Claim, etag string) error {
	prefix := m.runnerPrefix
	if strings.HasPrefix(etag, legacyETagPrefix) {
		prefix = m.legacyPrefix
		etag = strings.TrimPrefix(etag, legacyETagPrefix)
	}
	return m.put(ctx, prefix, operation.ClaimKey(claim.DispatchID), claim, "", etag)
}
func (m *Mailbox) PutRequest(ctx context.Context, req operation.Request) error {
	if m.legacyFallback {
		var legacy operation.Request
		found, err := m.get(ctx, joinMailboxKey(m.legacyPrefix, operation.RequestKey(req.DispatchID)), &legacy)
		if err != nil || found {
			return err
		}
	}
	err := m.put(ctx, m.controlPrefix, operation.RequestKey(req.DispatchID), req, "*", "")
	if isConditionFailed(err) {
		return nil
	}
	return err
}
func (m *Mailbox) PutCatalog(ctx context.Context, catalog operation.Catalog) error {
	return m.put(ctx, m.runnerPrefix, operation.CatalogKey, catalog, "", "")
}

func (m *Mailbox) PutBundleInfo(ctx context.Context, info operation.BundleInfo) error {
	return m.put(ctx, m.runnerPrefix, operation.BundleKey, info, "", "")
}

// PutBundleHistory records the digest's first activation; a pre-existing
// record wins so restarts and successor runners never rewrite history.
func (m *Mailbox) PutBundleHistory(ctx context.Context, info operation.BundleInfo) error {
	err := m.put(ctx, m.runnerPrefix, operation.BundleHistoryKey(info.BundleDigest), info, "*", "")
	if isConditionFailed(err) {
		return nil
	}
	return err
}

func (m *Mailbox) GetBundleHistory(ctx context.Context, digest string) (*operation.BundleInfo, bool, error) {
	var info operation.BundleInfo
	found, err := m.getOwned(ctx, operation.BundleHistoryKey(digest), &info)
	return &info, found, err
}

func (m *Mailbox) getOwned(ctx context.Context, key string, dst any) (bool, error) {
	found, err := m.get(ctx, joinMailboxKey(m.runnerPrefix, key), dst)
	if err == nil && !found && m.legacyFallback {
		return m.get(ctx, joinMailboxKey(m.legacyPrefix, key), dst)
	}
	return found, err
}

func (m *Mailbox) get(ctx context.Context, key string, dst any) (bool, error) {
	_, found, err := m.getETag(ctx, key, dst)
	return found, err
}
func (m *Mailbox) getETag(ctx context.Context, key string, dst any) (string, bool, error) {
	callCtx, cancel := context.WithTimeout(ctx, mailboxTimeout)
	defer cancel()
	out, err := m.client.GetObject(callCtx, &s3.GetObjectInput{Bucket: &m.bucket, Key: &key})
	if err != nil {
		if isNotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}
	defer out.Body.Close()
	if err := json.NewDecoder(out.Body).Decode(dst); err != nil {
		return "", false, err
	}
	return aws.ToString(out.ETag), true, nil
}
func (m *Mailbox) put(ctx context.Context, prefix, rel string, value any, ifNone, ifMatch string) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	key := joinMailboxKey(prefix, rel)
	in := &s3.PutObjectInput{Bucket: &m.bucket, Key: &key, Body: strings.NewReader(string(b))}
	if ifNone != "" {
		in.IfNoneMatch = &ifNone
	}
	if ifMatch != "" {
		in.IfMatch = &ifMatch
	}
	callCtx, cancel := context.WithTimeout(ctx, mailboxTimeout)
	defer cancel()
	_, err = m.client.PutObject(callCtx, in)
	return err
}
func isNotFound(err error) bool {
	var e smithy.APIError
	return errors.As(err, &e) && (e.ErrorCode() == "NoSuchKey" || e.ErrorCode() == "NotFound")
}
func isConditionFailed(err error) bool {
	var e smithy.APIError
	return errors.As(err, &e) && (e.ErrorCode() == "PreconditionFailed" || e.ErrorCode() == "ConditionalRequestConflict")
}
