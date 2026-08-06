package activities

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/transport"
)

type countingStore struct{ publishes int }

func (*countingStore) Configured() bool { return true }
func (s *countingStore) Publish(context.Context, transport.PublishRequest) (transport.Replica, error) {
	s.publishes++
	return transport.Replica{}, nil
}
func (*countingStore) Grant(context.Context, transport.Replica, string, time.Time) (transport.DownloadGrant, error) {
	return transport.DownloadGrant{}, nil
}

func TestValidateStackAssetURL(t *testing.T) {
	for _, source := range []string{
		"https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#default",
		"https://nuon-artifacts.s3.us-west-2.amazonaws.com/runner/install.sh",
		"https://templates.example/stacks/a.yaml",
	} {
		u, err := validateStackAssetURL(source, "https://templates.example/")
		require.NoError(t, err, source)
		require.NotNil(t, u)
	}

	for _, source := range []string{
		"https://example.com/init.sh",
		"http://raw.githubusercontent.com/nuonco/runner/init.sh",
		"https://user@raw.githubusercontent.com/nuonco/runner/init.sh",
	} {
		_, err := validateStackAssetURL(source, "https://templates.example/")
		require.Error(t, err, source)
	}
}

func TestExternalImageEntries(t *testing.T) {
	artifact := bundle.Artifact{MediaType: "application/vnd.oci.image.manifest.v1+json", Digest: "sha256:digest", Size: 42}
	connection := app.ComponentConfigConnection{
		ID:                           "connection-id",
		Type:                         app.ComponentTypeExternalImage,
		ExternalImageComponentConfig: &app.ExternalImageComponentConfig{ImageURL: "traefik/whoami"},
	}

	image, record := externalImageEntries(connection, "whoami", artifact, "bundle-repository", "sha256:config")
	require.Equal(t, &bundle.Image{Name: "whoami", Repository: "traefik/whoami", Artifact: artifact}, image)
	require.Equal(t, &app.AirgapBundleArtifact{
		Kind:                        "image",
		LogicalName:                 "whoami",
		ComponentConfigConnectionID: connection.ID,
		ConfigDigest:                "sha256:config",
		Repository:                  "bundle-repository",
		Digest:                      artifact.Digest,
		MediaType:                   artifact.MediaType,
		Size:                        artifact.Size,
	}, record)

	connection.Type = app.ComponentTypeHelmChart
	image, record = externalImageEntries(connection, "whoami", artifact, "bundle-repository", "sha256:config")
	require.Nil(t, image)
	require.Nil(t, record)
}

func TestCustomStackSource(t *testing.T) {
	hash := strings.Repeat("a", 64)
	source, err := customStackSource("https://templates.example/", "org-a", "app-a", hash, "./stack.yaml")
	require.NoError(t, err)
	require.Equal(t,
		"https://templates.example/stacks/org-a/app-a/"+hash+".yaml",
		source,
	)
	_, err = customStackSource("https://templates.example/", "org-a", "app-a", "", "https://templates.example/existing.yaml")
	require.ErrorContains(t, err, "64-character SHA-256")
	_, err = customStackSource("https://templates.example/", "org-a", "app-a", strings.Repeat("z", 64), "./stack.yaml")
	require.ErrorContains(t, err, "64-character SHA-256")
}

func TestPublishBundleReturnsWhenVerifiedPublicationAlreadyExists(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, database.Exec(`CREATE TABLE airgap_bundles (
		id text primary key, org_id text, app_id text, app_config_id text, manifest_digest text,
		oci_root_digest text, transport_checksum text, size integer, target_platform text,
		schema_version integer, status text, status_description text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE airgap_bundle_transport_replicas (
		id text primary key, org_id text, bundle_id text, verified_at datetime
	)`).Error)

	digest := "sha256:" + strings.Repeat("a", 64)
	bundleID := strings.Repeat("1", 26)
	require.NoError(t, database.Exec(`INSERT INTO airgap_bundles
		(id, org_id, app_id, app_config_id, manifest_digest, oci_root_digest, transport_checksum,
		size, target_platform, schema_version, status, status_description)
		VALUES (?, 'org-a', 'app-a', 'config-a', ?, ?, ?, 42, 'linux/amd64', 1, 'publishing', 'publishing')`, bundleID, digest, digest, digest).Error)
	require.NoError(t, database.Exec(`INSERT INTO airgap_bundle_transport_replicas
		(id, org_id, bundle_id, verified_at) VALUES (?, 'org-a', ?, ?)`, strings.Repeat("2", 26), bundleID, time.Now()).Error)

	store := &countingStore{}
	activities := &Activities{db: database, store: store}
	require.NoError(t, activities.PublishBundle(context.Background(), &PublishBundleRequest{BundleID: bundleID}))
	require.Zero(t, store.publishes)
	var published app.AirgapBundle
	require.NoError(t, database.Where(app.AirgapBundle{ID: bundleID}).First(&published).Error)
	require.Equal(t, app.AirgapBundleStatusActive, published.Status)
}
