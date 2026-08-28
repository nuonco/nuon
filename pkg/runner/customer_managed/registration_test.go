package customermanaged

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInstallationRegistrationCanonicalIdentity(t *testing.T) {
	input := InstallationRegistration{
		SchemaVersion: InstallationRegistrationSchemaVersion,
		ReleaseID:     "release-1", ReleaseDigest: "sha256:" + strings.Repeat("c", 64),
		PackageID: "package-1", PackageDigest: "sha256:" + strings.Repeat("d", 64),
		BundleDigest:  "SHA256:" + strings.Repeat("A", 64),
		ArchiveDigest: "sha256:" + strings.Repeat("b", 64),
		OperationID:   "install-run-1",
		DeploymentID:  "prod",
		InstallID:     "vinst1234-prod",
		Cloud:         InstallationRegistrationCloud{Provider: "AWS", AccountID: "123456789012", Region: "us-east-1"},
		Stack:         InstallationRegistrationStack{Type: "AWS-CloudFormation", ID: "arn:aws:cloudformation:us-east-1:123456789012:stack/install/id", Name: "install"},
		InstalledAt:   time.Date(2026, time.August, 24, 12, 0, 0, 0, time.FixedZone("offset", 2*60*60)),
	}

	first, err := NewInstallationRegistration(input)
	require.NoError(t, err)
	second, err := NewInstallationRegistration(first)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Equal(t, "aws", first.Cloud.Provider)
	require.Equal(t, time.UTC, first.InstalledAt.Location())
	require.NoError(t, first.Validate())
}

func TestInstallationRegistrationRejectsModifiedContents(t *testing.T) {
	registration, err := NewInstallationRegistration(InstallationRegistration{
		SchemaVersion: InstallationRegistrationSchemaVersion,
		ReleaseID:     "release-1", ReleaseDigest: "sha256:" + strings.Repeat("c", 64),
		PackageID: "package-1", PackageDigest: "sha256:" + strings.Repeat("d", 64),
		BundleDigest:  "sha256:" + strings.Repeat("a", 64),
		ArchiveDigest: "sha256:" + strings.Repeat("b", 64),
		OperationID:   "install-run-1",
		DeploymentID:  "prod",
		InstallID:     "vinst1234-prod",
		Cloud:         InstallationRegistrationCloud{Provider: "aws", AccountID: "123456789012", Region: "us-east-1"},
		Stack:         InstallationRegistrationStack{Type: "aws-cloudformation", ID: "stack-id", Name: "install"},
		InstalledAt:   time.Now(),
	})
	require.NoError(t, err)

	registration.Cloud.AccountID = "999999999999"
	require.ErrorContains(t, registration.Validate(), "does not match")
}

func TestInstallationRegistrationV1RemainsValidWithoutOperationID(t *testing.T) {
	registration, err := NewInstallationRegistration(InstallationRegistration{
		SchemaVersion: 1,
		ReleaseID:     "release-1", ReleaseDigest: "sha256:" + strings.Repeat("c", 64),
		PackageID: "package-1", PackageDigest: "sha256:" + strings.Repeat("d", 64),
		BundleDigest: "sha256:" + strings.Repeat("a", 64), ArchiveDigest: "sha256:" + strings.Repeat("b", 64),
		DeploymentID: "prod", InstallID: "vinst1234-prod",
		Cloud:       InstallationRegistrationCloud{Provider: "aws", AccountID: "123456789012", Region: "us-east-1"},
		Stack:       InstallationRegistrationStack{Type: "aws-cloudformation", ID: "stack-id", Name: "install"},
		InstalledAt: time.Now(),
	})
	require.NoError(t, err)
	require.Empty(t, registration.OperationID)
	require.NoError(t, registration.Validate())
}

func TestInstallationRegistrationValidatesRequiredIdentity(t *testing.T) {
	_, err := NewInstallationRegistration(InstallationRegistration{
		SchemaVersion: InstallationRegistrationSchemaVersion,
		ReleaseID:     "release-1", ReleaseDigest: "sha256:" + strings.Repeat("c", 64),
		PackageID: "package-1", PackageDigest: "sha256:" + strings.Repeat("d", 64),
	})
	require.ErrorContains(t, err, "bundle digest")
}
