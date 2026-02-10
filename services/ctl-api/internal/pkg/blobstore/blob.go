package blobstore

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// BlobMetadata represents the JSONB structure stored in the database
type BlobMetadata struct {
	BlobID      string `json:"blob_id"`                // S3 key (blob_id)
	S3Key       string `json:"s3_key"`                 // Full S3 path: org_id/blob_id
	Size        int64  `json:"size,omitempty"`         // Size in bytes
	ContentType string `json:"content_type,omitempty"` // MIME type
	Checksum    string `json:"checksum,omitempty"`     // SHA256 checksum
	CreatedBy   string `json:"created_by,omitempty"`   // Account ID who created the blob
	CreatedAt   string `json:"created_at,omitempty"`   // ISO 8601 timestamp
}

// Blob is a GORM custom type that stores large strings in S3
// The database column stores JSONB metadata including the S3 key
// The actual content is stored in S3 at: {org_id}/{owner_type}/{owner_id}/{blob_id}
type Blob struct {
	metadata BlobMetadata // Metadata stored in JSONB
	value    *string      // In-memory value (lazy loaded from S3)
	loaded   bool         // Whether value has been loaded from S3
	dirty    bool         // Whether value has been modified and needs upload
}

// Scan implements database/sql.Scanner
// Reads the blob metadata from database (JSONB column)
func (b *Blob) Scan(value interface{}) error {
	if value == nil {
		b.metadata = BlobMetadata{}
		b.value = nil
		b.loaded = false
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into Blob", value)
	}

	if err := json.Unmarshal(bytes, &b.metadata); err != nil {
		return fmt.Errorf("failed to unmarshal blob metadata: %w", err)
	}

	b.loaded = false // Not loaded from S3 yet
	return nil
}

// Value implements driver.Valuer
// Returns the blob metadata as JSONB to store in database
func (b *Blob) Value() (driver.Value, error) {
	if b.metadata.BlobID == "" {
		return nil, nil
	}

	bytes, err := json.Marshal(b.metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal blob metadata: %w", err)
	}

	return bytes, nil
}

// GormDataType returns the database column type
func (b Blob) GormDataType() string {
	return "jsonb"
}

// BeforeSave implements GORM hook for automatic S3 upload
func (b *Blob) BeforeSave(tx *gorm.DB) error {
	// Skip if not dirty (no changes)
	if !b.dirty {
		return nil
	}

	// Check context - is blob write enabled?
	if !IsBlobWriteEnabled(tx.Statement.Context) {
		return nil
	}

	// Get org ID from context using cctx
	orgID, err := cctxOrgIDFromContext(tx.Statement.Context)
	if err != nil {
		return fmt.Errorf("failed to get org_id from context: %w", err)
	}

	// Generate blob ID if new
	if b.metadata.BlobID == "" {
		b.metadata.BlobID = domains.NewBlobID()
	}

	// Construct S3 key: org_id/blob_id
	s3Key := fmt.Sprintf("%s/%s", orgID, b.metadata.BlobID)
	b.metadata.S3Key = s3Key

	// Get account ID for created_by
	if accountID, err := cctxAccountIDFromContext(tx.Statement.Context); err == nil {
		b.metadata.CreatedBy = accountID
	}

	// Set created timestamp
	b.metadata.CreatedAt = time.Now().Format(time.RFC3339)

	// If there's no value yet, write empty metadata but skip S3 upload
	// This supports the case where the blob entry is created first, then content is uploaded later
	if b.value == nil || *b.value == "" {
		b.metadata.Size = 0
		b.metadata.Checksum = ""
		if b.metadata.ContentType == "" {
			b.metadata.ContentType = "application/octet-stream"
		}
		b.dirty = false
		return nil
	}

	// Get blobstore service from context
	svc := GetBlobService(tx.Statement.Context)
	if svc == nil {
		return fmt.Errorf("blob service not set in context")
	}

	// Upload to S3 with streaming
	reader := strings.NewReader(*b.value)
	checksum, err := svc.UploadStream(tx.Statement.Context, s3Key, reader)
	if err != nil {
		return fmt.Errorf("failed to upload blob to S3: %w", err)
	}

	// Store metadata
	b.metadata.Checksum = checksum
	b.metadata.Size = int64(len(*b.value))
	if b.metadata.ContentType == "" {
		b.metadata.ContentType = "application/octet-stream" // Default
	}

	// Clear dirty flag
	b.dirty = false

	return nil
}

// AfterFind implements GORM hook for automatic S3 download
func (b *Blob) AfterFind(tx *gorm.DB) error {
	// Skip if no blob ID
	if b.metadata.BlobID == "" {
		return nil
	}

	// Skip if already loaded
	if b.loaded {
		return nil
	}

	// Detect if this is a single-row or multi-row query
	autoLoad := IsBlobAutoLoad(tx.Statement.Context)
	if !autoLoad {
		// Check if single row via reflection
		autoLoad = isSingleRowQuery(tx)
	}

	if !autoLoad {
		return nil // Skip auto-load for multi-row queries
	}

	// Get blobstore service
	svc := GetBlobService(tx.Statement.Context)
	if svc == nil {
		// No service available - skip silently
		return nil
	}

	// Use S3 key from metadata (already stored in JSONB)
	s3Key := b.metadata.S3Key

	// Download from S3 with streaming
	reader, err := svc.DownloadStream(tx.Statement.Context, s3Key)
	if err != nil {
		return fmt.Errorf("failed to download blob from S3: %w", err)
	}
	defer reader.Close()

	// Read into string
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read blob: %w", err)
	}

	// Set value
	valueStr := string(data)
	b.value = &valueStr
	b.loaded = true

	return nil
}

// isSingleRowQuery checks if the query returns a single struct vs slice
func isSingleRowQuery(tx *gorm.DB) bool {
	if tx.Statement.Dest == nil {
		return false
	}

	destType := reflect.TypeOf(tx.Statement.Dest)
	if destType.Kind() == reflect.Ptr {
		destType = destType.Elem()
	}

	// If dest is a slice, it's a multi-row query
	return destType.Kind() != reflect.Slice && destType.Kind() != reflect.Array
}

// Set sets the blob value and marks it dirty for upload
func (b *Blob) Set(value string) {
	b.value = &value
	b.dirty = true
	b.loaded = true
}

// Get returns the blob value, loading from S3 if needed
func (b *Blob) Get(ctx context.Context) (string, error) {
	if b.loaded && b.value != nil {
		return *b.value, nil
	}

	if b.metadata.BlobID == "" {
		return "", nil
	}

	svc := GetBlobService(ctx)
	if svc == nil {
		return "", fmt.Errorf("blob service not set in context")
	}

	// Use S3 key from metadata (already stored in JSONB)
	s3Key := b.metadata.S3Key

	reader, err := svc.DownloadStream(ctx, s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to download blob: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read blob: %w", err)
	}

	valueStr := string(data)
	b.value = &valueStr
	b.loaded = true

	return valueStr, nil
}

// IsSet returns whether the blob has a value (either loaded or with blob ID)
func (b *Blob) IsSet() bool {
	return b.metadata.BlobID != "" || (b.loaded && b.value != nil)
}

// String returns the blob value if loaded, or empty string
// Warning: Does not load from S3 if not already loaded
func (b *Blob) String() string {
	if b.loaded && b.value != nil {
		return *b.value
	}
	return ""
}

// BlobID returns the S3 key (blob ID)
func (b *Blob) BlobID() string {
	return b.metadata.BlobID
}

// Metadata returns the blob metadata
func (b *Blob) Metadata() BlobMetadata {
	return b.metadata
}

// SetContentType sets the content type for the blob
func (b *Blob) SetContentType(contentType string) {
	b.metadata.ContentType = contentType
}

// Helper functions to extract values from context using cctx

// cctxOrgIDFromContext extracts org ID from context using cctx package
func cctxOrgIDFromContext(ctx context.Context) (string, error) {
	return cctx.OrgIDFromContext(ctx)
}

// cctxAccountIDFromContext extracts account ID from context using cctx package
func cctxAccountIDFromContext(ctx context.Context) (string, error) {
	return cctx.AccountIDFromContext(ctx)
}
