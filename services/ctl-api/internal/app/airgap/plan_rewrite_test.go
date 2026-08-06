package airgap

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
)

func planRewriteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	for _, ddl := range []string{
		`CREATE TABLE component_builds (id text primary key, org_id text, component_config_connection_id text, deleted_at integer not null default 0)`,
		`CREATE TABLE component_config_connections (id text primary key, org_id text, component_id text, deleted_at integer not null default 0)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}
	return db
}

func testPins() BundlePins {
	return BundlePins{
		SandboxBuildID:    "bldsandboxnew0000000000000",
		SandboxRegistry:   &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypeECR, Repository: "123.dkr.ecr.us-west-2.amazonaws.com/org/app", Region: "us-west-2"},
		ComponentBuildIDs: map[string]string{"cmp00000000000000000000001": "bldnew00000000000000000001"},
	}
}

func rewriteTestEnvelope(t *testing.T, steps ...runnerairgap.Step) *runnerairgap.Envelope {
	t.Helper()
	return &runnerairgap.Envelope{Version: "v0", OrgID: "org1", AppID: "app1", InstallID: "inl1", Steps: steps}
}

func TestRewriteEnvelopeForBundleSandboxSource(t *testing.T) {
	db := planRewriteTestDB(t)
	plan := `{"sandbox_run_plan":{"install_id":"inl1","git_source":{"url":"https://github.com/nuonco/aws-eks-auto-sandbox","ref":"main","path":"."},"oci_source":null,"vars":{"keep":"me"}}}`
	envelope := rewriteTestEnvelope(t, runnerairgap.Step{ID: "job1", JobType: "sandbox-terraform", JobGroup: "sandbox", CompositePlan: json.RawMessage(plan)})

	require.NoError(t, RewriteEnvelopeForBundle(context.Background(), db, envelope, testPins()))

	var decoded map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(envelope.Steps[0].CompositePlan, &decoded))
	sandboxPlan := decoded["sandbox_run_plan"]
	require.Equal(t, "null", string(sandboxPlan["git_source"]))
	require.JSONEq(t, `{"keep":"me"}`, string(sandboxPlan["vars"]))

	var ociSource struct {
		Registry *configs.OCIRegistryRepository `json:"registry"`
		Tag      string                         `json:"tag"`
	}
	require.NoError(t, json.Unmarshal(sandboxPlan["oci_source"], &ociSource))
	require.Equal(t, "bldsandboxnew0000000000000", ociSource.Tag)
	require.Equal(t, "123.dkr.ecr.us-west-2.amazonaws.com/org/app", ociSource.Registry.Repository)
}

func TestRewriteEnvelopeForBundleRetagsSyncSources(t *testing.T) {
	db := planRewriteTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO component_config_connections (id, org_id, component_id) VALUES ('con1', 'org1', 'cmp00000000000000000000001')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO component_builds (id, org_id, component_config_connection_id) VALUES ('bldold0000000000000000001', 'org1', 'con1')`).Error)

	plan := `{"sync_oci_plan":{"src_registry":{"repository":"org/app"},"src_tag":"bldold0000000000000000001","dst_registry":{"repository":"customer/install"},"dst_tag":"dplhistoric000000000000001"}}`
	envelope := rewriteTestEnvelope(t, runnerairgap.Step{ID: "job1", JobType: "oci-sync", JobGroup: "sync", CompositePlan: json.RawMessage(plan)})

	require.NoError(t, RewriteEnvelopeForBundle(context.Background(), db, envelope, testPins()))

	var decoded map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(envelope.Steps[0].CompositePlan, &decoded))
	syncPlan := decoded["sync_oci_plan"]
	require.Equal(t, `"bldnew00000000000000000001"`, string(syncPlan["src_tag"]))
	require.Equal(t, `"dplhistoric000000000000001"`, string(syncPlan["dst_tag"]))
}

func TestRewriteEnvelopeForBundleKeepsAlreadyPinnedSyncSource(t *testing.T) {
	db := planRewriteTestDB(t)
	plan := `{"sync_oci_plan":{"src_registry":{"repository":"org/app"},"src_tag":"bldnew00000000000000000001","dst_registry":{"repository":"customer/install"},"dst_tag":"dpl1"}}`
	envelope := rewriteTestEnvelope(t, runnerairgap.Step{ID: "job1", JobType: "oci-sync", JobGroup: "sync", CompositePlan: json.RawMessage(plan)})

	require.NoError(t, RewriteEnvelopeForBundle(context.Background(), db, envelope, testPins()))

	var decoded map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(envelope.Steps[0].CompositePlan, &decoded))
	require.Equal(t, `"bldnew00000000000000000001"`, string(decoded["sync_oci_plan"]["src_tag"]))
}

func TestRewriteEnvelopeForBundleFailsClosedOnUnknownSyncSource(t *testing.T) {
	db := planRewriteTestDB(t)
	plan := `{"sync_oci_plan":{"src_registry":{"repository":"org/app"},"src_tag":"bldghost000000000000000001","dst_registry":{"repository":"customer/install"},"dst_tag":"dpl1"}}`
	envelope := rewriteTestEnvelope(t, runnerairgap.Step{ID: "job1", JobType: "oci-sync", JobGroup: "sync", CompositePlan: json.RawMessage(plan)})

	err := RewriteEnvelopeForBundle(context.Background(), db, envelope, testPins())
	require.ErrorContains(t, err, "bldghost000000000000000001")
}

func TestRewriteEnvelopeForBundleFailsClosedOnUnpinnedComponent(t *testing.T) {
	db := planRewriteTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO component_config_connections (id, org_id, component_id) VALUES ('con2', 'org1', 'cmpunpinned000000000000001')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO component_builds (id, org_id, component_config_connection_id) VALUES ('bldold0000000000000000002', 'org1', 'con2')`).Error)

	plan := `{"sync_oci_plan":{"src_registry":{"repository":"org/app"},"src_tag":"bldold0000000000000000002","dst_registry":{"repository":"customer/install"},"dst_tag":"dpl1"}}`
	envelope := rewriteTestEnvelope(t, runnerairgap.Step{ID: "job1", JobType: "oci-sync", JobGroup: "sync", CompositePlan: json.RawMessage(plan)})

	err := RewriteEnvelopeForBundle(context.Background(), db, envelope, testPins())
	require.ErrorContains(t, err, "no pinned build")
}

func TestRewriteEnvelopeForBundleRequiresSandboxPin(t *testing.T) {
	db := planRewriteTestDB(t)
	plan := `{"sandbox_run_plan":{"install_id":"inl1"}}`
	envelope := rewriteTestEnvelope(t, runnerairgap.Step{ID: "job1", JobType: "sandbox-terraform", JobGroup: "sandbox", CompositePlan: json.RawMessage(plan)})

	pins := testPins()
	pins.SandboxBuildID = ""
	err := RewriteEnvelopeForBundle(context.Background(), db, envelope, pins)
	require.ErrorContains(t, err, "no pinned sandbox build")
}
