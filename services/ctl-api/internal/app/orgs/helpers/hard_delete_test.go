package helpers

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
)

type retentionBlobStore struct{ deleted []string }

func (*retentionBlobStore) Upload(context.Context, string, []byte) error { return nil }
func (s *retentionBlobStore) Delete(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	return nil
}
func (*retentionBlobStore) Download(context.Context, string) ([]byte, error) { return nil, nil }
func (*retentionBlobStore) UploadStream(context.Context, string, io.Reader) (string, error) {
	return "", nil
}
func (*retentionBlobStore) DownloadStream(context.Context, string) (io.ReadCloser, error) {
	return nil, nil
}
func (*retentionBlobStore) GetMetadata(context.Context, string) (int64, string, error) {
	return 0, "", nil
}

type retentionCustomerManagedStore struct{ deleted []transport.Replica }

func (*retentionCustomerManagedStore) Configured() bool { return true }
func (*retentionCustomerManagedStore) Publish(context.Context, transport.PublishRequest) (transport.Replica, error) {
	return transport.Replica{}, nil
}
func (s *retentionCustomerManagedStore) Delete(_ context.Context, replica transport.Replica) error {
	s.deleted = append(s.deleted, replica)
	return nil
}
func (*retentionCustomerManagedStore) Grant(context.Context, transport.Replica, string, time.Time) (transport.DownloadGrant, error) {
	return transport.DownloadGrant{}, nil
}
func (*retentionCustomerManagedStore) PublishBlob(context.Context, string, string, []byte) error {
	return nil
}
func (*retentionCustomerManagedStore) GrantBlob(context.Context, string, string) (transport.BlobGrant, error) {
	return transport.BlobGrant{}, nil
}

func TestDeleteInstallSupportSnapshotObjects(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Exec(`CREATE TABLE install_support_snapshots (
		id text primary key, org_id text, archive_size integer, storage_provider text,
		storage_region text, storage_ref text, storage_version text, snapshot_blob text
	)`).Error)
	require.NoError(t, database.Exec(`INSERT INTO install_support_snapshots
		(id, org_id, archive_size, storage_provider, storage_region, storage_ref, storage_version, snapshot_blob)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "snapshot-a", "org-a", 42, transport.ProviderAWSS3, "us-east-1", "archives/a", "version-a", `{"blob_id":"blob-a","s3_key":"blobs/org-a/blob-a"}`).Error)
	blobs := &retentionBlobStore{}
	archives := &retentionCustomerManagedStore{}
	helpers := &Helpers{db: database, blobStore: blobs, customerManagedStore: archives}

	require.NoError(t, helpers.deleteCustomerManagedSupportSnapshotObjects(context.Background(), "org-a"))
	require.Equal(t, []string{"blobs/org-a/blob-a"}, blobs.deleted)
	require.Len(t, archives.deleted, 1)
	require.Equal(t, "archives/a", archives.deleted[0].StorageRef)
	require.Equal(t, "version-a", archives.deleted[0].StorageVersion)
}
