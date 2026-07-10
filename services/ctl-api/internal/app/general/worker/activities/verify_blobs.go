package activities

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"golang.org/x/time/rate"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

const maxVerifyMismatchSamples = 100

type VerifyBlobsRequest struct {
	Table     string `json:"table"`
	BatchSize int    `json:"batch_size"`
	// Day scopes the batch to a single UTC calendar day ("2006-01-02"); empty means the whole table.
	Day string `json:"day"`
	// Cursor is the last id from the previous batch; rows are walked in ascending id order.
	Cursor string `json:"cursor"`
}

type BlobMismatch struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

type VerifyBlobsResponse struct {
	Checked    int64          `json:"checked"`
	Mismatched int64          `json:"mismatched"`
	NotSet     int64          `json:"not_set"`
	Mismatches []BlobMismatch `json:"mismatches,omitempty"`
	NextCursor string         `json:"next_cursor"`
}

type blobVerifyRow struct {
	ID       string `gorm:"column:id"`
	BlobMeta string `gorm:"column:blob_meta"`
	Content  []byte `gorm:"column:content"`
}

// VerifyBlobs reads a batch of rows as-is and tallies each: null blob -> not_set;
// otherwise download from S3 and confirm checksum + content against the origin
// column (byte-for-byte, or JSON-semantic for jsonb) -> mismatched on a diff.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 15m
func (a *Activities) VerifyBlobs(ctx context.Context, req VerifyBlobsRequest) (*VerifyBlobsResponse, error) {
	target, ok := blobBackfillTargets[req.Table]
	if !ok {
		return nil, fmt.Errorf("unsupported blob verify table: %q", req.Table)
	}
	if req.BatchSize <= 0 {
		req.BatchSize = 1000
	}

	ratePerSecond := a.cfg.BlobBackfillRatePerSecond
	if ratePerSecond <= 0 {
		ratePerSecond = defaultBlobBackfillRatePerSecond
	}
	limiter := rate.NewLimiter(rate.Limit(ratePerSecond), ratePerSecond)

	q := a.db.WithContext(ctx).
		Table(req.Table).
		Select(fmt.Sprintf("id, %s::text AS blob_meta, %s AS content", target.blobColumn, target.contentExpr())).
		Where(fmt.Sprintf("%s IS NOT NULL", target.originColumn))
	q = target.applyNotDeleted(q)
	q, err := whereDay(q, req.Day)
	if err != nil {
		return nil, err
	}
	if req.Cursor != "" {
		q = q.Where("id > ?", req.Cursor)
	}

	var rows []blobVerifyRow
	if res := q.Order("id").Limit(req.BatchSize).Scan(&rows); res.Error != nil {
		return nil, fmt.Errorf("unable to select rows to verify from %s: %w", req.Table, res.Error)
	}

	resp := &VerifyBlobsResponse{}

	for _, row := range rows {
		resp.Checked++
		resp.NextCursor = row.ID

		if row.BlobMeta == "" {
			resp.NotSet++
			continue
		}

		if err := limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("blob verify rate limiter interrupted: %w", err)
		}

		reason, err := a.verifyBlobRow(ctx, target.jsonSemantic, row)
		if err != nil {
			return nil, err
		}
		if reason != "" {
			resp.Mismatched++
			if len(resp.Mismatches) < maxVerifyMismatchSamples {
				resp.Mismatches = append(resp.Mismatches, BlobMismatch{ID: row.ID, Reason: reason})
			}
		}
	}

	a.mw.Count("general.blob_verify.checked", resp.Checked, []string{"table:" + req.Table})
	a.mw.Count("general.blob_verify.mismatched", resp.Mismatched, []string{"table:" + req.Table})
	a.mw.Count("general.blob_verify.not_set", resp.NotSet, []string{"table:" + req.Table})
	return resp, nil
}

// verifyBlobRow returns an empty reason when the row's blob matches both its
// recorded checksum and its origin column; otherwise it returns a description of
// the first mismatch found.
func (a *Activities) verifyBlobRow(ctx context.Context, jsonSemantic bool, row blobVerifyRow) (string, error) {
	var metadata blobstore.BlobMetadata
	if err := json.Unmarshal([]byte(row.BlobMeta), &metadata); err != nil {
		return fmt.Sprintf("unable to parse blob metadata: %v", err), nil
	}
	if metadata.S3Key == "" {
		return "blob metadata has no s3_key", nil
	}

	reader, err := a.blobSvc.DownloadStream(ctx, metadata.S3Key)
	if err != nil {
		return fmt.Sprintf("unable to download blob from s3 key %q: %v", metadata.S3Key, err), nil
	}
	defer reader.Close()

	hash := sha256.New()
	content, err := io.ReadAll(io.TeeReader(reader, hash))
	if err != nil {
		return "", fmt.Errorf("unable to read blob from s3 key %q: %w", metadata.S3Key, err)
	}

	if got := fmt.Sprintf("sha256:%x", hash.Sum(nil)); metadata.Checksum != "" && got != metadata.Checksum {
		return fmt.Sprintf("s3 object checksum %s does not match recorded %s", got, metadata.Checksum), nil
	}

	if jsonSemantic {
		equal, err := jsonEqual(string(content), string(row.Content))
		if err != nil {
			return fmt.Sprintf("unable to compare blob json to origin column: %v", err), nil
		}
		if !equal {
			return "blob json content does not match origin column", nil
		}
		return "", nil
	}

	if !bytes.Equal(content, row.Content) {
		return "blob content does not match origin column", nil
	}
	return "", nil
}

// ListVerifyDays returns the UTC days with rows (origin set), regardless of blob state.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) ListVerifyDays(ctx context.Context, req ListBlobDaysRequest) (*ListBlobDaysResponse, error) {
	target, ok := blobBackfillTargets[req.Table]
	if !ok {
		return nil, fmt.Errorf("unsupported blob verify table: %q", req.Table)
	}

	days, err := a.listBlobDays(ctx, req.Table, target, "")
	if err != nil {
		return nil, err
	}
	return &ListBlobDaysResponse{Days: days}, nil
}

func jsonEqual(a, b string) (bool, error) {
	var av, bv any
	if err := json.Unmarshal([]byte(a), &av); err != nil {
		return false, fmt.Errorf("blob: %w", err)
	}
	if err := json.Unmarshal([]byte(b), &bv); err != nil {
		return false, fmt.Errorf("origin: %w", err)
	}
	return reflect.DeepEqual(av, bv), nil
}
