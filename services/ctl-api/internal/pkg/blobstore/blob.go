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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
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
	s3Prefix string       // Optional custom S3 prefix (bypasses org_id requirement)
}

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

	b.loaded = false
	return nil
}

func (b *Blob) Value() (driver.Value, error) {
	if b == nil || b.metadata.BlobID == "" {
		return nil, nil
	}

	bytes, err := json.Marshal(b.metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal blob metadata: %w", err)
	}

	return bytes, nil
}

func (b Blob) GormDataType() string {
	return "jsonb"
}

func (b *Blob) BeforeCreate(tx *gorm.DB) error {
	if b == nil {
		return nil
	}
	if !b.dirty {
		return nil
	}

	if !IsBlobWriteEnabled(tx.Statement.Context) {
		return nil
	}

	if b.metadata.BlobID == "" {
		b.metadata.BlobID = domains.NewBlobID()
	}

	var s3Key string
	if b.s3Prefix != "" {
		s3Key = fmt.Sprintf("%s/%s", b.s3Prefix, b.metadata.BlobID)
	} else {
		orgID, err := cctxOrgIDFromContext(tx.Statement.Context)
		if err != nil {
			return fmt.Errorf("failed to get org_id from context: %w", err)
		}
		s3Key = buildS3Key(orgID, b.metadata.BlobID)
	}
	b.metadata.S3Key = s3Key

	if accountID, err := cctxAccountIDFromContext(tx.Statement.Context); err == nil {
		b.metadata.CreatedBy = accountID
	}

	b.metadata.CreatedAt = time.Now().Format(time.RFC3339)

	if b.value == nil || *b.value == "" {
		b.metadata.Size = 0
		b.metadata.Checksum = ""
		if b.metadata.ContentType == "" {
			b.metadata.ContentType = "application/octet-stream"
		}
		b.dirty = false
		return nil
	}

	svc := GetBlobService(tx.Statement.Context)
	if svc == nil {
		return fmt.Errorf("blob service not set in context")
	}

	reader := strings.NewReader(*b.value)
	checksum, err := svc.UploadStream(tx.Statement.Context, s3Key, reader)
	if err != nil {
		return fmt.Errorf("failed to upload blob to S3 key (%s): %w", s3Key, err)
	}

	b.metadata.Checksum = checksum
	b.metadata.Size = int64(len(*b.value))
	if b.metadata.ContentType == "" {
		b.metadata.ContentType = "application/octet-stream"
	}

	b.dirty = false

	return nil
}

func (b *Blob) AfterFind(tx *gorm.DB) error {
	if b.metadata.BlobID == "" {
		return nil
	}

	if b.loaded {
		return nil
	}

	autoLoad := IsBlobAutoLoad(tx.Statement.Context)
	if !autoLoad {
		autoLoad = isSingleRowQuery(tx)
	}

	if !autoLoad {
		return nil
	}

	svc := GetBlobService(tx.Statement.Context)
	if svc == nil {
		return nil
	}

	s3Key := b.metadata.S3Key

	reader, err := svc.DownloadStream(tx.Statement.Context, s3Key)
	if err != nil {
		return fmt.Errorf("failed to download blob from S3: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read blob: %w", err)
	}

	valueStr := string(data)
	b.value = &valueStr
	b.loaded = true

	return nil
}

func isSingleRowQuery(tx *gorm.DB) bool {
	if tx.Statement.Dest == nil {
		return false
	}

	destType := reflect.TypeOf(tx.Statement.Dest)
	if destType.Kind() == reflect.Ptr {
		destType = destType.Elem()
	}

	return destType.Kind() != reflect.Slice && destType.Kind() != reflect.Array
}

func (b *Blob) Set(value string) {
	b.value = &value
	b.dirty = true
	b.loaded = true
}

func (b *Blob) Get(ctx context.Context) (string, error) {
	if b == nil {
		return "", nil
	}
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

func (b *Blob) IsSet() bool {
	return b.metadata.BlobID != "" || (b.loaded && b.value != nil)
}

// why: String returns the blob value if loaded, or empty string
// Warning: Does not load from S3 if not already loaded
func (b *Blob) String() string {
	if b.loaded && b.value != nil {
		return *b.value
	}
	return ""
}

func (b *Blob) BlobID() string {
	return b.metadata.BlobID
}

func (b *Blob) Metadata() BlobMetadata {
	return b.metadata
}

func (b *Blob) SetContentType(contentType string) {
	b.metadata.ContentType = contentType
}

func (b *Blob) SetS3Prefix(prefix string) {
	b.s3Prefix = prefix
}

func cctxOrgIDFromContext(ctx context.Context) (string, error) {
	orgID := keys.OrgIDFromContext(ctx)
	if orgID == "" {
		return "", fmt.Errorf("org ID not set on context")
	}
	return orgID, nil
}

func cctxAccountIDFromContext(ctx context.Context) (string, error) {
	accountID := keys.CreatedByIDFromContext(ctx)
	if accountID == "" {
		return "", fmt.Errorf("account ID not set on context")
	}
	return accountID, nil
}

func buildS3Key(orgID, blobID string) string {
	return fmt.Sprintf("blobs/%s/%s", orgID, blobID)
}
