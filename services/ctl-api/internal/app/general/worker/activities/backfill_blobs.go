package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

const (
	defaultBlobBackfillRatePerSecond = 500
	blobBackfillContentType          = "application/octet-stream"
)

// blobBackfillTargets whitelists the tables/columns the backfill is allowed to
// touch. Table and column names are interpolated into SQL, so they must never
// come from request input — only from this map.
var blobBackfillTargets = map[string]blobBackfillTarget{
	"install_states":                  {originColumn: "state", blobColumn: "state_blob", jsonSemantic: true},
	"install_workflow_step_approvals": {originColumn: "contents", blobColumn: "contents_blob"},
	"runner_job_plans":                {originColumn: "composite_plan", blobColumn: "composite_plan_blob", jsonSemantic: true},
	"runner_job_execution_outputs":    {originColumn: "outputs", blobColumn: "outputs_blob", jsonSemantic: true},
	"terraform_workspace_states":      {originColumn: "contents", blobColumn: "contents_blob", binary: true},
	"terraform_workspace_state_jsons": {originColumn: "contents", blobColumn: "contents_blob", binary: true, noSoftDelete: true},
}

type blobBackfillTarget struct {
	originColumn string
	blobColumn   string
	// jsonSemantic compares the blob against its origin column as parsed JSON
	// rather than byte-for-byte. Required for jsonb origin columns, whose
	// Postgres serialization differs from the Go-marshaled bytes uploaded to S3.
	jsonSemantic bool
	// binary marks a bytea origin column: it is read raw rather than cast to
	// text (a ::text cast would yield Postgres hex escapes, not the bytes), and
	// compared byte-for-byte.
	binary bool
	// noSoftDelete marks a table without a deleted_at column, so the
	// "deleted_at = 0" predicate must be skipped (it would be a SQL error).
	noSoftDelete bool
}

// contentExpr selects the origin column's bytes: raw for bytea, ::text otherwise.
func (t blobBackfillTarget) contentExpr() string {
	if t.binary {
		return t.originColumn
	}
	return t.originColumn + "::text"
}

// applyNotDeleted adds the soft-delete predicate unless the target's table has
// no deleted_at column.
func (t blobBackfillTarget) applyNotDeleted(q *gorm.DB) *gorm.DB {
	if t.noSoftDelete {
		return q
	}
	return q.Where("deleted_at = 0")
}

type BackfillBlobsRequest struct {
	Table     string `json:"table"`
	BatchSize int    `json:"batch_size"`
	// Day scopes the batch to a single UTC calendar day ("2006-01-02"); empty means the whole table.
	Day string `json:"day"`
}

type BackfillBlobsResponse struct {
	Processed int64 `json:"processed"`
}

type ListBlobDaysRequest struct {
	Table string `json:"table"`
}

type ListBlobDaysResponse struct {
	Days []string `json:"days"`
}

const (
	// dayLayout is the Go layout; pgDayFormat is the equivalent Postgres to_char
	// template — they are NOT interchangeable (different pattern languages).
	dayLayout   = "2006-01-02"
	pgDayFormat = "YYYY-MM-DD"
)

type blobBackfillRow struct {
	ID        string `gorm:"column:id"`
	OrgID     string `gorm:"column:org_id"`
	CreatedBy string `gorm:"column:created_by_id"`
	Content   []byte `gorm:"column:content"`
}

// BackfillBlobs mirrors a batch of rows with content but a null blob column into
// S3 and records the blob metadata. Self-draining (backfilled rows drop out of
// the predicate) and paced by an in-activity rate limiter.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 15m
func (a *Activities) BackfillBlobs(ctx context.Context, req BackfillBlobsRequest) (*BackfillBlobsResponse, error) {
	target, ok := blobBackfillTargets[req.Table]
	if !ok {
		return nil, fmt.Errorf("unsupported blob backfill table: %q", req.Table)
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
		Select(fmt.Sprintf("id, org_id, created_by_id, %s AS content", target.contentExpr())).
		Where(fmt.Sprintf("%s IS NULL", target.blobColumn)).
		Where(fmt.Sprintf("%s IS NOT NULL", target.originColumn))
	q = target.applyNotDeleted(q)
	q, err := whereDay(q, req.Day)
	if err != nil {
		return nil, err
	}

	var rows []blobBackfillRow
	res := q.Order("created_at").
		Limit(req.BatchSize).
		Scan(&rows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to select rows to backfill from %s: %w", req.Table, res.Error)
	}

	var processed int64
	for _, row := range rows {
		if err := limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("blob backfill rate limiter interrupted: %w", err)
		}

		if err := a.backfillBlobRow(ctx, req.Table, target.blobColumn, row); err != nil {
			return nil, err
		}
		processed++
	}

	a.mw.Count("general.blob_backfill.processed", processed, []string{"table:" + req.Table})
	return &BackfillBlobsResponse{Processed: processed}, nil
}

func (a *Activities) backfillBlobRow(ctx context.Context, table, blobColumn string, row blobBackfillRow) error {
	if row.OrgID == "" {
		return fmt.Errorf("row %s in %s has no org_id; cannot build blob key", row.ID, table)
	}

	blobID := domains.NewBlobID()
	s3Key := fmt.Sprintf("blobs/%s/%s", row.OrgID, blobID)

	checksum, err := a.blobSvc.UploadStream(ctx, s3Key, bytes.NewReader(row.Content))
	if err != nil {
		return fmt.Errorf("unable to upload blob for %s row %s: %w", table, row.ID, err)
	}

	metadata := blobstore.BlobMetadata{
		BlobID:      blobID,
		S3Key:       s3Key,
		Size:        int64(len(row.Content)),
		ContentType: blobBackfillContentType,
		Checksum:    checksum,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("unable to marshal blob metadata for %s row %s: %w", table, row.ID, err)
	}

	res := a.db.WithContext(ctx).
		Table(table).
		Where("id = ?", row.ID).
		Update(blobColumn, string(metadataJSON))
	if res.Error != nil {
		return fmt.Errorf("unable to record blob metadata on %s row %s: %w", table, row.ID, res.Error)
	}

	return nil
}

// ListBackfillDays returns the UTC days with un-mirrored rows (origin set, blob null).
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) ListBackfillDays(ctx context.Context, req ListBlobDaysRequest) (*ListBlobDaysResponse, error) {
	target, ok := blobBackfillTargets[req.Table]
	if !ok {
		return nil, fmt.Errorf("unsupported blob backfill table: %q", req.Table)
	}

	days, err := a.listBlobDays(ctx, req.Table, target, fmt.Sprintf("%s IS NULL", target.blobColumn))
	if err != nil {
		return nil, err
	}
	return &ListBlobDaysResponse{Days: days}, nil
}

// listBlobDays returns the distinct UTC days with a non-null origin column,
// optionally narrowed by blobPredicate. table/predicate/column come only from the
// whitelist map — never request input — since they are interpolated into SQL.
func (a *Activities) listBlobDays(ctx context.Context, table string, target blobBackfillTarget, blobPredicate string) ([]string, error) {
	q := a.db.WithContext(ctx).
		Table(table).
		Distinct(fmt.Sprintf("to_char(created_at AT TIME ZONE 'UTC', '%s') AS day", pgDayFormat)).
		Where(fmt.Sprintf("%s IS NOT NULL", target.originColumn))
	q = target.applyNotDeleted(q)
	if blobPredicate != "" {
		q = q.Where(blobPredicate)
	}

	var days []string
	if res := q.Order("day").Pluck("day", &days); res.Error != nil {
		return nil, fmt.Errorf("unable to list blob days for %s: %w", table, res.Error)
	}
	return days, nil
}

// whereDay scopes a query to rows created within a single UTC calendar day. An
// empty day is a no-op.
func whereDay(q *gorm.DB, day string) (*gorm.DB, error) {
	if day == "" {
		return q, nil
	}
	start, err := time.ParseInLocation(dayLayout, day, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("invalid day %q: %w", day, err)
	}
	return q.Where("created_at >= ? AND created_at < ?", start, start.Add(24*time.Hour)), nil
}
