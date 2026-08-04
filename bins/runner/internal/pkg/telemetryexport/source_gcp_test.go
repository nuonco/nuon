package telemetryexport

import (
	"context"
	"errors"
	"testing"
	"time"

	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gcpFactoryStub struct {
	client    gcpSecretClient
	projectID string
	err       error
}

func (f *gcpFactoryStub) New(context.Context) (gcpSecretClient, string, error) {
	return f.client, f.projectID, f.err
}

type gcpSecretClientStub struct {
	name   string
	result *secretmanagerpb.AccessSecretVersionResponse
	err    error
	closed chan struct{}
}

func (c *gcpSecretClientStub) AccessSecretVersion(_ context.Context, request *secretmanagerpb.AccessSecretVersionRequest, _ ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	c.name = request.Name
	return c.result, c.err
}

func (c *gcpSecretClientStub) Close() error {
	close(c.closed)
	return nil
}

func TestNewGCPConfigUpdateClassifiesResults(t *testing.T) {
	tests := []struct {
		name   string
		result *secretmanagerpb.AccessSecretVersionResponse
		err    error
		state  configUpdateState
		value  string
	}{
		{name: "available", result: &secretmanagerpb.AccessSecretVersionResponse{Payload: &secretmanagerpb.SecretPayload{Data: []byte("config")}}, state: configAvailable, value: "config"},
		{name: "empty", result: &secretmanagerpb.AccessSecretVersionResponse{}, state: configAvailable},
		{name: "not found", err: status.Error(codes.NotFound, "secret version not found"), state: configNotFound},
		{name: "disabled", err: status.Error(codes.FailedPrecondition, "secret version is disabled"), state: configUnavailable},
		{name: "permission denied", err: status.Error(codes.PermissionDenied, "permission denied"), state: configUnavailable},
		{name: "unauthenticated", err: status.Error(codes.Unauthenticated, "unauthenticated"), state: configUnavailable},
		{name: "lookup failure", err: errors.New("lookup failed"), state: configLookupFailed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			update := newGCPConfigUpdate(test.result, test.err)
			if update.state != test.state || update.value != test.value {
				t.Fatalf("expected state=%d value=%q, got state=%d value=%q", test.state, test.value, update.state, update.value)
			}
		})
	}
}

func TestGCPConfigSourceReadsLatestSecretAndClosesClient(t *testing.T) {
	client := &gcpSecretClientStub{
		result: &secretmanagerpb.AccessSecretVersionResponse{Payload: &secretmanagerpb.SecretPayload{Data: []byte("config")}},
		closed: make(chan struct{}),
	}
	factory := &gcpFactoryStub{client: client, projectID: "test-project"}
	source := newGCPConfigSource(factory, "inst-test")
	ctx, cancel := context.WithCancel(context.Background())

	update := <-source.Watch(ctx, time.Hour)
	if update.state != configAvailable || update.value != "config" {
		t.Fatalf("unexpected configuration update: %#v", update)
	}
	if client.name != "projects/test-project/secrets/inst-test-telemetry-export-config/versions/latest" {
		t.Fatalf("unexpected secret version name: %q", client.name)
	}

	cancel()
	select {
	case <-client.closed:
	case <-time.After(time.Second):
		t.Fatal("GCP Secret Manager client was not closed")
	}
}

func TestConfigObservationDeduplicatesEquivalentGCPErrors(t *testing.T) {
	var observation configObservation
	first := status.Error(codes.PermissionDenied, "permission denied")
	second := status.Error(codes.PermissionDenied, "permission denied")

	if !observation.changed(configUpdate{state: configLookupFailed, err: first, errorIdentity: gcpErrorIdentity(first)}) {
		t.Fatal("first GCP error was not emitted")
	}
	if observation.changed(configUpdate{state: configLookupFailed, err: second, errorIdentity: gcpErrorIdentity(second)}) {
		t.Fatal("equivalent GCP error was emitted again")
	}
}
