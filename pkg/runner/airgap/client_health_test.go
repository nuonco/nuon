package airgap

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func healthTestEnvelope() *Envelope {
	envelope := testEnvelope()
	envelope.Components = []ComponentSpec{
		{InstallComponentID: "inc-1", ComponentID: "cmp-1", ComponentName: "api", ComponentType: "helm_chart", HelmReleaseName: "api", HelmNamespace: "default"},
	}
	return envelope
}

func healthReport(health string) *models.ServiceCreateComponentHealthRequest {
	return &models.ServiceCreateComponentHealthRequest{
		Kind: "watch",
		Components: []*models.ServiceComponentHealthComponent{
			{ComponentID: "cmp-1", Resources: []*models.ServiceComponentHealthResource{{Kind: "Deployment", Name: "api", Health: health}}},
		},
	}
}

func readTransitions(t *testing.T, store *statestore.Disk) []HealthTransition {
	t.Helper()
	raw, ok, err := store.ReadHealthTransitions()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		return nil
	}
	var transitions []HealthTransition
	if err := json.Unmarshal(raw, &transitions); err != nil {
		t.Fatal(err)
	}
	return transitions
}

func TestClientComponentHealthPersistsSnapshotsAndTransitions(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(healthTestEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if _, err := client.CreateComponentHealth(ctx, healthReport("healthy")); err != nil {
		t.Fatal(err)
	}
	raw, ok, err := store.ReadHealth()
	if err != nil || !ok {
		t.Fatalf("latest health not persisted: %v %v", ok, err)
	}
	var snapshot HealthSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Components) != 1 || snapshot.Components[0].ComponentName != "api" || snapshot.Components[0].Health != "healthy" {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}

	transitions := readTransitions(t, store)
	if len(transitions) != 1 || transitions[0].From != "" || transitions[0].To != "healthy" {
		t.Fatalf("first report should record one transition: %+v", transitions)
	}

	if _, err := client.CreateComponentHealth(ctx, healthReport("healthy")); err != nil {
		t.Fatal(err)
	}
	if transitions = readTransitions(t, store); len(transitions) != 1 {
		t.Fatalf("unchanged health should not append transitions: %+v", transitions)
	}

	if _, err := client.CreateComponentHealth(ctx, healthReport("degraded")); err != nil {
		t.Fatal(err)
	}
	transitions = readTransitions(t, store)
	if len(transitions) != 2 || transitions[1].From != "healthy" || transitions[1].To != "degraded" {
		t.Fatalf("degradation should append a transition: %+v", transitions)
	}
}

func TestClientComponentHealthSeedsFromDiskAfterRestart(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(healthTestEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateComponentHealth(context.Background(), healthReport("healthy")); err != nil {
		t.Fatal(err)
	}

	restarted, err := NewClient(healthTestEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := restarted.CreateComponentHealth(context.Background(), healthReport("healthy")); err != nil {
		t.Fatal(err)
	}
	if transitions := readTransitions(t, store); len(transitions) != 1 {
		t.Fatalf("restart with unchanged health should not record transitions: %+v", transitions)
	}
}

func TestClientGetRunnerInstallComponents(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(healthTestEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.GetRunnerInstallComponents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.InstallID != "install" || len(resp.Components) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	c := resp.Components[0]
	if c.InstallComponentID != "inc-1" || c.ComponentID != "cmp-1" || c.ComponentName != "api" || c.HelmReleaseName != "api" || c.HelmNamespace != "default" {
		t.Fatalf("unexpected component: %+v", c)
	}
}

func TestClientComponentHealthContextRoundTrip(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(healthTestEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	info, releases, kinds, err := client.GetComponentHealthContext(ctx)
	if err != nil || info != "" || releases != nil || kinds != nil {
		t.Fatalf("missing context should be empty: %q %v %v %v", info, releases, kinds, err)
	}

	if err := client.PutComponentHealthContext(ctx, `{"cluster":"eks"}`, []string{"ingress"}, []string{"Deployment"}); err != nil {
		t.Fatal(err)
	}

	restarted, err := NewClient(healthTestEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	info, releases, kinds, err = restarted.GetComponentHealthContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info != `{"cluster":"eks"}` || len(releases) != 1 || releases[0] != "ingress" || len(kinds) != 1 || kinds[0] != "Deployment" {
		t.Fatalf("context did not survive restart: %q %v %v", info, releases, kinds)
	}
}
