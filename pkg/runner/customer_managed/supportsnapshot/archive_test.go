package supportsnapshot

import (
	"archive/tar"
	"bytes"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
)

func testSnapshot(t *testing.T) Snapshot {
	t.Helper()
	registration, err := customermanaged.NewInstallationRegistration(customermanaged.InstallationRegistration{
		SchemaVersion: customermanaged.InstallationRegistrationSchemaVersion,
		ReleaseID:     "release-1", ReleaseDigest: "sha256:" + string(bytes.Repeat([]byte("c"), 64)),
		PackageID: "package-1", PackageDigest: "sha256:" + string(bytes.Repeat([]byte("d"), 64)),
		BundleDigest:  "sha256:" + string(bytes.Repeat([]byte("a"), 64)),
		ArchiveDigest: "sha256:" + string(bytes.Repeat([]byte("b"), 64)),
		OperationID:   "install-run-1",
		DeploymentID:  "prod", InstallID: "vinst-prod",
		Cloud:       customermanaged.InstallationRegistrationCloud{Provider: "aws", AccountID: "123456789012", Region: "us-east-1"},
		Stack:       customermanaged.InstallationRegistrationStack{Type: "aws-cloudformation", ID: "stack-id", Name: "install"},
		InstalledAt: time.Now().UTC().Add(-time.Hour),
	})
	require.NoError(t, err)
	return Snapshot{
		SchemaVersion: SchemaVersion, CapturedAt: time.Now().UTC(), Registration: registration,
		Collection: CollectionReport{SchemaVersion: SchemaVersion, Redaction: "support-v1", Included: []string{"runs"}},
	}
}

func TestArchiveRoundTrip(t *testing.T) {
	snapshot := testSnapshot(t)
	var archive bytes.Buffer
	manifest, err := Write(&archive, snapshot, Producer{Name: "bundle-portal", Version: "test"})
	require.NoError(t, err)
	require.Len(t, manifest.Entries, 3)

	decoded, err := Read(bytes.NewReader(archive.Bytes()))
	require.NoError(t, err)
	require.Equal(t, snapshot.Registration.RegistrationID, decoded.Snapshot.Registration.RegistrationID)
	require.Equal(t, "support-v1", decoded.Collection.Redaction)
}

func TestArchiveRejectsUnsafeEntry(t *testing.T) {
	var archive bytes.Buffer
	encoder, err := zstd.NewWriter(&archive)
	require.NoError(t, err)
	tw := tar.NewWriter(encoder)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "../secret", Typeflag: tar.TypeReg, Size: 1}))
	_, err = tw.Write([]byte("x"))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, encoder.Close())

	_, err = Read(bytes.NewReader(archive.Bytes()))
	require.ErrorContains(t, err, "not a safe regular file")
}
