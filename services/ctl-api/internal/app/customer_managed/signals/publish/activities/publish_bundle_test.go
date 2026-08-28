package activities

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
)

type countingStore struct {
	publishes     int
	blobPublishes int
}

func (*countingStore) Configured() bool { return true }
func (s *countingStore) Publish(context.Context, transport.PublishRequest) (transport.Replica, error) {
	s.publishes++
	return transport.Replica{}, nil
}
func (*countingStore) Delete(context.Context, transport.Replica) error { return nil }
func (*countingStore) Grant(context.Context, transport.Replica, string, time.Time) (transport.DownloadGrant, error) {
	return transport.DownloadGrant{}, nil
}
func (s *countingStore) PublishBlob(context.Context, string, string, []byte) error {
	s.blobPublishes++
	return nil
}
func (*countingStore) GrantBlob(context.Context, string, string) (transport.BlobGrant, error) {
	return transport.BlobGrant{}, nil
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

func TestCanonicalActionDefinitionExcludesPersistenceMetadata(t *testing.T) {
	envValue := "{{.nuon.install.id}}"
	semantic := app.ActionWorkflowConfig{
		ID: "config-one", CreatedAt: time.Now(), Timeout: time.Minute, Role: "provisioner",
		BreakGlassRoleARN: generics.NewNullString("arn:aws:iam::123:role/breakglass"),
		EnableKubeConfig:  sql.NullBool{Bool: true, Valid: true}, KubernetesContextName: "primary",
		ComponentDependencyIDs: []string{"component-id"}, References: []string{".nuon.components.api.outputs.url"},
		Triggers: []app.ActionWorkflowTriggerConfig{{Type: app.ActionWorkflowTriggerTypeCron, CronSchedule: "0 * * * *"}},
		Steps:    []app.ActionWorkflowStepConfig{{ID: "step-one", Name: "restart", Idx: 1, Command: "bash run.sh", InlineContents: "echo restart", EnvVars: pgtype.Hstore{"INSTALL_ID": &envValue}}},
	}
	copyWithNewMetadata := semantic
	copyWithNewMetadata.ID = "config-two"
	copyWithNewMetadata.CreatedAt = semantic.CreatedAt.Add(time.Hour)
	copyWithNewMetadata.Steps = append([]app.ActionWorkflowStepConfig(nil), semantic.Steps...)
	copyWithNewMetadata.Steps[0].ID = "step-two"
	connections := []app.ComponentConfigConnection{{ComponentID: "component-id", ComponentName: "api"}}

	first := canonicalActionDefinition(semantic, connections)
	second := canonicalActionDefinition(copyWithNewMetadata, connections)
	require.Equal(t, first, second)
	require.Equal(t, objectDigest(first), objectDigest(second))
	require.Equal(t, []string{"api"}, first.ComponentDependencies)
	require.NotEmpty(t, first.Steps[0].InlineContentsDigest)
	require.NotEqual(t, envValue, first.Steps[0].Environment["INSTALL_ID"])
}

func TestCanonicalComponentDefinitionExcludesPersistenceMetadata(t *testing.T) {
	port := "8081"
	semantic := app.ComponentConfigConnection{
		ID: "connection-one", CreatedAt: time.Now(), ComponentID: "component-one", ComponentName: "api",
		Type: app.ComponentTypeHelmChart, ComponentDependencyIDs: []string{"component-two"},
		BuildTimeout: "30m", KubernetesContextName: "primary",
		HelmComponentConfig: &app.HelmComponentConfig{
			ID: "helm-one", ComponentConfigConnectionID: "connection-one",
			HelmConfig: &app.HelmConfig{ChartName: "whoami", Values: map[string]*string{"service.port": &port}},
		},
	}
	copyWithNewMetadata := semantic
	copyWithNewMetadata.ID = "connection-three"
	copyWithNewMetadata.CreatedAt = semantic.CreatedAt.Add(time.Hour)
	copyWithNewMetadata.ComponentID = "component-three"
	copyWithNewMetadata.HelmComponentConfig = &app.HelmComponentConfig{
		ID: "helm-three", ComponentConfigConnectionID: "connection-three",
		HelmConfig: semantic.HelmComponentConfig.HelmConfig,
	}
	connections := []app.ComponentConfigConnection{{ComponentID: "component-two", ComponentName: "database"}}

	first, err := canonicalComponentDefinition(semantic, connections)
	require.NoError(t, err)
	second, err := canonicalComponentDefinition(copyWithNewMetadata, connections)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Equal(t, objectDigest(first), objectDigest(second))
	require.Equal(t, []string{"database"}, first["dependencies"])
	require.NotContains(t, first, "id")
	require.NotContains(t, first["helm"], "id")
}

func TestCanonicalComponentDefinitionChangesWithSemanticConfig(t *testing.T) {
	port8080 := "8080"
	port8081 := "8081"
	connection := app.ComponentConfigConnection{
		Type: app.ComponentTypeHelmChart,
		HelmComponentConfig: &app.HelmComponentConfig{HelmConfig: &app.HelmConfig{
			ChartName: "whoami", Values: map[string]*string{"service.port": &port8080},
		}},
	}
	before, err := canonicalComponentDefinition(connection, nil)
	require.NoError(t, err)
	connection.HelmComponentConfig.HelmConfig.Values["service.port"] = &port8081
	after, err := canonicalComponentDefinition(connection, nil)
	require.NoError(t, err)
	require.NotEqual(t, objectDigest(before), objectDigest(after))
}

func TestCanonicalRunbookDefinitionsResolveSemanticReferences(t *testing.T) {
	envelope := &customermanaged.Envelope{
		Actions: []customermanaged.ActionTemplate{{ID: "action-id", Name: "restart"}},
		Drift:   []customermanaged.DriftTemplate{{ID: "drift-id", ComponentName: "database"}},
		Runbooks: []customermanaged.RunbookTemplate{{Name: "recover", Steps: []customermanaged.RunbookStep{
			{Kind: customermanaged.RunbookStepKindAction, RefID: "action-id"},
			{Kind: customermanaged.RunbookStepKindDrift, RefID: "drift-id"},
			{Kind: customermanaged.RunbookStepKindHealthGate, Component: "api"},
		}}},
	}

	definitions := canonicalEnvelopeRunbookDefinitions(envelope)
	require.Equal(t, []bundle.Runbook{{
		Name: "recover", ConfigDigest: objectDigest(bundle.RunbookDefinition{Steps: []bundle.RunbookStepDefinition{
			{Kind: "action", Reference: "action:restart"},
			{Kind: "drift", Reference: "drift:database"},
			{Kind: "health-gate", Component: "api"},
		}}), Definition: bundle.RunbookDefinition{Steps: []bundle.RunbookStepDefinition{
			{Kind: "action", Reference: "action:restart"},
			{Kind: "drift", Reference: "drift:database"},
			{Kind: "health-gate", Component: "api"},
		}},
	}}, definitions)
}

func TestExternalImageEntries(t *testing.T) {
	artifact := bundle.Artifact{MediaType: "application/vnd.oci.image.manifest.v1+json", Digest: "sha256:digest", Size: 42}
	connection := app.ComponentConfigConnection{
		ID:                           "connection-id",
		ComponentID:                  "component-id",
		Type:                         app.ComponentTypeExternalImage,
		ExternalImageComponentConfig: &app.ExternalImageComponentConfig{ImageURL: "traefik/whoami"},
	}

	image, record := externalImageEntries(connection, "whoami", artifact, "bundle-repository", "sha256:config")
	require.Equal(t, &bundle.Image{Name: "whoami", Repository: "traefik/whoami", Artifact: artifact}, image)
	require.Equal(t, &app.ReleasePackageMember{
		Kind:                        "image",
		LogicalName:                 "whoami",
		ComponentConfigConnectionID: connection.ID,
		ComponentID:                 "component-id",
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

func TestValidateRunnerImageBundleSchemaLabels(t *testing.T) {
	labels := map[string]string{
		"io.nuon.customer_managed.bundle-schema.min": "2",
		"io.nuon.customer_managed.bundle-schema.max": "3",
	}
	require.NoError(t, validateRunnerImageBundleSchemaLabels(labels, 3))
	require.ErrorContains(t, validateRunnerImageBundleSchemaLabels(labels, 4), "outside runner-supported range")
	require.ErrorContains(t, validateRunnerImageBundleSchemaLabels(nil, 3), "missing or invalid")
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

func TestPortalInputsPackagesStandaloneBinary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bundle-portal")
	require.NoError(t, os.WriteFile(path, []byte("portal-binary"), 0o755))
	activities := &Activities{}

	root, record, asset, err := activities.portalInputs(context.Background(), "file://"+path)
	require.NoError(t, err)
	require.Equal(t, "portal_binary", record.Kind)
	require.Equal(t, "portal", record.LogicalName)
	require.Equal(t, "portal_binary", asset.Role)
	require.Equal(t, root.Descriptor.Digest.String(), asset.Digest)
	require.Equal(t, root.Descriptor.Digest.String(), record.Digest)
}

func TestPublishBundleReturnsWhenVerifiedPublicationAlreadyExists(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, database.Exec(`CREATE TABLE app_releases (
		id text primary key, org_id text, app_id text, app_config_id text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE release_packages (
		id text primary key, updated_at datetime, org_id text, release_id text, manifest_digest text,
		oci_root_digest text, oci_index_digest text, archive_checksum text, archive_size integer, target_platform text,
		schema_version integer, status text, status_description text
	)`).Error)
	require.NoError(t, database.Exec(`CREATE TABLE release_package_replicas (
		id text primary key, org_id text, package_id text, verified_at datetime
	)`).Error)

	digest := "sha256:" + strings.Repeat("a", 64)
	releaseID := strings.Repeat("1", 26)
	packageID := strings.Repeat("2", 26)
	require.NoError(t, database.Exec(`INSERT INTO app_releases (id, org_id, app_id, app_config_id) VALUES (?, 'org-a', 'app-a', 'config-a')`, releaseID).Error)
	require.NoError(t, database.Exec(`INSERT INTO release_packages
		(id, org_id, release_id, manifest_digest, oci_root_digest, oci_index_digest, archive_checksum,
		archive_size, target_platform, schema_version, status, status_description)
		VALUES (?, 'org-a', ?, ?, ?, ?, ?, 42, 'linux/amd64', 1, 'publishing', 'publishing')`, packageID, releaseID, digest, digest, digest, digest).Error)
	require.NoError(t, database.Exec(`INSERT INTO release_package_replicas
		(id, org_id, package_id, verified_at) VALUES (?, 'org-a', ?, ?)`, strings.Repeat("3", 26), packageID, time.Now()).Error)

	store := &countingStore{}
	activities := &Activities{db: database, store: store}
	require.NoError(t, activities.PublishBundle(context.Background(), &PublishBundleRequest{PackageID: packageID}))
	require.Zero(t, store.publishes)
	var published app.ReleasePackage
	require.NoError(t, database.Where(app.ReleasePackage{ID: packageID}).First(&published).Error)
	require.Equal(t, app.ReleasePackageStatusActive, published.Status)
}
