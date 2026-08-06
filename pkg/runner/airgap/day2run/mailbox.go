package day2run

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

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

const mailboxTimeout = 15 * time.Second

var ErrAlreadyClaimed = errors.New("dispatch already claimed")

type S3Client interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type ListedRequest struct {
	DispatchID string
	Request    day2.Request
}

type Mailbox struct {
	client S3Client
	bucket string
	prefix string
	logger *zap.Logger
}

func NewMailbox(client S3Client, bucket, prefix string, logger ...*zap.Logger) *Mailbox {
	l := zap.NewNop()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}
	return &Mailbox{client: client, bucket: bucket, prefix: strings.Trim(prefix, "/"), logger: l}
}

func (m *Mailbox) key(name string) string {
	if m.prefix == "" {
		return name
	}
	return m.prefix + "/" + strings.TrimPrefix(name, "/")
}

func (m *Mailbox) ListRequests(ctx context.Context) ([]ListedRequest, error) {
	var requests []ListedRequest
	var token *string
	prefix := m.key(day2.RequestsPrefix)
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
			var req day2.Request
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

func (m *Mailbox) GetReceipt(ctx context.Context, id string) (*day2.Receipt, bool, error) {
	var receipt day2.Receipt
	found, err := m.get(ctx, m.key(day2.ReceiptKey(id)), &receipt)
	return &receipt, found, err
}

func (m *Mailbox) PutReceipt(ctx context.Context, receipt day2.Receipt) error {
	return m.put(ctx, day2.ReceiptKey(receipt.DispatchID), receipt, "", "")
}
func (m *Mailbox) ClaimNew(ctx context.Context, claim day2.Claim) error {
	err := m.put(ctx, day2.ClaimKey(claim.DispatchID), claim, "*", "")
	if isConditionFailed(err) {
		return ErrAlreadyClaimed
	}
	return err
}
func (m *Mailbox) GetClaim(ctx context.Context, id string) (*day2.Claim, string, bool, error) {
	var claim day2.Claim
	etag, found, err := m.getETag(ctx, m.key(day2.ClaimKey(id)), &claim)
	return &claim, etag, found, err
}
func (m *Mailbox) TakeOverClaim(ctx context.Context, claim day2.Claim, etag string) error {
	return m.put(ctx, day2.ClaimKey(claim.DispatchID), claim, "", etag)
}
func (m *Mailbox) PutRequest(ctx context.Context, req day2.Request) error {
	err := m.put(ctx, day2.RequestKey(req.DispatchID), req, "*", "")
	if isConditionFailed(err) {
		return nil
	}
	return err
}
func (m *Mailbox) PutCatalog(ctx context.Context, catalog day2.Catalog) error {
	return m.put(ctx, day2.CatalogKey, catalog, "", "")
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
func (m *Mailbox) put(ctx context.Context, rel string, value any, ifNone, ifMatch string) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	key := m.key(rel)
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
